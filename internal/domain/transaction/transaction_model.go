package transaction

import (
	"errors"
	"github.com/google/uuid"
	"time"
)

type TransactionType string

const (
	Income  TransactionType = "income"
	Expense TransactionType = "expense"
)

type Transaction struct {
	ID          uuid.UUID       `json:"ID"`
	Type        TransactionType `json:"Type"`
	Category    string          `json:"Category"`
	Amount      float64         `json:"Amount"`
	Date        time.Time       `json:"Date"`
	Description string          `json:"Description"`
}

func NewTransaction(trType TransactionType, Category string, Amount float64, Description string, Date time.Time) (*Transaction, error) {
	if trType != Income && trType != Expense {
		return nil, errors.New("invalid transaction type")
	}
	if Category == "" {
		return nil, errors.New("category cant be empty")
	}
	if Amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	var t time.Time
	if Date.IsZero() {
		t = time.Now()
	} else {
		t = Date
	}
	return &Transaction{
		ID:          uuid.New(),
		Type:        trType,
		Category:    Category,
		Amount:      Amount,
		Date:        t,
		Description: Description,
	}, nil
}

func (t *Transaction) TransactionChange(trType TransactionType, category string, amount float64, description string, date time.Time) error {

	if trType != Income && trType != Expense {
		return errors.New("invalid transaction type")
	}

	if category == "" {
		return errors.New("category cannot be empty")
	}

	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	if date.IsZero() {
		date = time.Now()
	}

	t.Type = trType
	t.Category = category
	t.Amount = amount
	t.Description = description
	t.Date = date

	return nil
}
