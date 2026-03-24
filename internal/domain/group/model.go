package group

import (
	"flexphish/internal/domain/target"
	"time"
)

type Group struct {
	Id int64 `gorm:"primaryKey" json:"id"`

	UserId   *int64 `gorm:"index" json:"user_id,omitempty"`
	Name     string `gorm:"not null" json:"name"`
	IsGlobal bool   `gorm:"default:false" json:"is_global"`

	Targets []target.Target `gorm:"many2many:group_targets;" json:"targets,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
