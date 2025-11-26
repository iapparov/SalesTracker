package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/wb-go/wbf/retry"
	wbzlog "github.com/wb-go/wbf/zlog"
	"salestracker/internal/domain/analytic"
	"time"
)

func (p *Postgres) GetAnalytics(from, to time.Time, groupBy, splitBy, sortBy, sortDir string) (*analytic.Analytics, error) {
	ctx := context.Background()

	var dateTrunc string
	switch groupBy {
	case "month":
		dateTrunc = "month"
	case "year":
		dateTrunc = "year"
	default:
		dateTrunc = "day"
	}

	var splitColumn string
	if splitBy == "category" {
		splitColumn = "category"
	} else {
		splitColumn = "transtype"
	}

	sortColumn := "group_key"
	switch sortBy {
	case "sum", "avg", "count", "median", "percentile90":
		sortColumn = sortBy
	}
	sortDirection := "DESC"
	if sortDir == "asc" {
		sortDirection = "ASC"
	}

	query := fmt.Sprintf(`
	WITH grouped AS (
	SELECT
		to_char(date_trunc('%s', transdate), 'YYYY-MM-DD') AS group_key,
		%s AS split_key,
		SUM(amount) AS sum,
		AVG(amount) AS avg,
		COUNT(*) AS count,
		percentile_cont(0.5) WITHIN GROUP (ORDER BY amount) AS median,
		percentile_cont(0.9) WITHIN GROUP (ORDER BY amount) AS percentile90,
		SUM(CASE WHEN transtype='income' THEN amount ELSE 0 END) 
		- SUM(CASE WHEN transtype='expense' THEN amount ELSE 0 END) AS sum_signed
	FROM transactions
	WHERE transdate >= $1 AND transdate <= $2
	GROUP BY group_key, split_key
	),
	all_grouped AS (
	SELECT
		group_key,
		SUM(sum_signed) AS sum,  -- это и будет All = income - expense
		SUM(count) AS count,
		AVG(sum / NULLIF(count,0)) AS avg,
		AVG(median) AS median,
		AVG(percentile90) AS percentile90
	FROM grouped
	GROUP BY group_key
	)
	SELECT 
		g.group_key, g.split_key, g.sum, g.avg, g.count, g.median, g.percentile90,
		a.sum AS all_sum, a.count AS all_count, a.avg AS all_avg, a.median AS all_median, a.percentile90 AS all_perc90
	FROM grouped g
	JOIN all_grouped a USING(group_key)
	ORDER BY %s %s;
	`, dateTrunc, splitColumn, sortColumn, sortDirection)

	rows, err := p.db.QueryWithRetry(ctx, retry.Strategy{Attempts: p.cfg.Attempts, Delay: p.cfg.Delay, Backoff: p.cfg.Backoffs}, query, from, to)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("Error executing analytics query")
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	result := &analytic.Analytics{}
	groupMap := map[string]*analytic.AnalyticByType{}

	for rows.Next() {
		var groupKey, splitKey string
		var sum, avg, median, perc90 float64
		var count int
		var allSum, allAvg, allMedian, allPerc90 float64
		var allCount int

		if err := rows.Scan(&groupKey, &splitKey, &sum, &avg, &count, &median, &perc90,
			&allSum, &allCount, &allAvg, &allMedian, &allPerc90); err != nil {
			wbzlog.Logger.Error().Err(err).Msg("Error scanning analytics row")
			return nil, err
		}

		if _, ok := groupMap[groupKey]; !ok {
			groupMap[groupKey] = &analytic.AnalyticByType{
				AllMap: map[string]analytic.Analytic{},
			}
		}

		a := analytic.Analytic{
			Sum: sum, Avg: avg, Count: count, Median: median, Percentile90: perc90,
		}

		groupMap[groupKey].AllMap[splitKey] = a
		groupMap[groupKey].All = analytic.Analytic{
			Sum:          allSum,
			Count:        allCount,
			Avg:          allAvg,
			Median:       allMedian,
			Percentile90: allPerc90,
		}

		// Если splitBy=transtype, присвоим Income/Expense
		if splitBy == "type" || splitBy == "transtype" {
			switch splitKey {
			case "income":
				groupMap[groupKey].Income = a
			case "expense":
				groupMap[groupKey].Expense = a
			}
		}
	}

	for k, v := range groupMap {
		result.Groups = append(result.Groups, analytic.AnalyticGroup{
			GroupKey: k,
			Data:     *v,
		})
	}

	summaryQuery := `
	SELECT
		COALESCE(SUM(CASE WHEN transtype='income' THEN amount END),0) AS income_sum,
		COUNT(CASE WHEN transtype='income' THEN 1 END) AS income_count,
		COALESCE(AVG(CASE WHEN transtype='income' THEN amount END),0) AS income_avg,
		COALESCE(percentile_cont(0.5) WITHIN GROUP (ORDER BY CASE WHEN transtype='income' THEN amount END),0) AS income_median,
		COALESCE(percentile_cont(0.9) WITHIN GROUP (ORDER BY CASE WHEN transtype='income' THEN amount END),0) AS income_perc90,
		COALESCE(SUM(CASE WHEN transtype='expense' THEN amount END),0) AS expense_sum,
		COUNT(CASE WHEN transtype='expense' THEN 1 END) AS expense_count,
		COALESCE(AVG(CASE WHEN transtype='expense' THEN amount END),0) AS expense_avg,
		COALESCE(percentile_cont(0.5) WITHIN GROUP (ORDER BY CASE WHEN transtype='expense' THEN amount END),0) AS expense_median,
		COALESCE(percentile_cont(0.9) WITHIN GROUP (ORDER BY CASE WHEN transtype='expense' THEN amount END),0) AS expense_perc90
	FROM transactions
	WHERE transdate >= $1 AND transdate <= $2;
	`

	row, err := p.db.QueryRowWithRetry(ctx, retry.Strategy{Attempts: p.cfg.Attempts, Delay: p.cfg.Delay, Backoff: p.cfg.Backoffs}, summaryQuery, from, to)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("Error executing analytics summary query")
		return nil, err
	}

	var incomeSum, incomeAvg, incomeMedian, incomePerc90 float64
	var incomeCount int
	var expenseSum, expenseAvg, expenseMedian, expensePerc90 float64
	var expenseCount int

	if err := row.Scan(&incomeSum, &incomeCount, &incomeAvg, &incomeMedian, &incomePerc90,
		&expenseSum, &expenseCount, &expenseAvg, &expenseMedian, &expensePerc90); err != nil {
		if err == sql.ErrNoRows {
			wbzlog.Logger.Info().Msg("No rows returned for analytics summary")
			return nil, nil
		}
		wbzlog.Logger.Error().Err(err).Msg("Error scanning analytics summary row")
		return nil, err
	}

	result.Summary = analytic.AnalyticByType{
		Income: analytic.Analytic{
			Sum: incomeSum, Count: incomeCount, Avg: incomeAvg, Median: incomeMedian, Percentile90: incomePerc90,
		},
		Expense: analytic.Analytic{
			Sum: expenseSum, Count: expenseCount, Avg: expenseAvg, Median: expenseMedian, Percentile90: expensePerc90,
		},
		All: analytic.Analytic{
			Sum:   incomeSum - expenseSum,
			Count: incomeCount + expenseCount,
			Avg: func() float64 {
				if incomeCount+expenseCount == 0 {
					return 0
				}
				return (incomeSum - expenseSum) / float64(incomeCount+expenseCount)
			}(),
			Median: func() float64 {
				if incomeCount == 0 && expenseCount == 0 {
					return 0
				}
				return (incomeMedian - expenseMedian) / 2
			}(),
			Percentile90: (incomePerc90 + expensePerc90) / 2,
		},
	}

	return result, nil
}
