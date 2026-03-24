package template

import "time"

type Template struct {
	Info        Info                   `yaml:"info" json:"info"`
	TemplateDir string                 `yaml:"template_dir" json:"template_dir"`
	Steps       []Step                 `yaml:"steps" json:"steps"`
	Hooks       HookConfig             `yaml:"hooks" json:"hooks"`
	GlobalVars  map[string]interface{} `yaml:"global_vars,omitempty" json:"global_vars,omitempty"`
}

type FileComplete struct {
	Filename string    `json:"filename"`
	Path     string    `json:"path"`
	Size     int64     `json:"size"`
	ModTime  time.Time `json:"mod_time"`
	Content  string    `json:"content"`
}

type Info struct {
	Name        string   `yaml:"name" json:"name"`
	Author      string   `yaml:"author" json:"author"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
	Category    string   `yaml:"category,omitempty" json:"category,omitempty"`
	System      bool     `yaml:"system" json:"system"`
	Tags        []string `yaml:"tags,omitempty" json:"tags,omitempty"`
}

type Step struct {
	ID              string                 `yaml:"id" json:"id"`
	Title           string                 `yaml:"title" json:"title"`
	Path            string                 `yaml:"path" json:"path"`
	Method          string                 `yaml:"method" json:"method"`
	TemplateFile    string                 `yaml:"template_file" json:"template_file"`
	SuccessMessage  string                 `yaml:"success_message,omitempty" json:"success_message,omitempty"`
	Next            string                 `yaml:"next,omitempty" json:"next,omitempty"`
	RedirectURL     string                 `yaml:"redirect_url,omitempty" json:"redirect_url,omitempty"`
	Capture         CaptureConfig          `yaml:"capture" json:"capture"`
	SimulateDelayMS int                    `yaml:"simulate_delay_ms,omitempty" json:"simulate_delay_ms,omitempty"`
	Vars            map[string]interface{} `yaml:"vars,omitempty" json:"vars,omitempty"`
}

type CaptureConfig struct {
	Enabled bool           `yaml:"enabled" json:"enabled"`
	Fields  []CaptureField `yaml:"fields,omitempty" json:"fields,omitempty"`
}

type CaptureField struct {
	Name          string `yaml:"name" json:"name"`
	Required      bool   `yaml:"required" json:"required"`
	ValidateRegex string `yaml:"validate_regex,omitempty" json:"validate_regex,omitempty"`
	ErrorMessage  string `yaml:"error_message,omitempty" json:"error_message,omitempty"`
}

type HookConfig struct {
	OnLoad []string `yaml:"onLoad,omitempty" json:"onLoad,omitempty"`
}

type HtmlFile struct {
	Filename string    `json:"filename"`
	Path     string    `json:"path"`
	Size     int64     `json:"size"`
	ModTime  time.Time `json:"mod_time"`
}

type TemplateMetadata struct {
	Content     string     `json:"content"`
	Filename    string     `json:"filename"`
	Name        string     `json:"name"`
	Author      string     `json:"author"`
	Description string     `json:"description,omitempty"`
	Category    string     `json:"category,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	TemplateDir string     `json:"template_dir"`
	Info        Info       `yaml:"info" json:"info"`
	Size        int64      `json:"size"`
	ModTime     time.Time  `json:"mod_time"`
	IsDir       bool       `json:"is_dir"`
	Mode        string     `json:"mode"`
	HtmlFiles   []HtmlFile `json:"html_files"`
}

type EmailTemplate struct {
	Id int64 `gorm:"primaryKey" json:"id"`

	UserId   *int64 `gorm:"index" json:"user_id,omitempty"`
	IsGlobal bool   `gorm:"default:false" json:"is_global"`
	Name     string `gorm:"not null" json:"name"`
	Category string `gorm:"default:'';index" json:"category,omitempty"`

	TrackOpens bool `gorm:"default:true" json:"track_opens"`

	Subject string `gorm:"not null" json:"subject"`
	Body    string `gorm:"type:text;not null" json:"body"`

	Attachments []EmailTemplateAttachment `gorm:"foreignKey:EmailTemplateId;constraint:OnDelete:CASCADE" json:"attachments,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type EmailTemplateAttachment struct {
	Id int64 `gorm:"primaryKey" json:"id"`

	EmailTemplateId int64  `gorm:"index;not null" json:"email_template_id"`
	Filename        string `gorm:"not null" json:"filename"`
	MimeType        string `gorm:"not null" json:"mime_type"`
	Size            int64  `gorm:"not null" json:"size"`
	Content         []byte `gorm:"type:blob;not null" json:"-"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
