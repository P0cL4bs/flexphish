package runtime

import (
	"crypto/rand"
	"encoding/hex"
	"flexphish/internal/domain/campaign"
	"flexphish/internal/domain/result"
	"net/http"
	"time"
)

type sessionService struct {
	repo         result.Repository
	geoService   *GeoService
	campaignRepo campaign.Repository
	testMode     bool
}

func NewSessionService(r result.Repository, geoService *GeoService, c campaign.Repository) SessionService {
	return &sessionService{repo: r, geoService: geoService, campaignRepo: c, testMode: false}
}

func (s *sessionService) SetTestMode(enabled bool) {
	s.testMode = enabled
}

func generateSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *sessionService) Resolve(w http.ResponseWriter, r *http.Request, campaignId int64, campaignToken string) (*result.Result, error) {

	cookie, err := r.Cookie("fp_session")
	var sessionID string

	if err != nil {
		sessionID = generateSessionID()

		http.SetCookie(w, &http.Cookie{
			Name:     "fp_session",
			Value:    sessionID,
			Path:     "/",
			HttpOnly: true,
		})
	} else {
		sessionID = cookie.Value
	}
	if !s.testMode {
		res, err := s.repo.FindBySessionID(sessionID)
		if err != nil {
			return nil, err
		}
		info := ExtractVisitorInfo(r)
		country, city := s.geoService.Lookup(info.IP)
		info.Country = country
		info.City = city

		if res == nil {
			res = &result.Result{
				CampaignId: campaignId,
				SessionID:  sessionID,
				Status:     result.ResultInProgress,
				FirstSeen:  time.Now(),
				LastSeen:   time.Now(),
				IP:         info.IP,
				UserAgent:  info.UserAgent,
				Device:     info.Device,
				OS:         info.OS,
				Browser:    info.Browser,
				Country:    info.Country,
				City:       info.City,
			}
			created, err := s.repo.Create(res)
			if err != nil {
				return nil, err
			}

			if campaignToken != "" {
				campaignTarget, targetErr := s.campaignRepo.GetCampaignTargetByToken(campaignId, campaignToken)
				if targetErr == nil && campaignTarget != nil {
					created.CampaignTargetId = &campaignTarget.Id
					_ = s.repo.Update(created)
					_ = s.campaignRepo.MarkCampaignTargetOpened(campaignTarget.Id, created.Id, info.IP, info.UserAgent, time.Now())
				}
			}

			_ = s.campaignRepo.IncrementOpened(campaignId)
			return created, nil
		}

		if campaignToken != "" {
			campaignTarget, targetErr := s.campaignRepo.GetCampaignTargetByToken(campaignId, campaignToken)
			if targetErr == nil && campaignTarget != nil {
				if res.CampaignTargetId == nil {
					res.CampaignTargetId = &campaignTarget.Id
				}
				_ = s.repo.Update(res)
				_ = s.campaignRepo.MarkCampaignTargetOpened(campaignTarget.Id, res.Id, info.IP, info.UserAgent, time.Now())
			}
		}

		res.Browser = info.Browser
		res.Device = info.Device
		res.OS = info.OS
		res.IP = info.IP
		res.UserAgent = info.UserAgent

		res.LastSeen = time.Now()
		s.repo.Update(res)
		return res, nil
	}
	info := ExtractVisitorInfo(r)
	country, city := s.geoService.Lookup(info.IP)
	info.Country = country
	info.City = city

	res := &result.Result{
		CampaignId: campaignId,
		SessionID:  sessionID,
		Status:     result.ResultInProgress,
		FirstSeen:  time.Now(),
		LastSeen:   time.Now(),
		IP:         info.IP,
		UserAgent:  info.UserAgent,
		Device:     info.Device,
		OS:         info.OS,
		Browser:    info.Browser,
		Country:    info.Country,
		City:       info.City,
	}
	if campaignToken != "" {
		campaignTarget, targetErr := s.campaignRepo.GetCampaignTargetByToken(campaignId, campaignToken)
		if targetErr == nil && campaignTarget != nil {
			res.CampaignTargetId = &campaignTarget.Id
		}
	}
	return res, nil
}
