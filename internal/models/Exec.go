package models

import (
	"database/sql"
)

type Exec struct {
	ID                int            `json:"id" db:"id"`
	FirstName         string         `json:"firstName" db:"firstName"`
	LastName          string         `json:"lastName" db:"lastName"`
	Email             string         `json:"email" db:"email"`
	Username          string         `json:"username" db:"username"`
	Password          string         `json:"password" db:"password"`
	PasswordChangedAt sql.NullString `json:"passwordChangedAt" db:"passwordChangedAt"`
	UserCreatedAt     sql.NullString `json:"userCreatedAt" db:"userCreatedAt"`
	CodeExpiresAt     sql.NullString `json:"tokenExpiresAt" db:"tokenExpiresAt"`
	ResetCode         sql.NullString `json:"resetCode" db:"passwordResetToken"`
	InactiveStatus    bool           `json:"inactiveStatus" db:"inactiveStatus"`
	Role              string         `json:"role" db:"role"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" db:"currentPassword"`
	NewPassword     string `json:"newPassword" db:"newPassword"`
}

type UpdatePasswordResponse struct {
	Token           string `json:"token" db:"token"`
	PasswordUpdated bool   `json:"passwordUpdated" db:"passwordUpdated"`
}
