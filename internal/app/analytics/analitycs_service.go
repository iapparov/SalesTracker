package analytics

import (
	"encoding/csv"
	"fmt"
	wbzlog "github.com/wb-go/wbf/zlog"
	"io"
	"salestracker/internal/domain/analytic"
	"time"
)

type AnalyticService struct {
	repo AnalyticStorageProvider
}

type AnalyticStorageProvider interface {
	GetAnalytics(from, to time.Time, groupBy, splitBy, sortBy, sortDir string) (*analytic.Analytics, error)
}

func NewAnalyticService(repo AnalyticStorageProvider) *AnalyticService {
	return &AnalyticService{
		repo: repo,
	}
}

func (s *AnalyticService) GetAnalytics(from, to time.Time, groupBy, splitBy, sortBy, sortDir string) (*analytic.Analytics, error) {
	if from.After(to) {
		err := fmt.Errorf("'from' date cannot be after 'to'")
		wbzlog.Logger.Warn().Err(err).Msg("invalid date range in analytics request")
		return nil, err
	}

	if groupBy == "" {
		groupBy = "day"
	}
	if splitBy == "" {
		splitBy = "transtype"
	}

	result, err := s.repo.GetAnalytics(from, to, groupBy, splitBy, sortBy, sortDir)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("analytics repository error")
		return nil, err
	}

	return result, nil
}

func (s *AnalyticService) GetCSV(from, to time.Time, groupBy, splitBy, sortBy, sortDir string, output io.Writer) error {
	anals, err := s.repo.GetAnalytics(from, to, groupBy, splitBy, sortBy, sortDir)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("repo get analytics error")
		return err
	}

	writer := csv.NewWriter(output)
	defer writer.Flush()

	headers := []string{"GroupKey", "Type", "Sum", "Avg", "Count", "Median", "Percentile90"}
	if err := writer.Write(headers); err != nil {
		wbzlog.Logger.Error().Err(err).Msg("Error writing CSV headers")
		return err
	}

	for _, group := range anals.Groups {
		typesMap := map[string]analytic.Analytic{
			"Income":  group.Data.Income,
			"Expense": group.Data.Expense,
			"All":     group.Data.All,
		}

		for typ, data := range typesMap {
			row := []string{
				group.GroupKey,
				typ,
				fmt.Sprintf("%.2f", data.Sum),
				fmt.Sprintf("%.2f", data.Avg),
				fmt.Sprintf("%d", data.Count),
				fmt.Sprintf("%.2f", data.Median),
				fmt.Sprintf("%.2f", data.Percentile90),
			}
			if err := writer.Write(row); err != nil {
				wbzlog.Logger.Error().Err(err).Msg("Error writing CSV row")
				return err
			}
		}
	}

	wbzlog.Logger.Info().Msg("CSV report generation completed")
	return nil
}
