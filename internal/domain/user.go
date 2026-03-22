package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (u *User) Validate() error {
	if strings.TrimSpace(u.Name) == "" {
		return errors.New("user name is required")
	}
	return nil
}
