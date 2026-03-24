package campaign

import (
	"flexphish/internal/domain/event"
	"flexphish/internal/domain/group"
	"flexphish/internal/domain/result"
	"flexphish/internal/domain/smtp"
	"flexphish/internal/domain/target"
	"flexphish/internal/domain/template"
	"time"
)

type CampaignStatus string
type EmailDispatchStatus string

const (
	StatusDraft     CampaignStatus = "draft"
	StatusScheduled CampaignStatus = "scheduled"
	StatusActive    CampaignStatus = "active"
	StatusStopped   CampaignStatus = "stopped"
	StatusCompleted CampaignStatus = "completed"
	StatusCancelled CampaignStatus = "cancelled"
)

const (
	EmailDispatchIdle       EmailDispatchStatus = "idle"
	EmailDispatchQueued     EmailDispatchStatus = "queued"
	EmailDispatchProcessing EmailDispatchStatus = "processing"
	EmailDispatchCompleted  EmailDispatchStatus = "completed"
	EmailDispatchFailed     EmailDispatchStatus = "failed"
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

	SendEmails      bool   `gorm:"default:false" json:"send_emails"`
	SMTPProfileId   *int64 `gorm:"index" json:"smtp_profile_id,omitempty"`
	EmailTemplateId *int64 `gorm:"index" json:"email_template_id,omitempty"`

	EmailDispatchStatus        EmailDispatchStatus `gorm:"type:text;default:'idle';index" json:"email_dispatch_status"`
	EmailDispatchQueuedAt      *time.Time          `json:"email_dispatch_queued_at,omitempty"`
	EmailDispatchStartedAt     *time.Time          `json:"email_dispatch_started_at,omitempty"`
	EmailDispatchCompletedAt   *time.Time          `json:"email_dispatch_completed_at,omitempty"`
	EmailDispatchLastAttemptAt *time.Time          `json:"email_dispatch_last_attempt_at,omitempty"`
	EmailDispatchLastError     string              `gorm:"type:text" json:"email_dispatch_last_error,omitempty"`

	EmailDispatchTotalTargets int64 `gorm:"default:0" json:"email_dispatch_total_targets"`
	EmailDispatchSent         int64 `gorm:"default:0" json:"email_dispatch_sent"`
	EmailDispatchFailed       int64 `gorm:"default:0" json:"email_dispatch_failed"`
	EmailDispatchPending      int64 `gorm:"default:0" json:"email_dispatch_pending"`

	SMTPProfile   *smtp.SMTPProfile       `gorm:"foreignKey:SMTPProfileId" json:"smtp_profile,omitempty"`
	EmailTemplate *template.EmailTemplate `gorm:"foreignKey:EmailTemplateId" json:"email_template,omitempty"`

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

	Groups []group.Group `gorm:"many2many:campaign_groups;" json:"groups,omitempty"`

	CampaignTargets []CampaignTarget `gorm:"foreignKey:CampaignId;constraint:OnDelete:CASCADE" json:"campaign_targets,omitempty"`

	Results []result.Result `gorm:"foreignKey:CampaignId;constraint:OnDelete:CASCADE" json:"results,omitempty"`
	Events  []event.Event   `gorm:"foreignKey:CampaignId;constraint:OnDelete:CASCADE" json:"events,omitempty"`
}

type CampaignTarget struct {
	Id int64 `gorm:"primaryKey" json:"id"`

	CampaignId int64  `gorm:"index;not null" json:"campaign_id"`
	TargetId   int64  `gorm:"index;not null" json:"target_id"`
	ResultId   *int64 `gorm:"index" json:"result_id,omitempty"`

	Target target.Target  `json:"target,omitempty"`
	Result *result.Result `json:"result,omitempty"`

	Token string `gorm:"uniqueIndex;size:64" json:"token"`

	Status string `gorm:"index" json:"status"`

	EmailSentAt *time.Time `json:"email_sent_at,omitempty"`
	OpenedAt    *time.Time `json:"opened_at,omitempty"`
	ClickedAt   *time.Time `json:"clicked_at,omitempty"`
	SubmittedAt *time.Time `json:"submitted_at,omitempty"`

	IP        string `json:"ip,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
