package transaction

import (
	"testing"
	"time"
)

func TestNewTransaction_Valid(t *testing.T) {
	date := time.Date(2025, 11, 27, 12, 0, 0, 0, time.UTC)
	tr, err := NewTransaction(Income, "salary", 100.0, "desc", date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr.Type != Income || tr.Category != "salary" || tr.Amount != 100.0 || tr.Description != "desc" || !tr.Date.Equal(date) {
		t.Fatal("transaction fields mismatch")
	}
}

func TestNewTransaction_InvalidType(t *testing.T) {
	_, err := NewTransaction("invalid", "salary", 100.0, "desc", time.Now())
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestNewTransaction_EmptyCategory(t *testing.T) {
	_, err := NewTransaction(Income, "", 100.0, "desc", time.Now())
	if err == nil {
		t.Fatal("expected error for empty category")
	}
}

func TestNewTransaction_NonPositiveAmount(t *testing.T) {
	_, err := NewTransaction(Income, "cat", 0, "desc", time.Now())
	if err == nil {
		t.Fatal("expected error for non-positive amount")
	}
}

func TestNewTransaction_ZeroDate(t *testing.T) {
	tr, err := NewTransaction(Income, "cat", 10.0, "desc", time.Time{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr.Date.IsZero() {
		t.Fatal("expected date to be set to now")
	}
}

func TestTransactionChange_Valid(t *testing.T) {
	tr, _ := NewTransaction(Income, "cat", 10.0, "desc", time.Now())
	newDate := time.Date(2025, 11, 27, 10, 0, 0, 0, time.UTC)
	err := tr.TransactionChange(Expense, "food", 50, "lunch", newDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr.Type != Expense || tr.Category != "food" || tr.Amount != 50 || tr.Description != "lunch" || !tr.Date.Equal(newDate) {
		t.Fatal("transaction fields not updated correctly")
	}
}

func TestTransactionChange_InvalidType(t *testing.T) {
	tr, _ := NewTransaction(Income, "cat", 10.0, "desc", time.Now())
	err := tr.TransactionChange("invalid", "cat", 10, "desc", time.Now())
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestTransactionChange_EmptyCategory(t *testing.T) {
	tr, _ := NewTransaction(Income, "cat", 10.0, "desc", time.Now())
	err := tr.TransactionChange(Income, "", 10, "desc", time.Now())
	if err == nil {
		t.Fatal("expected error for empty category")
	}
}

func TestTransactionChange_NonPositiveAmount(t *testing.T) {
	tr, _ := NewTransaction(Income, "cat", 10.0, "desc", time.Now())
	err := tr.TransactionChange(Income, "cat", 0, "desc", time.Now())
	if err == nil {
		t.Fatal("expected error for non-positive amount")
	}
}

func TestTransactionChange_ZeroDate(t *testing.T) {
	tr, _ := NewTransaction(Income, "cat", 10.0, "desc", time.Now())
	err := tr.TransactionChange(Income, "cat", 20, "desc2", time.Time{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr.Date.IsZero() {
		t.Fatal("expected date to be updated to now")
	}
	if tr.Amount != 20 || tr.Description != "desc2" {
		t.Fatal("fields not updated correctly")
	}
}
