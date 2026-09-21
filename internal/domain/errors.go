package domain

import "errors"

/*
Domain-level errors returned by usecases. Handlers map these to
specific HTTP status codes rather than exposing raw internal errors
to API clients.
*/
var (
	/*
		ErrEmailAlreadyExists is returned when registering with an email
		that is already associated with an existing account.
	*/
	ErrEmailAlreadyExists = errors.New("email already registered")

	/*
		ErrInvalidCredentials is returned when login fails, whether due
		to a non-existent email or a wrong password. Both cases share
		this single error to prevent user enumeration.
	*/
	ErrInvalidCredentials = errors.New("invalid email or password")

	/*
		ErrInvalidEmail is returned when a submitted email fails format
		validation.
	*/
	ErrInvalidEmail = errors.New("invalid email format")

	/*
		ErrTransactionNotFound is returned when a requested transaction
		does not exist.
	*/
	ErrTransactionNotFound = errors.New("transaction not found")

	/*
		ErrUnauthorizedAccess is returned when a user attempts to access
		or modify a resource — such as a transaction — that belongs to
		a different user.
	*/
	ErrUnauthorizedAccess = errors.New("you don't have access to this resource")

	/*
		ErrEmptyImageData is returned when no image bytes were provided to
		the vision extractor.
	*/
	ErrEmptyImageData = errors.New("image data is empty")

	/*
	   ErrImageSizeExceedsLimit is returned when the uploaded image exceeds
	   the configured maximum file size.
	*/
	ErrImageSizeExceedsLimit = errors.New("image size exceeds maximum allowed limit")
)
