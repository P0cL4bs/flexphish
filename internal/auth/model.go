package auth

import (
	"time"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type User struct {
	ID           int64  `gorm:"primaryKey;autoIncrement"`
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	Role         Role   `gorm:"type:text;not null;default:user"`
	IsActive     bool   `gorm:"default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
