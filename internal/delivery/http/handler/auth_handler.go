package handler

import (
	"errors"
	"log"
	"net/http"
	"pocketly/internal/delivery/http/dto"
	"pocketly/internal/domain"

	"github.com/gin-gonic/gin"
)

/*
AuthHandler exposes authentication endpoints over HTTP. It only handles
request parsing, response formatting, and HTTP status codes — all
business logic lives behind AuthService.
*/
type AuthHandler struct {
	authService AuthService
}

/*
NewAuthHandler builds an AuthHandler backed by the given AuthService,
in practice usecase.AuthUsecase.
*/
func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

/*
Register handles POST requests to create a new user account.
It validates the incoming JSON body, delegates account creation to
AuthService, and maps domain errors to the appropriate HTTP status:
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

	token, userData, err := h.authService.Register(ctx, req.Name, req.Email, req.Password)
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
to AuthService, and responds 401 Unauthorized for any invalid email or
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

	token, userData, err := h.authService.Login(ctx, req.Email, req.Password)
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
