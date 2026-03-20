package campaign

import "time"

type Repository interface {
	Create(c *Campaign) error
	GetByID(id int64, userId int64) (*Campaign, error)
	ListEmailDispatchCandidates() ([]Campaign, error)
	ListByUser(userId int64, status *CampaignStatus, page, pageSize int) ([]Campaign, int64, error)
	Update(c *Campaign) error
	SaveCampaignTarget(target *CampaignTarget) error
	GetCampaignTargetByTargetID(campaignID int64, targetID int64) (*CampaignTarget, error)
	GetCampaignTargetByToken(campaignID int64, token string) (*CampaignTarget, error)
	MarkCampaignTargetOpened(campaignTargetID int64, resultID int64, ip string, userAgent string, openedAt time.Time) error
	MarkCampaignTargetSubmitted(campaignTargetID int64, submittedAt time.Time) error
	Delete(id int64, userId int64) error
	FindActiveBySubdomain(subdomain string) (*Campaign, error)
	IncrementClicked(id int64) error
	IncrementOpened(id int64) error
	IncrementSubmitted(id int64) error
	FindBySubdomain(subdomain string) (*Campaign, error)
	UpdateStatus(id int64, status CampaignStatus) error
	CountCampaignsUsingTemplateId(templateId string) (int64, error)
	HasActiveCampaignUsingTemplate(templateId string) (bool, error)
	GetTopCampaigns(userID int64) ([]TopCampaignMetric, error)
	GetTimeline(userID int64, period string) ([]TimelineMetric, error)
	GetAnalytics(userID int64, period string) (*CampaignAnalytics, error)
	DeleteResult(resultID int64, campaignID int64) error
}
