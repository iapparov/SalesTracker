package handlers

import (
	"fmt"
	wbgin "github.com/wb-go/wbf/ginext"
	"io"
	"net/http"
	"salestracker/internal/domain/transaction"
	"salestracker/internal/web/dto"
	"time"
)

type TransactionHandler struct {
	Service TransactionIFace
}

type TransactionIFace interface {
	CreateTransaction(trType, category string, amount float64, date time.Time, descr string) (*transaction.Transaction, error)
	GetAllTransactions(from, to time.Time, trtype, category, sortBy, sortDir string) ([]*transaction.Transaction, error)
	PutTransaction(id string, trType string, category string, amount float64, date time.Time, descr string) (*transaction.Transaction, error)
	DeleteTransaction(id string) error
	GetCSV(from, to time.Time, trtype, category, sortBy, sortDir string, output io.Writer) error
	GetTransaction(id string) (*transaction.Transaction, error)
}

func NewTransactionHandler(service TransactionIFace) *TransactionHandler {
	return &TransactionHandler{
		Service: service,
	}
}

func (h *TransactionHandler) CreateTransaction(ctx *wbgin.Context) {
	var req dto.SaveTransactionReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, wbgin.H{"error": err.Error()})
		return
	}
	layout := "2006-01-02"
	trDate, err := time.ParseInLocation(layout, req.Date, time.Local)
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusBadRequest, wbgin.H{"error": "invalid date format"})
		return
	}
	res, err := h.Service.CreateTransaction(
		req.Type,
		req.Category,
		req.Amount,
		trDate,
		req.Description,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, wbgin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (h *TransactionHandler) DeleteTransaction(ctx *wbgin.Context) {
	trxId := ctx.Param("id")

	err := h.Service.DeleteTransaction(trxId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, wbgin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusNoContent, wbgin.H{"status": "deleted"})
}

func (h *TransactionHandler) PutTransaction(ctx *wbgin.Context) {
	var req dto.SaveTransactionReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, wbgin.H{"error": err.Error()})
		return
	}

	trxId, ok := ctx.Params.Get("id")
	if !ok {
		ctx.JSON(http.StatusBadRequest, wbgin.H{"error": "missing transaction id"})
		return
	}

	layout := "2006-01-02"
	trDate, err := time.ParseInLocation(layout, req.Date, time.Local)
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusBadRequest, wbgin.H{"error": "invalid date format"})
		return
	}

	res, err := h.Service.PutTransaction(
		trxId,
		req.Type,
		req.Category,
		req.Amount,
		trDate,
		req.Description,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, wbgin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, res)
}

func (h *TransactionHandler) GetTransaction(ctx *wbgin.Context) {
	id := ctx.Param("id")
	if id != "" {
		tr, err := h.Service.GetTransaction(id)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, wbgin.H{"error": err.Error()})
			return
		}
		if tr == nil {
			ctx.JSON(http.StatusNotFound, wbgin.H{"error": "transaction not found"})
			return
		}
		ctx.JSON(http.StatusOK, tr)
		return
	} else {
		ctx.JSON(http.StatusBadRequest, wbgin.H{"error": "missing transaction id"})
		return
	}
}

func (h *TransactionHandler) GetAllTransactions(ctx *wbgin.Context) {
	var req dto.GetTransactionReq
	req.From = ctx.Query("from")
	req.To = ctx.Query("to")
	req.Type = ctx.Query("type")
	req.Category = ctx.Query("category")
	req.SortBy = ctx.Query("sortBy")
	req.SortDir = ctx.Query("sortDir")

	var from time.Time
	var err error
	layout := "2006-01-02"
	if req.From != "" {
		from, err = time.ParseInLocation(layout, req.From, time.Local)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, wbgin.H{"error": "invalid from date format"})
			return
		}
	}
	var to time.Time
	if req.To != "" {
		to, err = time.ParseInLocation(layout, req.To, time.Local)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, wbgin.H{"error": "invalid to date format"})
			return
		}
	}

	res, err := h.Service.GetAllTransactions(from, to, req.Type, req.Category, req.SortBy, req.SortDir)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, wbgin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (h *TransactionHandler) GetCSV(ctx *wbgin.Context) {
	var req dto.GetTransactionReq
	req.From = ctx.Query("from")
	req.To = ctx.Query("to")
	req.Type = ctx.Query("type")
	req.Category = ctx.Query("category")
	req.SortBy = ctx.Query("sortBy")
	req.SortDir = ctx.Query("sortDir")

	var from time.Time
	var err error
	layout := "2006-01-02"
	if req.From != "" {
		from, err = time.ParseInLocation(layout, req.From, time.Local)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, wbgin.H{"error": "invalid from date format"})
			return
		}
	}
	var to time.Time
	if req.To != "" {
		to, err = time.ParseInLocation(layout, req.To, time.Local)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, wbgin.H{"error": "invalid to date format"})
			return
		}
	}

	ctx.Writer.Header().Set("Content-Disposition", "attachment; filename=transactions.csv")
	ctx.Writer.Header().Set("Content-Type", "text/csv")

	err = h.Service.GetCSV(from, to, req.Type, req.Category, req.SortBy, req.SortDir, ctx.Writer)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, wbgin.H{"error": err.Error()})
		return
	}
}
