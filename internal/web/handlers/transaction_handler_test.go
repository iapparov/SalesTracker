package handlers_test

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"net/http"
	"net/http/httptest"
	"salestracker/internal/domain/transaction"
	"salestracker/internal/web/dto"
	"salestracker/internal/web/handlers"
	"testing"
	"time"
)

// --------- MOCK SERVICE ---------

type MockTransactionService struct {
	CreateTransactionFn  func(trType, category string, amount float64, date time.Time, descr string) (*transaction.Transaction, error)
	GetAllTransactionsFn func(from, to time.Time, trtype, category, sortBy, sortDir string) ([]*transaction.Transaction, error)
	PutTransactionFn     func(id string, trType, category string, amount float64, date time.Time, descr string) (*transaction.Transaction, error)
	DeleteTransactionFn  func(id string) error
	GetCSVFn             func(from, to time.Time, trtype, category, sortBy, sortDir string, output io.Writer) error
	GetTransactionFn     func(id string) (*transaction.Transaction, error)
}

func (m *MockTransactionService) CreateTransaction(trType, category string, amount float64, date time.Time, descr string) (*transaction.Transaction, error) {
	return m.CreateTransactionFn(trType, category, amount, date, descr)
}
func (m *MockTransactionService) GetAllTransactions(from, to time.Time, trtype, category, sortBy, sortDir string) ([]*transaction.Transaction, error) {
	return m.GetAllTransactionsFn(from, to, trtype, category, sortBy, sortDir)
}
func (m *MockTransactionService) PutTransaction(id string, trType, category string, amount float64, date time.Time, descr string) (*transaction.Transaction, error) {
	return m.PutTransactionFn(id, trType, category, amount, date, descr)
}
func (m *MockTransactionService) DeleteTransaction(id string) error {
	return m.DeleteTransactionFn(id)
}
func (m *MockTransactionService) GetCSV(from, to time.Time, trtype, category, sortBy, sortDir string, output io.Writer) error {
	return m.GetCSVFn(from, to, trtype, category, sortBy, sortDir, output)
}
func (m *MockTransactionService) GetTransaction(id string) (*transaction.Transaction, error) {
	return m.GetTransactionFn(id)
}

// --------- UTILS ---------

func trperformRequest(hf func(*gin.Context), method, path string, body any, params map[string]string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req, _ := http.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	for k, v := range params {
		c.Params = append(c.Params, gin.Param{Key: k, Value: v})
	}
	hf(c)
	return w
}

// --------- TESTS ---------

func TestCreateTransaction_Success(t *testing.T) {
	mock := &MockTransactionService{
		CreateTransactionFn: func(trType, category string, amount float64, date time.Time, descr string) (*transaction.Transaction, error) {
			return &transaction.Transaction{Type: transaction.TransactionType(trType), Category: category, Amount: amount, Date: date, Description: descr}, nil
		},
	}
	h := handlers.NewTransactionHandler(mock)

	req := dto.SaveTransactionReq{
		Type:        "income",
		Category:    "food",
		Amount:      100,
		Date:        "2025-11-27",
		Description: "desc",
	}

	w := trperformRequest(h.CreateTransaction, "POST", "/transactions", req, nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestCreateTransaction_BadDate(t *testing.T) {
	mock := &MockTransactionService{}
	h := handlers.NewTransactionHandler(mock)
	req := dto.SaveTransactionReq{
		Type:        "income",
		Category:    "food",
		Amount:      100,
		Date:        "bad-date",
		Description: "desc",
	}
	w := trperformRequest(h.CreateTransaction, "POST", "/transactions", req, nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetTransaction_Success(t *testing.T) {
	mock := &MockTransactionService{
		GetTransactionFn: func(id string) (*transaction.Transaction, error) {
			return &transaction.Transaction{ID: uuid.New()}, nil
		},
	}
	h := handlers.NewTransactionHandler(mock)
	w := trperformRequest(h.GetTransaction, "GET", "/transactions/123", nil, map[string]string{"id": "123"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetTransaction_NotFound(t *testing.T) {
	mock := &MockTransactionService{
		GetTransactionFn: func(id string) (*transaction.Transaction, error) {
			return nil, nil
		},
	}
	h := handlers.NewTransactionHandler(mock)
	w := trperformRequest(h.GetTransaction, "GET", "/transactions/123", nil, map[string]string{"id": "123"})
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestDeleteTransaction_Success(t *testing.T) {
	mock := &MockTransactionService{
		DeleteTransactionFn: func(id string) error { return nil },
	}
	h := handlers.NewTransactionHandler(mock)
	w := trperformRequest(h.DeleteTransaction, "DELETE", "/transactions/123", nil, map[string]string{"id": "123"})
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestPutTransaction_Success(t *testing.T) {
	mock := &MockTransactionService{
		PutTransactionFn: func(id string, trType, category string, amount float64, date time.Time, descr string) (*transaction.Transaction, error) {
			return &transaction.Transaction{ID: uuid.New(), Type: transaction.TransactionType(trType)}, nil
		},
	}
	h := handlers.NewTransactionHandler(mock)
	req := dto.SaveTransactionReq{
		Type:        "expense",
		Category:    "food",
		Amount:      50,
		Date:        "2025-11-27",
		Description: "desc",
	}
	w := trperformRequest(h.PutTransaction, "PUT", "/transactions/123", req, map[string]string{"id": "123"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetAllTransactions_Success(t *testing.T) {
	mock := &MockTransactionService{
		GetAllTransactionsFn: func(from, to time.Time, trtype, category, sortBy, sortDir string) ([]*transaction.Transaction, error) {
			return []*transaction.Transaction{
				{ID: uuid.New(), Type: transaction.Income},
			}, nil
		},
	}
	h := handlers.NewTransactionHandler(mock)
	w := trperformRequest(h.GetAllTransactions, "GET", "/transactions", nil, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetCSVTr_Success(t *testing.T) {
	mock := &MockTransactionService{
		GetCSVFn: func(from, to time.Time, trtype, category, sortBy, sortDir string, output io.Writer) error {
			_, err := output.Write([]byte("csv data"))
			return err
		},
	}
	h := handlers.NewTransactionHandler(mock)
	w := trperformRequest(h.GetCSV, "GET", "/transactions/csv", nil, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "csv data" {
		t.Fatalf("expected csv data, got %s", w.Body.String())
	}
}
