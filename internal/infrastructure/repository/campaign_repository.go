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
		Preload("Groups.Targets").
		Preload("CampaignTargets").
		Preload("CampaignTargets.Target").
		Preload("CampaignTargets.Result").
		Preload("SMTPProfile").
		Preload("EmailTemplate").
		Where("id = ? AND user_id = ?", id, userId).
		First(&c).Error

	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (r *CampaignRepository) ListScheduledStartCandidates(now time.Time) ([]campaign.Campaign, error) {
	var campaigns []campaign.Campaign

	err := r.db.
		Where("status = ? AND launch_date IS NOT NULL AND launch_date <= ? AND is_archived = ? AND deleted_at IS NULL",
			campaign.StatusScheduled, now, false).
		Order("launch_date ASC").
		Order("id ASC").
		Find(&campaigns).Error
	if err != nil {
		return nil, err
	}

	return campaigns, nil
}

func (r *CampaignRepository) ListEmailDispatchCandidates() ([]campaign.Campaign, error) {
	var campaigns []campaign.Campaign

	err := r.db.
		Preload("Groups").
		Where("status = ? AND send_emails = ? AND launch_date IS NOT NULL AND is_archived = ? AND deleted_at IS NULL",
			campaign.StatusActive, true, false).
		Order("launch_date ASC").
		Order("id ASC").
		Find(&campaigns).Error
	if err != nil {
		return nil, err
	}

	return campaigns, nil
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
		Preload("CampaignTargets").
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
				"name":                           c.Name,
				"subdomain":                      c.Subdomain,
				"launch_date":                    c.LaunchDate,
				"template_id":                    c.TemplateId,
				"completed_date":                 c.CompletedDate,
				"status":                         c.Status,
				"dev_mode":                       c.DevMode,
				"send_emails":                    c.SendEmails,
				"smtp_profile_id":                c.SMTPProfileId,
				"email_template_id":              c.EmailTemplateId,
				"email_dispatch_status":          c.EmailDispatchStatus,
				"email_dispatch_queued_at":       c.EmailDispatchQueuedAt,
				"email_dispatch_started_at":      c.EmailDispatchStartedAt,
				"email_dispatch_completed_at":    c.EmailDispatchCompletedAt,
				"email_dispatch_last_attempt_at": c.EmailDispatchLastAttemptAt,
				"email_dispatch_last_error":      c.EmailDispatchLastError,
				"email_dispatch_total_targets":   c.EmailDispatchTotalTargets,
				"email_dispatch_sent":            c.EmailDispatchSent,
				"email_dispatch_failed":          c.EmailDispatchFailed,
				"email_dispatch_pending":         c.EmailDispatchPending,
				"total_sent":                     c.TotalSent,
			}).Error; err != nil {
			return err
		}

		return tx.Model(c).Association("Groups").Replace(c.Groups)
	})
}

func (r *CampaignRepository) SaveCampaignTarget(target *campaign.CampaignTarget) error {
	if target.Id > 0 {
		return r.db.Save(target).Error
	}
	return r.db.Create(target).Error
}

func (r *CampaignRepository) GetCampaignTargetByTargetID(campaignID int64, targetID int64) (*campaign.CampaignTarget, error) {
	var ct campaign.CampaignTarget
	err := r.db.
		Where("campaign_id = ? AND target_id = ?", campaignID, targetID).
		First(&ct).Error
	if err != nil {
		return nil, err
	}
	return &ct, nil
}

func (r *CampaignRepository) GetCampaignTargetByToken(campaignID int64, token string) (*campaign.CampaignTarget, error) {
	var ct campaign.CampaignTarget
	err := r.db.
		Where("campaign_id = ? AND token = ?", campaignID, token).
		First(&ct).Error
	if err != nil {
		return nil, err
	}
	return &ct, nil
}

