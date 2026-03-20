package runtime

import (
	"flexphish/internal/domain/event"
	"flexphish/internal/domain/result"
	"net/http"
)

type SessionService interface {
	Resolve(w http.ResponseWriter, r *http.Request, campaignId int64, campaignToken string) (*result.Result, error)
	SetTestMode(enabled bool)
}

type EventService interface {
	RegisterEvent(campaignId int64, resultId *int64, typ event.EventType, stepID string, r *http.Request, metadata map[string]interface{}) error
	RegisterStepSubmit(
		campaignId int64,
		res *result.Result,
		stepId string,
		data map[string]interface{},
		r *http.Request,
		isFinal bool,
	) error
	SetTestMode(enabled bool)
}
