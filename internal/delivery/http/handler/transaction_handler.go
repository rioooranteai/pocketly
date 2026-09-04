package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"pocketly/internal/delivery/http/dto"
	"pocketly/internal/domain"
	"pocketly/internal/usecase"
)

/*
TransactionHandler exposes financial transaction management endpoints over HTTP.
It only handles request parsing, response formatting, and HTTP status codes — all
business logic lives in TransactionUsecase.
*/
type TransactionHandler struct {
	transactionUsecase *usecase.TransactionUsecase
}

/*
NewTransactionHandler builds a TransactionHandler backed by the given TransactionUsecase.
*/
func NewTransactionHandler(transactionUsecase *usecase.TransactionUsecase) *TransactionHandler {
	return &TransactionHandler{transactionUsecase: transactionUsecase}
}

/*
Create handles POST requests to record a new financial transaction.
It validates the incoming JSON body, converts DTO items to domain items,
delegates creation to TransactionUsecase, and maps domain errors to the
appropriate HTTP status codes. On success it responds 201 Created with the
newly created transaction data.
*/
func (h *TransactionHandler) Create(c *gin.Context) {
	var req dto.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var domainItems []domain.TransactionItem
	for _, item := range req.Items {
		domainItems = append(domainItems, domain.TransactionItem{
			Name:     item.Name,
			Quantity: item.Quantity,
			Price:    item.Price,
		})
	}

	ctx := c.Request.Context()

	// Sesuai signature Usecase: (ctx, userID, descriptions, items)
	transaction, err := h.transactionUsecase.CreateTransaction(ctx, userID, req.Description, domainItems)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create transaction"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Transaction created successfully",
		"data":    toTransactionResponse(transaction),
	})
}

/*
Get handles GET requests to retrieve a single transaction by its ID.
It extracts the ID from path parameters, delegates retrieval to
TransactionUsecase, and maps domain errors like ErrTransactionNotFound to 404
Not Found. On success it responds 200 OK with the requested transaction.
*/
func (h *TransactionHandler) Get(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id := c.Param("id")
	ctx := c.Request.Context()

	// Sesuai signature Usecase: (ctx, userID, transactionID)
	transaction, err := h.transactionUsecase.GetTransaction(ctx, userID, id)
	if err != nil {
		if errors.Is(err, domain.ErrTransactionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": toTransactionResponse(transaction),
	})
}

/*
List handles GET requests to fetch all transactions belonging to the authenticated user.
It delegates retrieval to TransactionUsecase and maps internal errors to 500 Internal
Server Error. On success it responds 200 OK with a list of user transactions.
*/
func (h *TransactionHandler) List(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	ctx := c.Request.Context()

	// Sesuai nama method di Usecase: ListMyTransaction (singular)
	transaction, err := h.transactionUsecase.ListMyTransaction(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	// Jika usecase mengembalikan single pointer (*domain.Transaction)
	var responseData []dto.TransactionResponse
	if transaction != nil {
		responseData = append(responseData, toTransactionResponse(transaction))
	} else {
		responseData = []dto.TransactionResponse{}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": responseData,
	})
}

/*
Update handles PUT requests to modify an existing transaction.
It validates the incoming JSON body, converts DTO items to domain items,
delegates update execution to TransactionUsecase, and maps ErrTransactionNotFound
to 404 Not Found. On success it responds 200 OK with the updated transaction details.
*/
func (h *TransactionHandler) Update(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req dto.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	var domainItems []domain.TransactionItem
	for _, item := range req.Items {
		domainItems = append(domainItems, domain.TransactionItem{
			Name:     item.Name,
			Quantity: item.Quantity,
			Price:    item.Price,
		})
	}

	ctx := c.Request.Context()

	// Sesuai signature Usecase: (ctx, userID, descriptions, items)
	transaction, err := h.transactionUsecase.UpdateTransaction(ctx, userID, req.Description, domainItems)
	if err != nil {
		if errors.Is(err, domain.ErrTransactionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Transaction updated successfully",
		"data":    toTransactionResponse(transaction),
	})
}

/*
Delete handles DELETE requests to remove a transaction by its ID.
It extracts the transaction ID from path parameters, delegates deletion to
TransactionUsecase, and maps ErrTransactionNotFound to 404 Not Found. On success
it responds 204 No Content with an empty response body.
*/
func (h *TransactionHandler) Delete(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id := c.Param("id")
	ctx := c.Request.Context()

	// Sesuai signature Usecase: (ctx, userID, transactionID)
	err := h.transactionUsecase.DeleteTransaction(ctx, userID, id)
	if err != nil {
		if errors.Is(err, domain.ErrTransactionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete transaction"})
		return
	}

	c.Status(http.StatusNoContent)
}

// --- Helper Conversion ---

func toTransactionResponse(t *domain.Transaction) dto.TransactionResponse {
	if t == nil {
		return dto.TransactionResponse{}
	}

	var itemResponses []dto.TransactionItemResponse
	for _, item := range t.Items {
		itemResponses = append(itemResponses, dto.TransactionItemResponse{
			Name:     item.Name,
			Quantity: item.Quantity,
			Price:    item.Price,
		})
	}

	if itemResponses == nil {
		itemResponses = []dto.TransactionItemResponse{}
	}

	return dto.TransactionResponse{
		ID:          t.ID,
		Description: t.Description,
		Category:    t.Category,
		TotalAmount: t.TotalAmount,
		Date:        t.Date,
		Items:       itemResponses,
	}
}
