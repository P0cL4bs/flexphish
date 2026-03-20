package repository

import (
	"errors"
	"flexphish/internal/domain/campaign"
	"flexphish/internal/domain/event"
	"flexphish/internal/domain/result"
	"flexphish/internal/domain/template"
	"time"

	"gorm.io/gorm"
)

type CampaignRepository struct {
	db *gorm.DB
}

func NewCampaignRepository(db *gorm.DB, trepo template.TemplateRepository) campaign.Repository {
	db.Exec("PRAGMA foreign_keys = ON")
	return &CampaignRepository{db: db}

}

func (r *CampaignRepository) Create(c *campaign.Campaign) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		groups := c.Groups
		c.Groups = nil

		if err := tx.Create(c).Error; err != nil {
			return err
		}

		if err := tx.Model(c).Association("Groups").Replace(groups); err != nil {
			return err
		}

		c.Groups = groups
		return nil
	})
}

func (r *CampaignRepository) GetByID(id int64, userId int64) (*campaign.Campaign, error) {
	var c campaign.Campaign

	err := r.db.
		Preload("Results").
		Preload("Events").
		Preload("Groups").
		Preload("SMTPProfile").
		Preload("EmailTemplate").
		Where("id = ? AND user_id = ?", id, userId).
		First(&c).Error

	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (r *CampaignRepository) GetByURL(url string) (*campaign.Campaign, error) {
	var c campaign.Campaign

	err := r.db.
		Where("url = ?", url).
		First(&c).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (r *CampaignRepository) ListByUser(
	userId int64,
	status *campaign.CampaignStatus,
	page, pageSize int,
) ([]campaign.Campaign, int64, error) {

	var campaigns []campaign.Campaign
	var total int64

	baseQuery := r.db.Model(&campaign.Campaign{}).
		Where("user_id = ?", userId)

	if status != nil {
		baseQuery = baseQuery.Where("status = ?", *status)
	}

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	err := baseQuery.
		Preload("Groups").
		Preload("SMTPProfile").
		Preload("EmailTemplate").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&campaigns).Error

	if err != nil {
		return nil, 0, err
	}

	return campaigns, total, nil
}

func (r *CampaignRepository) Update(c *campaign.Campaign) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Model(&campaign.Campaign{}).
			Where("id = ? AND user_id = ?", c.Id, c.UserId).
			Updates(map[string]interface{}{
				"name":              c.Name,
				"subdomain":         c.Subdomain,
				"launch_date":       c.LaunchDate,
				"template_id":       c.TemplateId,
				"completed_date":    c.CompletedDate,
				"status":            c.Status,
				"dev_mode":          c.DevMode,
				"send_emails":       c.SendEmails,
				"smtp_profile_id":   c.SMTPProfileId,
				"email_template_id": c.EmailTemplateId,
			}).Error; err != nil {
			return err
		}

		return tx.Model(c).Association("Groups").Replace(c.Groups)
	})
}

func (r *CampaignRepository) Delete(id int64, userId int64) error {
	return r.db.
		Where("id = ? AND user_id = ?", id, userId).
		Delete(&campaign.Campaign{}).Error
}

func (r *CampaignRepository) FindActiveBySubdomain(subdomain string) (*campaign.Campaign, error) {
	var camp campaign.Campaign

	err := r.db.
		Where("subdomain = ?", subdomain).
		Where("(status = ? OR dev_mode = ?)", "active", true).
		First(&camp).Error

	if err != nil {
		return nil, err
	}

	return &camp, nil
}

func (r *CampaignRepository) FindBySubdomain(subdomain string) (*campaign.Campaign, error) {
	var camp campaign.Campaign

	err := r.db.
		Where("subdomain = ?", subdomain).
		First(&camp).Error

	if err != nil {
		return nil, err
	}

	return &camp, nil
}

func (r *CampaignRepository) IncrementClicked(id int64) error {
	return r.db.Model(&campaign.Campaign{}).
		Where("id = ?", id).
		UpdateColumn("total_clicked", gorm.Expr("total_clicked + ?", 1)).
		Error
}

func (r *CampaignRepository) IncrementSubmitted(id int64) error {
	return r.db.Model(&campaign.Campaign{}).
		Where("id = ?", id).
		UpdateColumn("total_submitted", gorm.Expr("total_submitted + ?", 1)).
		Error
}

func (r *CampaignRepository) IncrementOpened(id int64) error {
	return r.db.Model(&campaign.Campaign{}).
		Where("id = ?", id).
		UpdateColumn("total_opened", gorm.Expr("total_opened + ?", 1)).
		Error
}

