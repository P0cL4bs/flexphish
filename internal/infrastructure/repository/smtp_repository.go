package repository

import (
	"flexphish/internal/domain/smtp"

	"gorm.io/gorm"
)

type SMTPRepository struct {
	db *gorm.DB
}

func NewSMTPRepository(db *gorm.DB) smtp.Repository {
	return &SMTPRepository{db: db}
}

func (r *SMTPRepository) Create(profile *smtp.SMTPProfile) error {
	return r.db.Create(profile).Error
}

func (r *SMTPRepository) Update(profile *smtp.SMTPProfile) error {
	return r.db.Save(profile).Error
}

func (r *SMTPRepository) Delete(id int64) error {
	return r.db.Delete(&smtp.SMTPProfile{}, id).Error
}

func (r *SMTPRepository) GetByID(id int64) (*smtp.SMTPProfile, error) {
	var profile smtp.SMTPProfile
	err := r.db.First(&profile, id).Error
	return &profile, err
}

func (r *SMTPRepository) GetAll(userID int64) ([]smtp.SMTPProfile, error) {
	var profiles []smtp.SMTPProfile

	err := r.db.
		Where("is_global = ?", true).
		Or("user_id = ?", userID).
		Find(&profiles).Error

	return profiles, err
}

func (r *SMTPRepository) ExistsByConnection(host string, port int, username string, userID int64, isGlobal bool, excludeID *int64) (bool, error) {
	var count int64

	query := r.db.Model(&smtp.SMTPProfile{}).
		Where("LOWER(host) = LOWER(?)", host).
		Where("port = ?", port).
		Where("LOWER(username) = LOWER(?)", username)

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
