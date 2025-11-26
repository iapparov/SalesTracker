package transactions

import (
	"bytes"
	"errors"
	"github.com/google/uuid"
	"salestracker/internal/domain/transaction"
	"testing"
	"time"
)

// --- Mock repository ---
type mockRepo struct {
	GetTr     *transaction.Transaction
	GetAllTrs []*transaction.Transaction
	Err       error
	UpdatedTr *transaction.Transaction
	DeletedID string
	SavedTr   *transaction.Transaction
}

func (m *mockRepo) GetTransaction(id string) (*transaction.Transaction, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.GetTr, nil
}
func (m *mockRepo) GetAllTransactions(from, to time.Time, trtype, category, sortBy, sortDir string) ([]*transaction.Transaction, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.GetAllTrs, nil
}
func (m *mockRepo) SaveTransaction(tr *transaction.Transaction) error {
	if m.Err != nil {
		return m.Err
	}
	m.SavedTr = tr
	return nil
}
func (m *mockRepo) UpdateTransaction(tr *transaction.Transaction) error {
	if m.Err != nil {
		return m.Err
	}
	m.UpdatedTr = tr
	return nil
}
func (m *mockRepo) DeleteTransaction(id string) error {
	if m.Err != nil {
		return m.Err
	}
	m.DeletedID = id
	return nil
}

// --- Helpers ---
func sampleTransaction(t *testing.T) *transaction.Transaction {
	tr, err := transaction.NewTransaction("income", "salary", 100.0, "desc", time.Now())
	if err != nil {
		t.Fatalf("failed to create sample transaction: %v", err)
	}
	return tr
}

func TestGetTransaction_InvalidUUID(t *testing.T) {
	svc := NewTransactionService(&mockRepo{})
	_, err := svc.GetTransaction("invalid-uuid")
	if err == nil {
		t.Fatal("expected error for invalid UUID")
	}
}

func TestGetTransaction_RepoError(t *testing.T) {
	svc := NewTransactionService(&mockRepo{Err: errors.New("repo fail")})
	id := uuid.New().String()
	_, err := svc.GetTransaction(id)
	if err == nil || err.Error() != "repo fail" {
		t.Fatal("expected repo error")
	}
}

func TestGetTransaction_Success(t *testing.T) {
	tr := sampleTransaction(t)
	svc := NewTransactionService(&mockRepo{GetTr: tr})
	res, err := svc.GetTransaction(tr.ID.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ID != tr.ID {
		t.Fatal("unexpected transaction returned")
	}
}

func TestCreateTransaction_RepoError(t *testing.T) {
	svc := NewTransactionService(&mockRepo{Err: errors.New("repo fail")})
	_, err := svc.CreateTransaction("income", "cat", 10, time.Now(), "desc")
	if err == nil || err.Error() != "repo fail" {
		t.Fatal("expected repo error")
	}
}

func TestCreateTransaction_Success(t *testing.T) {
	svc := NewTransactionService(&mockRepo{})
	tr, err := svc.CreateTransaction("income", "cat", 10, time.Now(), "desc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr.Type != "income" {
		t.Fatal("transaction type mismatch")
	}
}

func TestPutTransaction_InvalidUUID(t *testing.T) {
	svc := NewTransactionService(&mockRepo{})
	_, err := svc.PutTransaction("bad-uuid", "income", "cat", 10, time.Now(), "desc")
	if err == nil {
		t.Fatal("expected error for invalid UUID")
	}
}

func TestPutTransaction_RepoGetError(t *testing.T) {
	svc := NewTransactionService(&mockRepo{Err: errors.New("get fail")})
	id := uuid.New().String()
	_, err := svc.PutTransaction(id, "income", "cat", 10, time.Now(), "desc")
	if err == nil || err.Error() != "get fail" {
		t.Fatal("expected repo get error")
	}
}

func TestPutTransaction_Success(t *testing.T) {
	tr := sampleTransaction(t)
	svc := NewTransactionService(&mockRepo{GetTr: tr})
	newAmount := 200.0
	res, err := svc.PutTransaction(tr.ID.String(), "income", "cat", newAmount, time.Now(), "updated")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Amount != newAmount {
		t.Fatal("transaction not updated")
	}
}

func TestDeleteTransaction_InvalidUUID(t *testing.T) {
	svc := NewTransactionService(&mockRepo{})
	err := svc.DeleteTransaction("bad-uuid")
	if err == nil {
		t.Fatal("expected error for invalid UUID")
	}
}

func TestDeleteTransaction_RepoError(t *testing.T) {
	id := uuid.New().String()
	svc := NewTransactionService(&mockRepo{Err: errors.New("delete fail")})
	err := svc.DeleteTransaction(id)
	if err == nil || err.Error() != "delete fail" {
		t.Fatal("expected repo delete error")
	}
}

func TestDeleteTransaction_Success(t *testing.T) {
	id := uuid.New().String()
	svc := NewTransactionService(&mockRepo{})
	err := svc.DeleteTransaction(id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc.repo.(*mockRepo).DeletedID != id {
		t.Fatal("transaction ID not recorded in mock delete")
	}
}

func TestGetAllTransactions_RepoError(t *testing.T) {
	svc := NewTransactionService(&mockRepo{Err: errors.New("fail")})
	_, err := svc.GetAllTransactions(time.Now(), time.Now(), "", "", "", "")
	if err == nil || err.Error() != "fail" {
		t.Fatal("expected repo error")
	}
}

func TestGetAllTransactions_Success(t *testing.T) {
	trs := []*transaction.Transaction{sampleTransaction(nil)}
	svc := NewTransactionService(&mockRepo{GetAllTrs: trs})
	res, err := svc.GetAllTransactions(time.Now(), time.Now(), "", "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 1 {
		t.Fatal("unexpected number of transactions returned")
	}
}

func TestGetCSV_Success(t *testing.T) {
	tr := sampleTransaction(nil)
	svc := NewTransactionService(&mockRepo{GetAllTrs: []*transaction.Transaction{tr}})
	var buf bytes.Buffer
	err := svc.GetCSV(time.Now(), time.Now(), "", "", "", "", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if out == "" || !bytes.Contains([]byte(out), []byte(tr.ID.String())) {
		t.Fatal("CSV output incorrect")
	}
}
