package handlers

import (
	wbgin "github.com/wb-go/wbf/ginext"
	"io"
	"net/http"
	"salestracker/internal/domain/analytic"
	"salestracker/internal/web/dto"
	"time"
)

// AnalyticsHandler обрабатывает запросы для аналитики транзакций
type AnalyticsHandler struct {
	Service AnalyticsIFace
}

// AnalyticsIFace описывает интерфейс сервиса аналитики
type AnalyticsIFace interface {
	GetAnalytics(from, to time.Time, groupBy, splitBy, sortBy, sortDir string) (*analytic.Analytics, error)
	GetCSV(from, to time.Time, groupBy, splitBy, sortBy, sortDir string, output io.Writer) error
}

// NewAnalyticHandler создает новый AnalyticsHandler
func NewAnalyticHandler(service AnalyticsIFace) *AnalyticsHandler {
	return &AnalyticsHandler{
		Service: service,
	}
}

// GetAnalys godoc
// @Summary Получить агрегированную аналитику
// @Description Возвращает агрегированные данные транзакций за указанный период с возможностью группировки, разделения и сортировки
// @Tags Analytics
// @Produce json
// @Param from query string true "Дата начала (YYYY-MM-DD)"
// @Param to query string true "Дата конца (YYYY-MM-DD)"
// @Param groupby query string false "Группировка (day/week/month/category)"
// @Param splitby query string false "Разделение данных (например по типу транзакции)"
// @Param sortby query string false "Поле для сортировки"
// @Param sortdir query string false "Направление сортировки (asc/desc)"
// @Success 200 {object} analytic.Analytics
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/analytics [get]
func (h *AnalyticsHandler) GetAnalys(ctx *wbgin.Context) {
	var AnalyticsReq dto.AnalyticsReq
	AnalyticsReq.From = ctx.Query("from")
	if AnalyticsReq.From == "" {
		ctx.JSON(http.StatusBadRequest, wbgin.H{"error": "missing from date"})
		return
	}
	AnalyticsReq.To = ctx.Query("to")
	if AnalyticsReq.To == "" {
		ctx.JSON(http.StatusBadRequest, wbgin.H{"error": "missing to date"})
		return
	}
	AnalyticsReq.GroupBy = ctx.Query("groupby")
	AnalyticsReq.SplitBy = ctx.Query("splitby")
	AnalyticsReq.SortBy = ctx.Query("sortby")
	AnalyticsReq.SortDir = ctx.Query("sortdir")

	layout := "2006-01-02"
	from, err := time.ParseInLocation(layout, AnalyticsReq.From, time.Local)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, wbgin.H{"error": "invalid from date format"})
		return
	}
	to, err := time.ParseInLocation(layout, AnalyticsReq.To, time.Local)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, wbgin.H{"error": "invalid to date format"})
		return
	}

	res, err := h.Service.GetAnalytics(from, to, AnalyticsReq.GroupBy, AnalyticsReq.SplitBy, AnalyticsReq.SortBy, AnalyticsReq.SortDir)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, wbgin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, res)
}

// GetCSV godoc
// @Summary Экспорт аналитики в CSV
// @Description Экспортирует агрегированные данные транзакций за указанный период в CSV-файл
// @Tags Analytics
// @Param from query string true "Дата начала (YYYY-MM-DD)"
// @Param to query string true "Дата конца (YYYY-MM-DD)"
// @Param groupby query string false "Группировка (day/week/month/category)"
// @Param splitby query string false "Разделение данных (например по типу транзакции)"
// @Param sortby query string false "Поле для сортировки"
// @Param sortdir query string false "Направление сортировки (asc/desc)"
// @Success 200 {file} file "CSV файл"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/analytics/export [get]
func (h *AnalyticsHandler) GetCSV(ctx *wbgin.Context) {
	var AnalyticsReq dto.AnalyticsReq
	AnalyticsReq.From = ctx.Query("from")
	if AnalyticsReq.From == "" {
		ctx.JSON(http.StatusBadRequest, wbgin.H{"error": "missing from date"})
		return
	}
	AnalyticsReq.To = ctx.Query("to")
	if AnalyticsReq.To == "" {
		ctx.JSON(http.StatusBadRequest, wbgin.H{"error": "missing to date"})
		return
	}
	AnalyticsReq.GroupBy = ctx.Query("groupby")
	AnalyticsReq.SplitBy = ctx.Query("splitby")
	AnalyticsReq.SortBy = ctx.Query("sortby")
	AnalyticsReq.SortDir = ctx.Query("sortdir")

	layout := "2006-01-02"
	from, err := time.ParseInLocation(layout, AnalyticsReq.From, time.Local)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, wbgin.H{"error": "invalid from date format"})
		return
	}
	to, err := time.ParseInLocation(layout, AnalyticsReq.To, time.Local)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, wbgin.H{"error": "invalid to date format"})
		return
	}

	ctx.Writer.Header().Set("Content-Disposition", "attachment; filename=transactions.csv")
	ctx.Writer.Header().Set("Content-Type", "text/csv")
	err = h.Service.GetCSV(from, to, AnalyticsReq.GroupBy, AnalyticsReq.SplitBy, AnalyticsReq.SortBy, AnalyticsReq.SortDir, ctx.Writer)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, wbgin.H{"error": err.Error()})
		return
	}
}
