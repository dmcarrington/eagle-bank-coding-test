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

var accountNumberRe = regexp.MustCompile(`^01\d{6}$`)

type accountHandler struct {
	accounts *service.AccountService
}

// Request types

type createAccountRequest struct {
	Name        string `json:"name" binding:"required"`
	AccountType string `json:"accountType" binding:"required,oneof=personal"`
}

type updateAccountRequest struct {
	Name        *string `json:"name"`
	AccountType *string `json:"accountType" binding:"omitempty,oneof=personal"`
}

// Response types

type accountResponse struct {
	AccountNumber    string  `json:"accountNumber"`
	SortCode         string  `json:"sortCode"`
	Name             string  `json:"name"`
	AccountType      string  `json:"accountType"`
	Balance          float64 `json:"balance"`
	Currency         string  `json:"currency"`
	CreatedTimestamp string  `json:"createdTimestamp"`
	UpdatedTimestamp string  `json:"updatedTimestamp"`
}

type listAccountsResponse struct {
	Accounts []accountResponse `json:"accounts"`
}

func toAccountResponse(a *domain.Account) accountResponse {
	return accountResponse{
		AccountNumber:    a.AccountNumber,
		SortCode:         a.SortCode,
		Name:             a.Name,
		AccountType:      a.AccountType,
		Balance:          a.Balance.GBPFloat(),
		Currency:         a.Currency,
		CreatedTimestamp: a.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedTimestamp: a.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func requireValidAccountNumber(c *gin.Context) (string, bool) {
	n := c.Param("accountNumber")
	if !accountNumberRe.MatchString(n) {
		c.JSON(http.StatusBadRequest, badRequestResponse{
			Message: "validation failed",
			Details: []validationDetail{{
				Field:   "accountNumber",
				Message: "accountNumber must match pattern ^01\\d{6}$",
				Type:    "pattern",
			}},
		})
		return "", false
	}
	return n, true
}

// Handlers

func (h *accountHandler) createAccount(c *gin.Context) {
	var req createAccountRequest
	if !bindJSON(c, &req) {
		return
	}

	acc, err := h.accounts.Create(c.Request.Context(), middleware.CallerID(c), service.CreateAccountInput{
		Name:        req.Name,
		AccountType: req.AccountType,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toAccountResponse(acc))
}

func (h *accountHandler) listAccounts(c *gin.Context) {
	accounts, err := h.accounts.List(c.Request.Context(), middleware.CallerID(c))
	if err != nil {
		writeError(c, err)
		return
	}

	resp := listAccountsResponse{Accounts: make([]accountResponse, 0, len(accounts))}
	for _, a := range accounts {
		resp.Accounts = append(resp.Accounts, toAccountResponse(a))
	}
	c.JSON(http.StatusOK, resp)
}

func (h *accountHandler) getAccount(c *gin.Context) {
	n, ok := requireValidAccountNumber(c)
	if !ok {
		return
	}
	acc, err := h.accounts.Get(c.Request.Context(), middleware.CallerID(c), n)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toAccountResponse(acc))
}

func (h *accountHandler) updateAccount(c *gin.Context) {
	n, ok := requireValidAccountNumber(c)
	if !ok {
		return
	}
	var req updateAccountRequest
	if !bindJSON(c, &req) {
		return
	}

	acc, err := h.accounts.Update(c.Request.Context(), middleware.CallerID(c), n, service.UpdateAccountInput{
		Name:        req.Name,
		AccountType: req.AccountType,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toAccountResponse(acc))
}

func (h *accountHandler) deleteAccount(c *gin.Context) {
	n, ok := requireValidAccountNumber(c)
	if !ok {
		return
	}
	if err := h.accounts.Delete(c.Request.Context(), middleware.CallerID(c), n); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
