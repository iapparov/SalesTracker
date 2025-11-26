package transactions

import (
	"encoding/csv"
	"fmt"
	"github.com/google/uuid"
	wbzlog "github.com/wb-go/wbf/zlog"
	"io"
	"salestracker/internal/domain/transaction"
	"time"
)

type TransactionService struct {
	repo TransactionStorageProvider
}

type TransactionStorageProvider interface {
	DeleteTransaction(id string) error
	GetTransaction(id string) (*transaction.Transaction, error)
	GetAllTransactions(from, to time.Time, trtype, category, sortBy, sortDir string) ([]*transaction.Transaction, error)
	SaveTransaction(tr *transaction.Transaction) error
	UpdateTransaction(tr *transaction.Transaction) error
}

func NewTransactionService(repo TransactionStorageProvider) *TransactionService {
	return &TransactionService{
		repo: repo,
	}
}

func (s *TransactionService) GetTransaction(id string) (*transaction.Transaction, error) {
	_, err := uuid.Parse(id)
	if err != nil {
		wbzlog.Logger.Warn().Str("id", id).Msg("invalid uuid")
		return nil, err
	}
	tr, err := s.repo.GetTransaction(id)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("repo get transaction error")
		return nil, err
	}
	return tr, nil
}

func (s *TransactionService) CreateTransaction(trType, category string, amount float64, date time.Time, descr string) (*transaction.Transaction, error) {
	tr, err := transaction.NewTransaction(transaction.TransactionType(trType), category, amount, descr, date)
	if err != nil {
		wbzlog.Logger.Warn().Err(err).Msg("invalid data for new transaction")
		return nil, err
	}
	err = s.repo.SaveTransaction(tr)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("repo save transaction error")
		return nil, err
	}
	return tr, nil
}

func (s *TransactionService) GetAllTransactions(from, to time.Time, trtype, category, sortBy, sortDir string) ([]*transaction.Transaction, error) {
	trs, err := s.repo.GetAllTransactions(from, to, trtype, category, sortBy, sortDir)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("repo get all transactions error")
		return nil, err
	}
	return trs, nil
}

func (s *TransactionService) PutTransaction(id string, trType string, category string, amount float64, date time.Time, descr string) (*transaction.Transaction, error) {
	_, err := uuid.Parse(id)
	if err != nil {
		wbzlog.Logger.Warn().Str("id", id).Msg("invalid uuid")
		return nil, err
	}
	tr, err := s.repo.GetTransaction(id)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("repo get (for put) transaction error")
		return nil, err
	}
	err = tr.TransactionChange(transaction.TransactionType(trType), category, amount, descr, date)
	if err != nil {
		wbzlog.Logger.Warn().Err(err).Msg("invalid data for transaction change")
		return nil, err
	}
	err = s.repo.UpdateTransaction(tr)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("repo update transaction error")
		return nil, err
	}
	return tr, err
}

func (s *TransactionService) DeleteTransaction(id string) error {
	_, err := uuid.Parse(id)
	if err != nil {
		wbzlog.Logger.Warn().Str("id", id).Msg("invalid uuid")
		return err
	}
	err = s.repo.DeleteTransaction(id)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("repo delete transaction error")
		return err
	}
	return nil
}

func (s *TransactionService) GetCSV(from, to time.Time, trtype, category, sortBy, sortDir string, output io.Writer) error {
	trs, err := s.repo.GetAllTransactions(from, to, trtype, category, sortBy, sortDir)
	if err != nil {
		wbzlog.Logger.Error().Err(err).Msg("repo get all transactions error")
		return err
	}

	writer := csv.NewWriter(output)
	defer writer.Flush()

	headers := []string{"ID", "Type", "Category", "Amount", "Date", "Description"}
	if err := writer.Write(headers); err != nil {
		wbzlog.Logger.Error().Err(err).Msg("error writing CSV headers")
		return err
	}

	for _, tr := range trs {
		row := []string{
			tr.ID.String(),
			string(tr.Type),
			tr.Category,
			fmt.Sprintf("%.2f", tr.Amount),
			tr.Date.Format(time.RFC3339),
			tr.Description,
		}
		if err := writer.Write(row); err != nil {
			wbzlog.Logger.Error().Err(err).Msg("error writing CSV row")
			return err
		}
	}

	wbzlog.Logger.Info().Msg("CSV generation completed")
	return nil
}
