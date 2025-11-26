package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/wb-go/wbf/retry"
	wbzlog "github.com/wb-go/wbf/zlog"
	"salestracker/internal/domain/transaction"
	"time"
)

func (p *Postgres) SaveTransaction(tr *transaction.Transaction) error {
	query := `
		INSERT INTO transactions (id, transtype, category, amount, transdate, description)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	ctx := context.Background()
	_, err := p.db.ExecWithRetry(ctx, retry.Strategy{Attempts: p.cfg.Attempts, Delay: p.cfg.Delay, Backoff: p.cfg.Backoffs}, query, tr.ID, tr.Type, tr.Category, tr.Amount, tr.Date, tr.Description)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("failed to insert transaction")
		return err
	}
	return nil
}

func (p *Postgres) GetTransaction(id string) (*transaction.Transaction, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		wbzlog.Logger.Warn().Str("id", id).Msg("invalid uuid")
		return nil, err
	}

	query := `
		SELECT id, transtype, category, amount, transdate, description
		FROM transactions
		WHERE id = $1
	`
	ctx := context.Background()
	row, err := p.db.QueryRowWithRetry(ctx, retry.Strategy{Attempts: p.cfg.Attempts, Delay: p.cfg.Delay, Backoff: p.cfg.Backoffs}, query, uid)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("failed to query transaction by id")
		return nil, err
	}
	var tr transaction.Transaction
	err = row.Scan(&tr.ID, &tr.Type, &tr.Category, &tr.Amount, &tr.Date, &tr.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		wbzlog.Logger.Error().Err(err).Msg("failed to get transaction by id")
		return nil, err
	}

	return &tr, nil
}

func (p *Postgres) GetAllTransactions(
	from, to time.Time,
	trtype, category, sortBy, sortDir string,
) ([]*transaction.Transaction, error) {

	query := `
		SELECT id, transtype, category, amount, transdate, description
		FROM transactions
		WHERE 1=1
	`
	args := []any{}
	argIndex := 1

	if !from.IsZero() {
		query += fmt.Sprintf(" AND transdate >= $%d", argIndex)
		args = append(args, from)
		argIndex++
	}

	if !to.IsZero() {
		query += fmt.Sprintf(" AND transdate <= $%d", argIndex)
		args = append(args, to)
		argIndex++
	}

	if trtype != "" {
		query += fmt.Sprintf(" AND transtype = $%d", argIndex)
		args = append(args, trtype)
		argIndex++
	}

	if category != "" {
		query += fmt.Sprintf(" AND category = $%d", argIndex)
		args = append(args, category)
	}

	if sortBy != "" {
		if sortBy == "type" {
			sortBy = "transtype"
		}
		if sortBy == "date" {
			sortBy = "transdate"
		}
		validSortColumns := map[string]bool{
			"id": true, "transtype": true, "category": true,
			"amount": true, "transdate": true,
		}
		if !validSortColumns[sortBy] {
			sortBy = "transdate"
		}

		dir := "DESC"
		if sortDir == "asc" {
			dir = "ASC"
		}

		query += fmt.Sprintf(" ORDER BY %s %s", sortBy, dir)
	} else {
		query += " ORDER BY transdate DESC"
	}

	ctx := context.Background()
	rows, err := p.db.QueryWithRetry(
		ctx,
		retry.Strategy{Attempts: p.cfg.Attempts, Delay: p.cfg.Delay, Backoff: p.cfg.Backoffs},
		query,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var result []*transaction.Transaction
	for rows.Next() {
		var tr transaction.Transaction
		if err := rows.Scan(&tr.ID, &tr.Type, &tr.Category, &tr.Amount, &tr.Date, &tr.Description); err != nil {
			return nil, err
		}
		result = append(result, &tr)
	}

	return result, rows.Err()
}

func (p *Postgres) UpdateTransaction(tr *transaction.Transaction) error {
	query := `
		UPDATE transactions
		SET transtype = $1, category = $2, amount = $3, transdate = $4, description = $5
		WHERE id = $6
	`
	ctx := context.Background()
	_, err := p.db.ExecWithRetry(ctx, retry.Strategy{Attempts: p.cfg.Attempts, Delay: p.cfg.Delay, Backoff: p.cfg.Backoffs}, query, tr.Type, tr.Category, tr.Amount, tr.Date, tr.Description, tr.ID)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("failed to update transaction")
		return err
	}
	return nil
}

func (p *Postgres) DeleteTransaction(id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		wbzlog.Logger.Warn().Str("id", id).Msg("invalid uuid")
		return err
	}
	ctx := context.Background()
	query := `DELETE FROM transactions WHERE id = $1`
	_, err = p.db.ExecWithRetry(ctx, retry.Strategy{Attempts: p.cfg.Attempts, Delay: p.cfg.Delay, Backoff: p.cfg.Backoffs}, query, uid)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("failed to delete transaction")
		return err
	}
	return nil
}
