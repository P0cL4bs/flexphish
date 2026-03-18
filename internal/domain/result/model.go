package result

import (
	"flexphish/internal/domain/event"
	"time"
)

type ResultStatus string

const (
	ResultInProgress ResultStatus = "in_progress"
	ResultCompleted  ResultStatus = "completed"
	ResultAbandoned  ResultStatus = "abandoned"
)

type Result struct {
	Id         int64 `gorm:"primaryKey" json:"id"`
	CampaignId int64 `gorm:"index;not null" json:"campaign_id"`

	SessionID string `gorm:"uniqueIndex;not null" json:"session_id"`

	Email    string `gorm:"index" json:"email,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`

	IP        string `json:"ip,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
	Country   string `json:"country,omitempty"`
	City      string `json:"city,omitempty"`

	Device  string `json:"device,omitempty"`
	OS      string `json:"os,omitempty"`
	Browser string `json:"browser,omitempty"`

	Status ResultStatus `gorm:"type:text;default:'in_progress'" json:"status"`

	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`

	Events []event.Event `gorm:"foreignKey:ResultId;constraint:OnDelete:CASCADE" json:"events,omitempty"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at,omitempty"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at,omitempty"`
}
