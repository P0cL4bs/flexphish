package smtp

import "time"

type SMTPProfile struct {
	Id int64 `gorm:"primaryKey" json:"id"`

	UserId int64 `gorm:"index;not null" json:"-"`

	Name string `gorm:"not null" json:"name"`

	Host string `gorm:"not null" json:"host"`
	Port int    `gorm:"not null" json:"port"`

	Username string `gorm:"not null" json:"username"`

	Password string `gorm:"not null" json:"-"`

	FromName  string `json:"from_name,omitempty"`
	FromEmail string `json:"from_email,omitempty"`

	IsActive bool `gorm:"default:true" json:"is_active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
