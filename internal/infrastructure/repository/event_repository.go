package repository

import (
	"flexphish/internal/domain/event"

	"gorm.io/gorm"
)

type eventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) event.Repository {
	return &eventRepository{db: db}
}

func (e *eventRepository) Create(ev *event.Event) error {
	return e.db.Create(ev).Error
}

func (e *eventRepository) FindByCampaign(campaignId int64) ([]event.Event, error) {
	var list []event.Event
	err := e.db.Where("campaign_id = ?", campaignId).
		Order("created_at desc").
		Find(&list).Error
	return list, err
}

func (e *eventRepository) FindByResult(resultId int64) ([]event.Event, error) {
	var list []event.Event
	err := e.db.Where("result_id = ?", resultId).
		Order("created_at desc").
		Find(&list).Error
	return list, err
}
