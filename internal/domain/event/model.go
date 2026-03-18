package event

import "time"

type EventType string

const (
	EventVisit    EventType = "visit"
	EventPageView EventType = "page_view"
	EventSubmit   EventType = "submit"
	EventClick    EventType = "click"
	EventOpen     EventType = "open"
	EventRedirect EventType = "redirect"
	EventError    EventType = "error"
)

type Event struct {
	Id         int64  `gorm:"primaryKey" json:"id"`
	CampaignId int64  `gorm:"index;not null" json:"campaign_id"`
	ResultId   *int64 `gorm:"index" json:"result_id,omitempty"`

	Type      EventType `gorm:"type:text;index" json:"type"`
	StepID    string    `gorm:"index" json:"step_id"`
	Path      string    `json:"path"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Referrer  string    `json:"referrer"`

	Metadata  string    `gorm:"type:json" json:"metadata,omitempty"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}