func (r *CampaignRepository) MarkCampaignTargetOpened(campaignTargetID int64, resultID int64, ip string, userAgent string, openedAt time.Time) error {
	return r.db.Model(&campaign.CampaignTarget{}).
		Where("id = ?", campaignTargetID).
		Updates(map[string]interface{}{
			"result_id":  gorm.Expr("CASE WHEN result_id IS NULL THEN ? ELSE result_id END", resultID),
			"status":     gorm.Expr("CASE WHEN status = 'pending' THEN 'sent' ELSE status END"),
			"opened_at":  gorm.Expr("COALESCE(opened_at, ?)", openedAt),
			"ip":         ip,
			"user_agent": userAgent,
			"updated_at": time.Now(),
		}).Error
}

func (r *CampaignRepository) MarkCampaignTargetOpenedIfFirst(
	campaignTargetID int64,
	resultID *int64,
	ip string,
	userAgent string,
	openedAt time.Time,
) (bool, error) {
	updateData := map[string]interface{}{
		"status":     gorm.Expr("CASE WHEN status = 'pending' THEN 'sent' ELSE status END"),
		"opened_at":  openedAt,
		"ip":         ip,
		"user_agent": userAgent,
		"updated_at": time.Now(),
	}

	if resultID != nil {
		updateData["result_id"] = gorm.Expr("CASE WHEN result_id IS NULL THEN ? ELSE result_id END", *resultID)
	}

	res := r.db.Model(&campaign.CampaignTarget{}).
		Where("id = ? AND opened_at IS NULL", campaignTargetID).
		Updates(updateData)
	if res.Error != nil {
		return false, res.Error
	}

	return res.RowsAffected > 0, nil
}

func (r *CampaignRepository) MarkCampaignTargetSubmitted(campaignTargetID int64, submittedAt time.Time) error {
	return r.db.Model(&campaign.CampaignTarget{}).
		Where("id = ?", campaignTargetID).
		Updates(map[string]interface{}{
			"submitted_at": gorm.Expr("COALESCE(submitted_at, ?)", submittedAt),
			"updated_at":   time.Now(),
		}).Error
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

func (r *CampaignRepository) ResetEmailDelivery(campaignID int64, userID int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var camp campaign.Campaign
		if err := tx.Where("id = ? AND user_id = ?", campaignID, userID).First(&camp).Error; err != nil {
			return err
		}

		if err := tx.Where("campaign_id = ?", campaignID).Delete(&campaign.CampaignTarget{}).Error; err != nil {
			return err
		}

		return tx.Model(&campaign.Campaign{}).
			Where("id = ? AND user_id = ?", campaignID, userID).
			Updates(map[string]interface{}{
				"email_dispatch_status":          campaign.EmailDispatchIdle,
				"email_dispatch_queued_at":       nil,
				"email_dispatch_started_at":      nil,
				"email_dispatch_completed_at":    nil,
				"email_dispatch_last_attempt_at": nil,
				"email_dispatch_last_error":      "",
				"email_dispatch_total_targets":   0,
				"email_dispatch_sent":            0,
				"email_dispatch_failed":          0,
				"email_dispatch_pending":         0,
			}).Error
	})
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
		return "strftime('%Y-%m-%dT%H:00:00', events.created_at)"

	case "week":
		return "DATE(events.created_at)"

	case "month":
		return "DATE(events.created_at)"

	case "year":
		return "strftime('%Y-%m', events.created_at)"

	default:
		return "DATE(events.created_at)"
	}

}

