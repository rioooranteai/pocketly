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
)
