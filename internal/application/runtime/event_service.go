package runtime

import (
	"encoding/json"
	"flexphish/internal/domain/campaign"
	"flexphish/internal/domain/event"
	"flexphish/internal/domain/result"
	"net/http"
	"time"
)

type eventService struct {
	eventRepo    event.Repository
	resultRepo   result.Repository
	campaignRepo campaign.Repository
	testMode     bool
}

func GetClientIP(r *http.Request) string {

	ip := r.Header.Get("X-Forwarded-For")

	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}

	if ip == "" {
		ip = r.RemoteAddr
	}

	return ip
}

func NewEventService(e event.Repository, r result.Repository, c campaign.Repository) EventService {
	return &eventService{
		eventRepo:    e,
		resultRepo:   r,
		campaignRepo: c,
		testMode:     false,
	}
}

func (s *eventService) SetTestMode(enabled bool) {
	s.testMode = enabled
}

func (s *eventService) RegisterEvent(
	campaignId int64,
	resultId *int64,
	typ event.EventType,
	stepID string,
	r *http.Request,
	metadata map[string]interface{},
) error {
	if s.testMode {
		return nil
	}
	var metaStr string
	if metadata != nil {
		b, _ := json.Marshal(metadata)
		metaStr = string(b)
	}
	ev := &event.Event{
		CampaignId: campaignId,
		ResultId:   resultId,
		Type:       typ,
		StepID:     stepID,
		Path:       r.URL.Path,
		IP:         GetClientIP(r),
		UserAgent:  r.UserAgent(),
		Referrer:   r.Referer(),
		Metadata:   metaStr,
		CreatedAt:  time.Now(),
	}

	return s.eventRepo.Create(ev)
}

func (s *eventService) RegisterStepSubmit(
	campaignId int64,
	res *result.Result,
	stepID string,
	captured map[string]interface{},
	r *http.Request,
	isFinal bool,
) error {
	if s.testMode {
		return nil
	}
	if v, ok := captured["email"]; ok {
		res.Email = v.(string)
	}
	if v, ok := captured["username"]; ok {
		res.Username = v.(string)
	}
	if v, ok := captured["password"]; ok {
		res.Password = v.(string)
	}

	res.LastSeen = time.Now()

	if isFinal {
		res.Status = result.ResultCompleted
	}

	if err := s.resultRepo.Update(res); err != nil {
		return err
	}

	if res.CampaignTargetId != nil {
		_ = s.campaignRepo.MarkCampaignTargetSubmitted(*res.CampaignTargetId, time.Now())
	}

	_ = s.campaignRepo.IncrementClicked(campaignId)
	_ = s.campaignRepo.IncrementSubmitted(campaignId)

	return s.RegisterEvent(
		campaignId,
		&res.Id,
		event.EventSubmit,
		stepID,
		r,
		captured,
	)
}