func periodFilter(period string) (string, []interface{}) {

	switch period {

	case "day":
		return "date(events.created_at) = date('now')", nil

	case "week":
		return "date(events.created_at) >= date('now', '-6 days')", nil

	case "month":
		return "date(events.created_at) >= date('now', 'start of month')", nil

	case "year":
		return "date(events.created_at) >= date('now', 'start of year')", nil

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
	return r.db.Transaction(func(tx *gorm.DB) error {
		type eventDelta struct {
			Opened    int64
			Clicked   int64
			Submitted int64
		}
		var resultEventDelta eventDelta
		if err := tx.Model(&event.Event{}).
			Select(`
				SUM(CASE WHEN type = ? THEN 1 ELSE 0 END) AS opened,
				SUM(CASE WHEN type = ? THEN 1 ELSE 0 END) AS clicked,
				SUM(CASE WHEN type = ? THEN 1 ELSE 0 END) AS submitted
			`, event.EventOpen, event.EventClick, event.EventSubmit).
			Where("campaign_id = ? AND result_id = ?", campaignID, resultID).
			Scan(&resultEventDelta).Error; err != nil {
			return err
		}

		type targetDelta struct {
			OpenedTargets    int64
			ClickedTargets   int64
			SubmittedTargets int64
		}
		var linkedTargetDelta targetDelta
		if err := tx.Model(&campaign.CampaignTarget{}).
			Select(`
				SUM(CASE WHEN opened_at IS NOT NULL THEN 1 ELSE 0 END) AS opened_targets,
				SUM(CASE WHEN clicked_at IS NOT NULL THEN 1 ELSE 0 END) AS clicked_targets,
				SUM(CASE WHEN submitted_at IS NOT NULL THEN 1 ELSE 0 END) AS submitted_targets
			`).
			Where("campaign_id = ? AND result_id = ?", campaignID, resultID).
			Scan(&linkedTargetDelta).Error; err != nil {
			return err
		}

		deleteRes := tx.
			Where("id = ? AND campaign_id = ?", resultID, campaignID).
			Delete(&result.Result{})
		if deleteRes.Error != nil {
			return deleteRes.Error
		}

		// Keep campaign_targets consistent when a linked result is removed.
		if err := tx.Model(&campaign.CampaignTarget{}).
			Where("campaign_id = ? AND result_id = ?", campaignID, resultID).
			Updates(map[string]interface{}{
				"result_id":    nil,
				"opened_at":    nil,
				"clicked_at":   nil,
				"submitted_at": nil,
				"ip":           "",
				"user_agent":   "",
				"updated_at":   time.Now(),
			}).Error; err != nil {
			return err
		}

		return tx.Model(&campaign.Campaign{}).
			Where("id = ?", campaignID).
			Updates(map[string]interface{}{
				"total_opened": gorm.Expr(
					"CASE WHEN total_opened > ? THEN total_opened - ? ELSE 0 END",
					linkedTargetDelta.OpenedTargets,
					linkedTargetDelta.OpenedTargets,
				),
				// Current runtime increments clicks on submit events.
				"total_clicked": gorm.Expr(
					"CASE WHEN total_clicked > ? THEN total_clicked - ? ELSE 0 END",
					resultEventDelta.Submitted,
					resultEventDelta.Submitted,
				),
				"total_submitted": gorm.Expr(
					"CASE WHEN total_submitted > ? THEN total_submitted - ? ELSE 0 END",
					resultEventDelta.Submitted,
					resultEventDelta.Submitted,
				),
				"unique_opened": gorm.Expr(
					"CASE WHEN unique_opened > ? THEN unique_opened - ? ELSE 0 END",
					linkedTargetDelta.OpenedTargets,
					linkedTargetDelta.OpenedTargets,
				),
				"unique_clicked": gorm.Expr(
					"CASE WHEN unique_clicked > ? THEN unique_clicked - ? ELSE 0 END",
					linkedTargetDelta.ClickedTargets,
					linkedTargetDelta.ClickedTargets,
				),
				"unique_submitted": gorm.Expr(
					"CASE WHEN unique_submitted > ? THEN unique_submitted - ? ELSE 0 END",
					linkedTargetDelta.SubmittedTargets,
					linkedTargetDelta.SubmittedTargets,
				),
				"updated_at": time.Now(),
			}).Error
	})
}
