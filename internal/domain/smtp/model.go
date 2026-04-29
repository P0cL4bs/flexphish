package smtp

import "time"

const (
	SecurityModeStartTLS    = "starttls"
	SecurityModeImplicitTLS = "implicit_tls"
	SecurityModeNone        = "none"
)

type SMTPProfile struct {
	Id int64 `gorm:"primaryKey" json:"id"`

	UserId   *int64 `gorm:"index" json:"user_id,omitempty"`
	IsGlobal bool   `gorm:"default:false" json:"is_global"`

	Name string `gorm:"not null" json:"name"`

	Host         string `gorm:"not null" json:"host"`
	Port         int    `gorm:"not null" json:"port"`
	SecurityMode string `gorm:"not null;default:starttls" json:"security_mode"`

	Username string `gorm:"not null" json:"username"`

	Password string `gorm:"not null" json:"-"`

	FromName  string `json:"from_name,omitempty"`
	FromEmail string `json:"from_email,omitempty"`

	IsActive bool `gorm:"default:true" json:"is_active"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
