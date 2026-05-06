package api

import (
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/davidcarrington/eagle-bank/internal/api/middleware"
	"github.com/davidcarrington/eagle-bank/internal/domain"
	"github.com/davidcarrington/eagle-bank/internal/service"
)

var transactionIDRe = regexp.MustCompile(`^tan-[A-Za-z0-9]+$`)

type transactionHandler struct {
	txns *service.TransactionService
}

// Request types

type createTransactionRequest struct {
	Amount    *float64 `json:"amount" binding:"required,gte=0,lte=10000"`
	Currency  string   `json:"currency" binding:"required,oneof=GBP"`
	Type      string   `json:"type" binding:"required,oneof=deposit withdrawal"`
	Reference string   `json:"reference"`
}

// Response types

type transactionResponse struct {
	ID               string  `json:"id"`
	Amount           float64 `json:"amount"`
	Currency         string  `json:"currency"`
	Type             string  `json:"type"`
	Reference        string  `json:"reference,omitempty"`
	UserID           string  `json:"userId,omitempty"`
	CreatedTimestamp string  `json:"createdTimestamp"`
}

type listTransactionsResponse struct {
	Transactions []transactionResponse `json:"transactions"`
}

func toTransactionResponse(t *domain.Transaction) transactionResponse {
	return transactionResponse{
		ID:               t.ID,
		Amount:           t.Amount.GBPFloat(),
		Currency:         t.Currency,
		Type:             t.Type,
		Reference:        t.Reference,
		UserID:           t.UserID,
		CreatedTimestamp: t.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func (h *transactionHandler) requireValidTransactionID(c *gin.Context) (string, bool) {
	id := c.Param("transactionId")
	if !transactionIDRe.MatchString(id) {
		c.JSON(http.StatusBadRequest, badRequestResponse{
			Message: "validation failed",
			Details: []validationDetail{{
				Field:   "transactionId",
				Message: "transactionId must match pattern ^tan-[A-Za-z0-9]+$",
				Type:    "pattern",
			}},
		})
		return "", false
	}
	return id, true
}

// Handlers

func (h *transactionHandler) createTransaction(c *gin.Context) {
	accNum, ok := requireValidAccountNumber(c)
	if !ok {
		return
	}
	var req createTransactionRequest
	if !bindJSON(c, &req) {
		return
	}

	txn, err := h.txns.Create(c.Request.Context(), middleware.CallerID(c), accNum, service.CreateTransactionInput{
		Amount:    domain.FromGBPFloat(*req.Amount),
		Currency:  req.Currency,
		Type:      req.Type,
		Reference: req.Reference,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toTransactionResponse(txn))
}

func (h *transactionHandler) listTransactions(c *gin.Context) {
	accNum, ok := requireValidAccountNumber(c)
	if !ok {
		return
	}

	txns, err := h.txns.List(c.Request.Context(), middleware.CallerID(c), accNum)
	if err != nil {
		writeError(c, err)
		return
	}

	resp := listTransactionsResponse{Transactions: make([]transactionResponse, 0, len(txns))}
	for _, t := range txns {
		resp.Transactions = append(resp.Transactions, toTransactionResponse(t))
	}
	c.JSON(http.StatusOK, resp)
}

func (h *transactionHandler) getTransaction(c *gin.Context) {
	accNum, ok := requireValidAccountNumber(c)
	if !ok {
		return
	}
	txnID, ok := h.requireValidTransactionID(c)
	if !ok {
		return
	}

	txn, err := h.txns.Get(c.Request.Context(), middleware.CallerID(c), accNum, txnID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toTransactionResponse(txn))
}
