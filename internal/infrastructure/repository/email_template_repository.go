package repository

import (
	"flexphish/internal/domain/template"

	"gorm.io/gorm"
)

type EmailTemplateRepository struct {
	db *gorm.DB
}

func NewEmailTemplateRepository(db *gorm.DB) template.EmailTemplateRepository {
	return &EmailTemplateRepository{db: db}
}

func (r *EmailTemplateRepository) Create(emailTemplate *template.EmailTemplate) error {
	return r.db.Create(emailTemplate).Error
}

func (r *EmailTemplateRepository) Update(emailTemplate *template.EmailTemplate) error {
	return r.db.Save(emailTemplate).Error
}

func (r *EmailTemplateRepository) Delete(id int64) error {
	return r.db.Delete(&template.EmailTemplate{}, id).Error
}

func (r *EmailTemplateRepository) GetByID(id int64) (*template.EmailTemplate, error) {
	var emailTemplate template.EmailTemplate
	err := r.db.First(&emailTemplate, id).Error
	return &emailTemplate, err
}

func (r *EmailTemplateRepository) GetAll(userID int64) ([]template.EmailTemplate, error) {
	var templates []template.EmailTemplate

	err := r.db.
		Where("is_global = ?", true).
		Or("user_id = ?", userID).
		Find(&templates).Error

	return templates, err
}

func (r *EmailTemplateRepository) ExistsByName(name string, userID int64, isGlobal bool, excludeID *int64) (bool, error) {
	var count int64

	query := r.db.Model(&template.EmailTemplate{}).Where("LOWER(name) = LOWER(?)", name)
	if excludeID != nil {
		query = query.Where("id <> ?", *excludeID)
	}

	if isGlobal {
		query = query.Where("is_global = ?", true)
	} else {
		query = query.Where("is_global = ? AND user_id = ?", false, userID)
	}

	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}
