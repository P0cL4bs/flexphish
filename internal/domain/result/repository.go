package result

type Repository interface {
	Create(r *Result) (*Result, error)
	Update(r *Result) error
	FindBySessionID(sessionID string) (*Result, error)
	FindByCampaign(campaignId int64) ([]Result, error)
}
