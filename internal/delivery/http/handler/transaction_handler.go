package handler

import (
	"errors"
	"io"
	"log"
	"net/http"
	"time"

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

	// Sesuai signature Usecase: (ctx, userID, description, items, date)
	transaction, err := h.transactionUsecase.CreateTransaction(ctx, userID, req.Description, domainItems, req.Date)
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
TransactionUsecase, and responds 404 Not Found when the transaction does
not exist or belongs to another user (see isTransactionNotFound). On
success it responds 200 OK with the requested transaction.
*/
func (h *TransactionHandler) Get(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id := c.Param("id")
	ctx := c.Request.Context()

	transaction, err := h.transactionUsecase.GetTransaction(ctx, userID, id)
	if err != nil {
		if isTransactionNotFound(err) {
			logUnauthorizedAccess(c, err, userID, id)
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

	transactions, err := h.transactionUsecase.ListMyTransactions(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	responseData := make([]dto.TransactionResponse, 0, len(transactions))
	for _, t := range transactions {
		responseData = append(responseData, toTransactionResponse(&t))
	}

	c.JSON(http.StatusOK, gin.H{
		"data": responseData,
	})
}

/*
Update handles PUT requests to modify an existing transaction.
It validates the incoming JSON body, converts DTO items to domain items,
delegates update execution to TransactionUsecase, and responds 404 Not Found when
the transaction does not exist or belongs to another user (see
isTransactionNotFound). On success it responds 200 OK with the updated transaction details.
*/
func (h *TransactionHandler) Update(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	id := c.Param("id")

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

	// Sesuai signature Usecase: (ctx, userID, transactionID, description, items, date)
	transaction, err := h.transactionUsecase.UpdateTransaction(ctx, userID, id, req.Description, domainItems, req.Date)
	if err != nil {
		if isTransactionNotFound(err) {
			logUnauthorizedAccess(c, err, userID, id)
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
TransactionUsecase, and responds 404 Not Found when the transaction does not
exist or belongs to another user (see isTransactionNotFound). On success
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

	err := h.transactionUsecase.DeleteTransaction(ctx, userID, id)
	if err != nil {
		if isTransactionNotFound(err) {
			logUnauthorizedAccess(c, err, userID, id)
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete transaction"})
		return
	}

	c.Status(http.StatusNoContent)
}

/*
Scan handles POST requests to record a new transaction from a receipt
image upload. It reads the uploaded file into memory, delegates
extraction and creation to TransactionUsecase, and responds the same
way Create does. Date defaults to the current time since a scanned
receipt has no explicit date field in the request. The route's
MaxBodySize middleware bounds how much is read; a body past that limit
is answered with 413 Request Entity Too Large.
*/
func (h *TransactionHandler) Scan(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	fileHeader, err := c.FormFile("receipt")
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "receipt image too large"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid file field 'receipt'"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
		return
	}
	defer file.Close()

	imageData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read uploaded file"})
		return
	}

	ctx := c.Request.Context()

	transaction, err := h.transactionUsecase.CreateTransactionFromImage(ctx, userID, imageData, time.Now())
	if err != nil {
		if errors.Is(err, domain.ErrEmptyImageData) || errors.Is(err, domain.ErrImageSizeExceedsLimit) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process receipt image"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Transaction created successfully from receipt",
		"data":    toTransactionResponse(transaction),
	})
}

/*
isTransactionNotFound reports whether err should be answered with
404 Not Found. A transaction owned by someone else is reported the
same way as one that does not exist, so a client cannot probe IDs to
learn which transactions exist. The usecase still returns the two
errors separately; hiding the difference is an HTTP concern.
*/
func isTransactionNotFound(err error) bool {
	return errors.Is(err, domain.ErrTransactionNotFound) || errors.Is(err, domain.ErrUnauthorizedAccess)
}

/*
logUnauthorizedAccess records an attempt to reach a transaction owned
by another user. The client only ever sees 404, so this log line is
the one place those attempts stay visible, e.g. to spot a user probing
IDs. Plain not-found errors are not logged, since mistyped or stale
IDs are routine.
*/
func logUnauthorizedAccess(c *gin.Context, err error, userID, transactionID string) {
	if !errors.Is(err, domain.ErrUnauthorizedAccess) {
		return
	}
	log.Printf("unauthorized transaction access: user=%s transaction=%s method=%s ip=%s",
		userID, transactionID, c.Request.Method, c.ClientIP())
}

// --- Helper Conversion ---

func toTransactionResponse(t *domain.Transaction) dto.TransactionResponse {
	if t == nil {
		return dto.TransactionResponse{}
	}

	itemResponses := make([]dto.TransactionItemResponse, 0, len(t.Items))
	for _, item := range t.Items {
		itemResponses = append(itemResponses, dto.TransactionItemResponse{
			Name:     item.Name,
			Quantity: item.Quantity,
			Price:    item.Price,
		})
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
