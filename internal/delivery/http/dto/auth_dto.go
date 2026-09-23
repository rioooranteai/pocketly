package dto

/*
RegisterRequest is the expected JSON body for the user registration
endpoint. It intentionally exposes only the fields a client is allowed
to set — system-generated fields like ID and CreatedAt live in
domain.User and are never part of this struct.
Upper bounds keep oversized values out of the database and stop a
huge password from being fed into the (expensive) hasher.
*/
type RegisterRequest struct {
	Name     string `json:"name" binding:"required,max=100"`
	Email    string `json:"email" binding:"required,max=254"`
	Password string `json:"password" binding:"required,min=8,max=128"`
}

/*
LoginRequest is the expected JSON body for the login endpoint.
A minimum password length is intentionally not validated here: format
rules belong to registration, while login only needs to check the
value against what was already stored. The upper bounds match
RegisterRequest, so no stored account can be locked out by them.
*/
type LoginRequest struct {
	Email    string `json:"email" binding:"required,max=254"`
	Password string `json:"password" binding:"required,max=128"`
}

/*
AuthResponse is the JSON shape returned after a successful register
or login. It deliberately omits the password field so a hashed
credential can never leak into an API response.
*/
type AuthResponse struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Token string `json:"token"`
}
