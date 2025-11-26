package handlers_test

import (
	"errors"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/http/httptest"
	"salestracker/internal/domain/analytic"
	"salestracker/internal/web/handlers"
	"testing"
	"time"
)

// ---------------- MOCK --------------------

type MockAnalyticsService struct {
	GetAnalyticsFn func(from, to time.Time, groupBy, splitBy, sortBy, sortDir string) (*analytic.Analytics, error)
	GetCSVFn       func(from, to time.Time, groupBy, splitBy, sortBy, sortDir string, output io.Writer) error
}

func (m *MockAnalyticsService) GetAnalytics(from, to time.Time, groupBy, splitBy, sortBy, sortDir string) (*analytic.Analytics, error) {
	return m.GetAnalyticsFn(from, to, groupBy, splitBy, sortBy, sortDir)
}

func (m *MockAnalyticsService) GetCSV(from, to time.Time, groupBy, splitBy, sortBy, sortDir string, output io.Writer) error {
	return m.GetCSVFn(from, to, groupBy, splitBy, sortBy, sortDir, output)
}

// ---------------- UTILS --------------------

func performRequest(hf func(*gin.Context), method, path string, query map[string]string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	q := req.URL.Query()
	for k, v := range query {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	hf(c)
	return w
}

// ---------------- TESTS --------------------

func TestGetAnalys_Success(t *testing.T) {
	mockSvc := &MockAnalyticsService{
		GetAnalyticsFn: func(from, to time.Time, groupBy, splitBy, sortBy, sortDir string) (*analytic.Analytics, error) {
			return &analytic.Analytics{
				Groups: []analytic.AnalyticGroup{
					{
						GroupKey: "group1",
						Data: analytic.AnalyticByType{
							Income:  analytic.Analytic{Sum: 100},
							Expense: analytic.Analytic{Sum: 50},
							All:     analytic.Analytic{Sum: 150},
						},
					},
				},
			}, nil
		},
	}

	h := handlers.NewAnalyticHandler(mockSvc)

	w := performRequest(h.GetAnalys, "GET", "/analytics", map[string]string{
		"from": "2025-11-01",
		"to":   "2025-11-27",
	})

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetAnalys_MissingFrom(t *testing.T) {
	h := handlers.NewAnalyticHandler(&MockAnalyticsService{})
	w := performRequest(h.GetAnalys, "GET", "/analytics", map[string]string{
		"to": "2025-11-27",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetAnalys_InvalidDate(t *testing.T) {
	h := handlers.NewAnalyticHandler(&MockAnalyticsService{})
	w := performRequest(h.GetAnalys, "GET", "/analytics", map[string]string{
		"from": "bad-date",
		"to":   "2025-11-27",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetAnalys_ServiceError(t *testing.T) {
	mockSvc := &MockAnalyticsService{
		GetAnalyticsFn: func(from, to time.Time, groupBy, splitBy, sortBy, sortDir string) (*analytic.Analytics, error) {
			return nil, errors.New("service failed")
		},
	}
	h := handlers.NewAnalyticHandler(mockSvc)
	w := performRequest(h.GetAnalys, "GET", "/analytics", map[string]string{
		"from": "2025-11-01",
		"to":   "2025-11-27",
	})
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestGetCSV_Success(t *testing.T) {
	mockSvc := &MockAnalyticsService{
		GetCSVFn: func(from, to time.Time, groupBy, splitBy, sortBy, sortDir string, output io.Writer) error {
			// просто пишем что-то в writer
			_, err := output.Write([]byte("csv data"))
			return err
		},
	}
	h := handlers.NewAnalyticHandler(mockSvc)

	w := performRequest(h.GetCSV, "GET", "/analytics/csv", map[string]string{
		"from": "2025-11-01",
		"to":   "2025-11-27",
	})

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "csv data" {
		t.Fatalf("expected csv data, got %s", w.Body.String())
	}
}

func TestGetCSV_MissingTo(t *testing.T) {
	h := handlers.NewAnalyticHandler(&MockAnalyticsService{})
	w := performRequest(h.GetCSV, "GET", "/analytics/csv", map[string]string{
		"from": "2025-11-01",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestGetCSV_InvalidFrom(t *testing.T) {
	h := handlers.NewAnalyticHandler(&MockAnalyticsService{})
	w := performRequest(h.GetCSV, "GET", "/analytics/csv", map[string]string{
		"from": "bad-date",
		"to":   "2025-11-27",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
