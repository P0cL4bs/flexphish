package event

type Repository interface {
	Create(e *Event) error
	FindByCampaign(campaignId int64) ([]Event, error)
	FindByResult(resultId int64) ([]Event, error)
}
