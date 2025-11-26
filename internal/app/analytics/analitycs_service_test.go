package analytics

import (
	"bytes"
	"errors"
	"salestracker/internal/domain/analytic"
	"testing"
	"time"
)

// --- Mock repository ---
type mockRepo struct {
	Analytics *analytic.Analytics
	Err       error
}

func (m *mockRepo) GetAnalytics(from, to time.Time, groupBy, splitBy, sortBy, sortDir string) (*analytic.Analytics, error) {
	return m.Analytics, m.Err
}

// --- Helpers ---
func sampleAnalytics() *analytic.Analytics {
	return &analytic.Analytics{
		Groups: []analytic.AnalyticGroup{
			{
				GroupKey: "2025-11-27",
				Data: analytic.AnalyticByType{
					Income: analytic.Analytic{
						Sum: 100, Avg: 50, Count: 2, Median: 50, Percentile90: 90,
					},
					Expense: analytic.Analytic{
						Sum: 40, Avg: 20, Count: 2, Median: 20, Percentile90: 35,
					},
					All: analytic.Analytic{
						Sum: 60, Avg: 30, Count: 4, Median: 35, Percentile90: 62.5,
					},
				},
			},
		},
	}
}

func TestGetAnalytics_InvalidDateRange(t *testing.T) {
	svc := NewAnalyticService(&mockRepo{})
	from := time.Now()
	to := from.Add(-time.Hour)

	_, err := svc.GetAnalytics(from, to, "", "", "", "")
	if err == nil {
		t.Fatal("expected error for invalid date range")
	}
}

func TestGetAnalytics_RepoError(t *testing.T) {
	svc := NewAnalyticService(&mockRepo{Err: errors.New("repo failure")})
	from := time.Now()
	to := from.Add(time.Hour)

	_, err := svc.GetAnalytics(from, to, "", "", "", "")
	if err == nil || err.Error() != "repo failure" {
		t.Fatal("expected repo error")
	}
}

func TestGetAnalytics_Success(t *testing.T) {
	mockData := sampleAnalytics()
	svc := NewAnalyticService(&mockRepo{Analytics: mockData})
	from := time.Now()
	to := from.Add(time.Hour)

	result, err := svc.GetAnalytics(from, to, "", "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Groups) != 1 || result.Groups[0].GroupKey != "2025-11-27" {
		t.Fatal("unexpected analytics data")
	}
}

func TestGetCSV_RepoError(t *testing.T) {
	svc := NewAnalyticService(&mockRepo{Err: errors.New("repo fail")})
	var buf bytes.Buffer
	from := time.Now()
	to := from.Add(time.Hour)

	err := svc.GetCSV(from, to, "", "", "", "", &buf)
	if err == nil || err.Error() != "repo fail" {
		t.Fatal("expected repo error")
	}
}

func TestGetCSV_Success(t *testing.T) {
	mockData := sampleAnalytics()
	svc := NewAnalyticService(&mockRepo{Analytics: mockData})
	var buf bytes.Buffer
	from := time.Now()
	to := from.Add(time.Hour)

	err := svc.GetCSV(from, to, "", "", "", "", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	// Проверяем, что CSV содержит заголовки
	if !bytes.Contains([]byte(out), []byte("GroupKey,Type,Sum,Avg,Count,Median,Percentile90")) {
		t.Fatal("CSV headers missing")
	}
	// Проверяем, что CSV содержит Income и Expense
	if !bytes.Contains([]byte(out), []byte("Income")) || !bytes.Contains([]byte(out), []byte("Expense")) {
		t.Fatal("CSV content missing Income/Expense")
	}
	// Проверяем All
	if !bytes.Contains([]byte(out), []byte("All")) {
		t.Fatal("CSV content missing All")
	}
}
