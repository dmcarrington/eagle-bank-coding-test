package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/davidcarrington/eagle-bank/internal/service"
)

type errorResponse struct {
	Message string `json:"message"`
}

type validationDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

type badRequestResponse struct {
	Message string             `json:"message"`
	Details []validationDetail `json:"details"`
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		c.JSON(http.StatusNotFound, errorResponse{Message: err.Error()})
	case errors.Is(err, service.ErrForbidden):
		c.JSON(http.StatusForbidden, errorResponse{Message: err.Error()})
	case errors.Is(err, service.ErrConflict):
		c.JSON(http.StatusConflict, errorResponse{Message: err.Error()})
	case errors.Is(err, service.ErrEmailTaken):
		c.JSON(http.StatusBadRequest, badRequestResponse{
			Message: "validation failed",
			Details: []validationDetail{{
				Field:   "email",
				Message: err.Error(),
				Type:    "unique",
			}},
		})
	case errors.Is(err, service.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, errorResponse{Message: "invalid email or password"})
	case errors.Is(err, service.ErrInsufficientFunds):
		c.JSON(http.StatusUnprocessableEntity, errorResponse{Message: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, errorResponse{Message: "an unexpected error occurred"})
	}
}
