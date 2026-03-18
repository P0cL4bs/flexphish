package campaign

import (
	"flexphish/internal/domain/event"
	"flexphish/internal/domain/result"
	"time"
)

type CampaignStatus string

const (
	StatusDraft     CampaignStatus = "draft"
	StatusScheduled CampaignStatus = "scheduled"
	StatusActive    CampaignStatus = "active"
	StatusStopped   CampaignStatus = "stopped"
	StatusCompleted CampaignStatus = "completed"
	StatusCancelled CampaignStatus = "cancelled"
)

type Campaign struct {
	Id     int64 `gorm:"primaryKey" json:"id"`
	UserId int64 `gorm:"index;not null" json:"-"`

	Name string `gorm:"not null" json:"name"`

	Subdomain string `gorm:"uniqueIndex;not null" json:"subdomain"`

	Status CampaignStatus `gorm:"type:text;check:status IN ('draft','scheduled','active','stopped','completed','cancelled');default:'draft';index" json:"status"`

	LaunchDate    *time.Time `json:"launch_date,omitempty" json:"launch_date,omitempty"`
	CompletedDate *time.Time `json:"completed_date,omitempty" json:"completed_date,omitempty"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	TemplateId string `gorm:"not null;index" json:"template_id"`
	DevMode    bool   `gorm:"default:false;index" json:"dev_mode"`

	TrackOpens       bool `gorm:"default:true" json:"track_opens"`
	TrackClicks      bool `gorm:"default:true" json:"track_clicks"`
	TrackGeoLocation bool `gorm:"default:true" json:"track_geo_location"`
	TrackDeviceInfo  bool `gorm:"default:true" json:"track_device_info"`
	TrackIP          bool `gorm:"default:true" json:"track_ip"`
	TrackUserAgent   bool `gorm:"default:true" json:"track_user_agent"`
	TrackReferrer    bool `gorm:"default:true" json:"track_referrer"`

	EnableFingerprinting bool `gorm:"default:true" json:"enable_fingerprinting"`

	TotalSent      int64 `gorm:"default:0" json:"total_sent"`
	TotalOpened    int64 `gorm:"default:0" json:"total_opened"`
	TotalClicked   int64 `gorm:"default:0" json:"total_clicked"`
	TotalSubmitted int64 `gorm:"default:0" json:"total_submitted"`

	UniqueOpened    int64 `gorm:"default:0" json:"unique_opened"`
	UniqueClicked   int64 `gorm:"default:0" json:"unique_clicked"`
	UniqueSubmitted int64 `gorm:"default:0" json:"unique_submitted"`

	IsArchived bool       `gorm:"default:false;index" json:"is_archived"`
	DeletedAt  *time.Time `gorm:"index" json:"deleted_at"`

	Results []result.Result `gorm:"foreignKey:CampaignId;constraint:OnDelete:CASCADE" json:"results,omitempty"`
	Events  []event.Event   `gorm:"foreignKey:CampaignId;constraint:OnDelete:CASCADE" json:"events,omitempty"`
}
