package domain

import (
	"regexp"
	"strings"
	"time"
)

/*
User represents an application user. This struct is pure business
data — it must never import database, HTTP, or any other framework
specific packages.
*/
type User struct {
	ID        string
	Name      string
	Email     string
	Password  string
	CreatedAt time.Time
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

/*
NormalizeEmail trims surrounding whitespace and lowercases the email,
so "User@Mail.com" and " user@mail.com" resolve to the same account.
It must be applied everywhere an email is stored or looked up.
*/
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

/*
IsValidEmail reports whether the user's email matches a standard
email format. This is a format check only — it does not verify that
the address actually exists or is reachable.
*/
func (u User) IsValidEmail() bool {
	userEmail := strings.TrimSpace(u.Email)
	if userEmail == "" {
		return false
	}
	return emailRegex.MatchString(userEmail)
}
