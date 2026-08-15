package domain

import (
	"regexp"
	"strings"
	"time"
)

type User struct {
	ID        string
	Name      string
	Email     string
	Password  string
	CreatedAt time.Time
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func (u *User) IsValidEmail() bool {
	userEmail := strings.TrimSpace(u.Email)
	if userEmail == "" {
		return false
	}
	return emailRegex.MatchString(userEmail)
}
