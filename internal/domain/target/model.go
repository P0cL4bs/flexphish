package target

import "time"

type Target struct {
	Id int64 `gorm:"primaryKey" json:"id"`

	UserId int64 `gorm:"index;not null" json:"-"`

	FirstName string `gorm:"size:100" json:"first_name"`
	LastName  string `gorm:"size:100" json:"last_name"`
	Email     string `gorm:"index;not null" json:"email"`
	Position  string `gorm:"size:150" json:"position,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
