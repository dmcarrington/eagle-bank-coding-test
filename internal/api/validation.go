package api

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" || name == "" {
				return fld.Name
			}
			return name
		})
	}
}

// bindJSON binds the request body and writes an error response on failure, returning false.
func bindJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			details := make([]validationDetail, 0, len(ve))
			for _, fe := range ve {
				details = append(details, validationDetail{
					Field:   fieldPath(fe),
					Message: humanMessage(fe),
					Type:    fe.Tag(),
				})
			}
			c.JSON(http.StatusBadRequest, badRequestResponse{
				Message: "validation failed",
				Details: details,
			})
			return false
		}
		c.JSON(http.StatusBadRequest, badRequestResponse{
			Message: "invalid request body",
			Details: []validationDetail{},
		})
		return false
	}
	return true
}

func fieldPath(fe validator.FieldError) string {
	ns := fe.Namespace()
	// Namespace is "StructName.field.subfield" — strip the leading struct name.
	if idx := strings.Index(ns, "."); idx != -1 {
		return ns[idx+1:]
	}
	return fe.Field()
}

func humanMessage(fe validator.FieldError) string {
	field := fieldPath(fe)
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be at least %s", field, fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be at most %s", field, fe.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, fe.Param())
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}