func (r *CampaignRepository) UpdateStatus(id int64, status campaign.CampaignStatus) error {

	return r.db.Model(&campaign.Campaign{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      status,
			"launch_date": time.Now(),
		}).Error
}

func (r *CampaignRepository) CountCampaignsUsingTemplateId(templateId string) (int64, error) {
	var count int64

	err := r.db.
		Model(&campaign.Campaign{}).
		Where("template_id = ?", templateId).
		Count(&count).
		Error

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *CampaignRepository) HasActiveCampaignUsingTemplate(templateId string) (bool, error) {
	var count int64

	err := r.db.
		Model(&campaign.Campaign{}).
		Where("template_id = ? AND status = ?", templateId, campaign.StatusActive).
		Count(&count).
		Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *CampaignRepository) GetTopCampaigns(userID int64) ([]campaign.TopCampaignMetric, error) {

	var result []campaign.TopCampaignMetric

	err := r.db.Model(&campaign.Campaign{}).
		Select(`
			id as campaign_id,
			name,
			total_clicked as clicked,
			total_submitted as submitted,
			CASE 
				WHEN total_clicked > 0 
				THEN (total_submitted * 100.0 / total_clicked)
				ELSE 0 
			END as conversion_rate
		`).
		Where("user_id = ? AND total_clicked > 0", userID).
		Order("conversion_rate DESC").
		Limit(10).
		Scan(&result).Error

	return result, err
}

func periodExpression(period string) string {

	switch period {

	case "day":
		return "DATE(events.created_at)"

	case "week":
		return "DATE(events.created_at, 'weekday 1', '-7 days')"

	case "month":
		return "strftime('%Y-%m-01', events.created_at)"

	case "year":
		return "strftime('%Y-01-01', events.created_at)"

	default:
		return "DATE(events.created_at)"
	}

}

func periodFilter(period string) (string, []interface{}) {

	switch period {

	case "day":
		return "date(events.created_at) = date('now')", nil

	case "week":
		return "date(events.created_at) >= date('now', '-7 days')", nil

	case "month":
		return "date(events.created_at) >= date('now', '-1 month')", nil

	case "year":
		return "date(events.created_at) >= date('now', '-1 year')", nil

	default:
		return "", nil
	}
}

func (r *CampaignRepository) GetTimeline(userID int64, period string) ([]campaign.TimelineMetric, error) {

	result := make([]campaign.TimelineMetric, 0)

	group := periodExpression(period)
	filter, args := periodFilter(period)

	query := r.db.Table("events").
		Select(`
			events.campaign_id,
			campaigns.name as campaign_name,
			`+group+` as period,
			count(*) as count
		`).
		Joins("JOIN campaigns ON campaigns.id = events.campaign_id").
		Where("campaigns.user_id = ?", userID)

	if filter != "" {
		query = query.Where(filter, args...)
	}

	err := query.
		Group("events.campaign_id, campaigns.name, period").
		Order("period ASC").
		Scan(&result).Error

	return result, err
}

func (r *CampaignRepository) GetAnalytics(userID int64, period string) (*campaign.CampaignAnalytics, error) {

	var analytics campaign.CampaignAnalytics

	r.db.Model(&campaign.Campaign{}).
		Where("user_id = ?", userID).
		Count(&analytics.TotalCampaigns)

	r.db.Model(&campaign.Campaign{}).
		Where("user_id = ? AND status = ?", userID, campaign.StatusActive).
		Count(&analytics.ActiveCampaigns)

	r.db.Model(&event.Event{}).
		Joins("JOIN campaigns ON campaigns.id = events.campaign_id").
		Where("campaigns.user_id = ?", userID).
		Count(&analytics.EventsCaptured)

	r.db.Model(&event.Event{}).
		Joins("JOIN campaigns ON campaigns.id = events.campaign_id").
		Where("campaigns.user_id = ? AND events.type = ?", userID, event.EventSubmit).
		Count(&analytics.CredentialsCaptured)

	// event types
	r.db.Model(&event.Event{}).
		Select("type, count(*) as count").
		Joins("JOIN campaigns ON campaigns.id = events.campaign_id").
		Where("campaigns.user_id = ?", userID).
		Group("type").
		Scan(&analytics.EventTypes)

	// timeline
	timeline, _ := r.GetTimeline(userID, period)
	analytics.Timeline = timeline

	// top campaigns
	top, _ := r.GetTopCampaigns(userID)
	analytics.TopCampaigns = top

	return &analytics, nil
}

func (r *CampaignRepository) DeleteResult(resultID int64, campaignID int64) error {

	return r.db.
		Where("id = ? AND campaign_id = ?", resultID, campaignID).
		Delete(&result.Result{}).
		Error
}
