package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"pocketly/internal/delivery/http/dto"
	"pocketly/internal/domain"
	"pocketly/internal/usecase"
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
message. Validation failures are reported per field using the JSON
field name; anything else (malformed JSON, wrong types) is reported
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
			messages = append(messages, fmt.Sprintf("%s must be at least %s characters", field, fe.Param()))
		case "max":
			messages = append(messages, fmt.Sprintf("%s must be at most %s characters", field, fe.Param()))
		default:
			messages = append(messages, fmt.Sprintf("%s is invalid", field))
		}
	}

	return strings.Join(messages, "; ")
}

/*
AuthHandler exposes authentication endpoints over HTTP. It only handles
request parsing, response formatting, and HTTP status codes — all
business logic lives in AuthUsecase.
*/
type AuthHandler struct {
	AuthUsecase *usecase.AuthUsecase
}

/*
NewAuthHandler builds an AuthHandler backed by the given AuthUsecase.
*/
func NewAuthHandler(AuthUsecase *usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{
		AuthUsecase: AuthUsecase,
	}
}

/*
Register handles POST requests to create a new user account.
It validates the incoming JSON body, delegates account creation to
AuthUsecase, and maps domain errors to the appropriate HTTP status:
409 for a duplicate email, 400 for an invalid email or name, and 500
for any unexpected failure. On success it responds 201 Created with
a signed JWT and the new user's public data, so the client is signed
in right away without a separate login call.
*/
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest

	err := c.ShouldBindJSON(&req)
	if err != nil {
		respondBindError(c, err)
		return
	}

	ctx := c.Request.Context()

	token, userData, err := h.AuthUsecase.Register(ctx, req.Name, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}

		if errors.Is(err, domain.ErrInvalidEmail) || errors.Is(err, domain.ErrInvalidName) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Printf("register failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	response := dto.AuthResponse{
		Name:  userData.Name,
		Email: userData.Email,
		Token: token,
	}

	c.JSON(http.StatusCreated, response)
}

/*
Login handles POST requests to authenticate an existing user.
It validates the incoming JSON body, delegates credential verification
to AuthUsecase, and responds 401 Unauthorized for any invalid email or
password combination, or 500 for any unexpected failure. On success it
responds 200 OK with a signed JWT and the user's public data.
*/
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

	err := c.ShouldBindJSON(&req)
	if err != nil {
		respondBindError(c, err)
		return
	}

	ctx := c.Request.Context()

	token, userData, err := h.AuthUsecase.Login(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		log.Printf("login failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	response := dto.AuthResponse{
		Name:  userData.Name,
		Email: userData.Email,
		Token: token,
	}

	c.JSON(http.StatusOK, response)
}
