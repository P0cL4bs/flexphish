package repository

import (
	"flexphish/internal/domain/result"

	"gorm.io/gorm"
)

type resultRepository struct {
	db *gorm.DB
}

func NewResultRepository(db *gorm.DB) result.Repository {
	return &resultRepository{db: db}
}

func (r *resultRepository) Create(res *result.Result) (*result.Result, error) {
	return res, r.db.Create(res).Error
}

func (r *resultRepository) Update(res *result.Result) error {
	return r.db.Save(res).Error
}

func (r *resultRepository) FindBySessionID(sessionID string) (*result.Result, error) {
	var res result.Result
	err := r.db.Where("session_id = ?", sessionID).First(&res).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &res, err
}

func (r *resultRepository) FindByCampaign(campaignId int64) ([]result.Result, error) {
	var list []result.Result
	err := r.db.Where("campaign_id = ?", campaignId).Find(&list).Error
	return list, err
}
