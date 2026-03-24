package template

type TemplateRepository interface {
	LoadAll() (map[string]*TemplateMetadata, error)
	LoadByFilename(filename string) (*TemplateMetadata, error)
	GetTemplateByFilename(tempalte_id string) (*Template, error)
	Exists(filename string) (bool, error)
	Save(filename, content string) error
	CreateTemplateDir(templateDir string) error
	CopyTemplateDir(srcTemplateDir, dstTemplateDir string) error
	TemplateDirExists(templateDir string) (bool, error)
	Delete(filename string) error
	DeleteTemplateDir(templateDir string) error
}

type HtmlfilesRepository interface {
	List(templateDir string) ([]FileComplete, error)
	Get(templateDir, filename string) (*FileComplete, error)
	Create(templateDir, filename, content string) error
	Update(templateDir, filename, content string) error
	Delete(templateDir, filename string) error
	Exists(templateDir, filename string) (bool, error)
	GetAllByTemplateFilename(templateFilename string) ([]FileComplete, error)
	CreateByTemplateFilename(templateFilename, filename, content string) (*FileComplete, error)
	UpdateByTemplateFilename(templateFilename, filename, content string) (*FileComplete, error)
	DeleteByTemplateFilename(templateFilename, filename string) (*FileComplete, error)
}

type StaticFileRepository interface {
	List(templateDir string) ([]FileComplete, error)
	Get(templateDir, filename string) (*FileComplete, error)
	Create(templateDir, filename string, content []byte) error
	Update(templateDir, filename, content string) error
	Delete(templateDir, filename string) error
	Exists(templateDir, filename string) (bool, error)
	GetAllByTemplateFilename(templateFilename string) ([]FileComplete, error)
	CreateByTemplateFilename(templateFilename, filename string, content []byte) (*FileComplete, error)
	UpdateByTemplateFilename(templateFilename, filename, content string) (*FileComplete, error)
	DeleteByTemplateFilename(templateFilename, filename string) (*FileComplete, error)
}

type EmailTemplateRepository interface {
	Create(emailTemplate *EmailTemplate) error
	Update(emailTemplate *EmailTemplate) error
	Delete(id int64) error

	GetByID(id int64) (*EmailTemplate, error)
	GetAll(userID int64) ([]EmailTemplate, error)
	ExistsByName(name string, userID int64, isGlobal bool, excludeID *int64) (bool, error)

	CreateAttachment(attachment *EmailTemplateAttachment) error
	GetAttachments(emailTemplateID int64) ([]EmailTemplateAttachment, error)
	DeleteAttachment(emailTemplateID int64, attachmentID int64) error
}
