package handlers

import (
	"flexphish/internal/config"
	"flexphish/internal/domain/campaign"
	"flexphish/pkg/logger"
	"sync"
	"time"

	"go.uber.org/zap"
)

type CampaignEmailScheduler struct {
	repo     campaign.Repository
	handler  *CampaignHandler
	enabled  bool
	interval time.Duration

	queue chan campaign.Campaign

	queued   sync.Map
	inFlight sync.Map
}

func NewCampaignEmailScheduler(
	repo campaign.Repository,
	handler *CampaignHandler,
	cfg config.EmailSchedulerConfig,
) *CampaignEmailScheduler {
	intervalSeconds := cfg.PollIntervalSeconds
	if intervalSeconds <= 0 {
		intervalSeconds = 15
	}

	const queueBufferSize = 512

	return &CampaignEmailScheduler{
		repo:     repo,
		handler:  handler,
		enabled:  cfg.Enabled,
		interval: time.Duration(intervalSeconds) * time.Second,
		queue:    make(chan campaign.Campaign, queueBufferSize),
	}
}

func (s *CampaignEmailScheduler) Start() {
	if s.enabled {
		logger.Log.Info("Campaign email scheduler started",
			zap.Duration("interval", s.interval),
			zap.String("mode", "global_fifo_queue"),
		)
		go s.workerLoop()
	} else {
		logger.Log.Info("Campaign email dispatch queue disabled; scheduled campaign activation is still enabled",
			zap.Duration("interval", s.interval),
		)
	}

	go func() {
		s.processScheduledCampaignStarts()
		if s.enabled {
			s.processTick()
		}

		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for range ticker.C {
			s.processScheduledCampaignStarts()
			if s.enabled {
				s.processTick()
			}
		}
	}()
}

func (s *CampaignEmailScheduler) workerLoop() {
	for camp := range s.queue {
		s.queued.Delete(camp.Id)
		s.inFlight.Store(camp.Id, true)

		s.handler.sendCampaignEmailsInBackground(camp.Id, camp.UserId)

		s.inFlight.Delete(camp.Id)
	}
}

func (s *CampaignEmailScheduler) processTick() {
	campaigns, err := s.repo.ListEmailDispatchCandidates()
	if err != nil {
		logger.Log.Error("scheduler tick failed to list campaigns", zap.Error(err))
		return
	}

	for _, camp := range campaigns {
		if camp.DevMode {
			continue
		}

		if _, running := s.inFlight.Load(camp.Id); running {
			continue
		}

		if _, alreadyQueued := s.queued.LoadOrStore(camp.Id, true); alreadyQueued {
			continue
		}

		if camp.EmailDispatchStatus != campaign.EmailDispatchQueued {
			now := time.Now()
			camp.EmailDispatchStatus = campaign.EmailDispatchQueued
			camp.EmailDispatchQueuedAt = &now
			camp.EmailDispatchLastError = ""
			_ = s.repo.Update(&camp)
		}

		select {
		case s.queue <- camp:
		default:
			s.queued.Delete(camp.Id)
			logger.Log.Warn("email scheduler queue is full; campaign will be retried on next tick",
				zap.Int64("campaign_id", camp.Id),
				zap.Int("queue_depth", len(s.queue)),
			)
		}
	}
}

func (s *CampaignEmailScheduler) processScheduledCampaignStarts() {
	campaigns, err := s.repo.ListScheduledStartCandidates(time.Now().UTC())
	if err != nil {
		logger.Log.Error("scheduler tick failed to list scheduled campaigns", zap.Error(err))
		return
	}

	for _, camp := range campaigns {
		if err := s.handler.activateScheduledCampaign(camp.Id, camp.UserId); err != nil {
			logger.Log.Error("failed to activate scheduled campaign",
				zap.Int64("campaign_id", camp.Id),
				zap.Int64("user_id", camp.UserId),
				zap.Error(err),
			)
		}
	}
}
