package handler

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

/*
respondBindError writes the response for a failed ShouldBindJSON call:
413 when the body exceeded the limit set by middleware.MaxBodySize,
400 with a client-friendly message for everything else.
*/
func respondBindError(c *gin.Context, err error) {
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body too large"})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": bindErrorMessage(err)})
}

/*
bindErrorMessage turns a ShouldBindJSON error into a client-friendly
message. Validation failures are reported per field using the
lowercased field name; anything else (malformed JSON, wrong types) is reported
as a generic invalid body, so Go struct names never reach the client.
*/
func bindErrorMessage(err error) string {
	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) {
		return "invalid request body"
	}

	messages := make([]string, 0, len(validationErrs))
	for _, fe := range validationErrs {
		field := strings.ToLower(fe.Field())

		switch fe.Tag() {
		case "required":
			messages = append(messages, fmt.Sprintf("%s is required", field))
		case "min":
			messages = append(messages, fmt.Sprintf("%s must be at least %s%s", field, fe.Param(), boundUnit(fe.Kind())))
		case "max":
			messages = append(messages, fmt.Sprintf("%s must be at most %s%s", field, fe.Param(), boundUnit(fe.Kind())))
		default:
			messages = append(messages, fmt.Sprintf("%s is invalid", field))
		}
	}

	return strings.Join(messages, "; ")
}

/*
boundUnit returns the unit to print after a min/max bound, based on
the kind of the field that failed: a length for strings, a count for
lists, and nothing for numbers, where the bound is the value itself.
*/
func boundUnit(kind reflect.Kind) string {
	switch kind {
	case reflect.String:
		return " characters"
	case reflect.Slice, reflect.Array, reflect.Map:
		return " items"
	default:
		return ""
	}
}
