package dto

/*
RegisterRequest is the expected JSON body for the user registration
endpoint. It intentionally exposes only the fields a client is allowed
to set — system-generated fields like ID and CreatedAt live in
domain.User and are never part of this struct.
*/
type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

/*
LoginRequest is the expected JSON body for the login endpoint.
Password length is intentionally not validated here: format rules
belong to registration, while login only needs to check the value
against what was already stored.
*/
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

/*
RegisterResponse is the JSON shape returned after a successful
registration. No token is issued here — the client must log in to
obtain one. It deliberately omits the password field so a hashed
credential can never leak into an API response.
*/
type RegisterResponse struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

/*
AuthResponse is the JSON shape returned after a successful login.
It deliberately omits the password field so a hashed credential can
never leak into an API response.
*/
type AuthResponse struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Token string `json:"token"`
}
