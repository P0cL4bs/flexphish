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
	if !s.enabled {
		logger.Log.Info("Campaign email scheduler disabled")
		return
	}

	logger.Log.Info("Campaign email scheduler started",
		zap.Duration("interval", s.interval),
		zap.String("mode", "global_fifo_queue"),
	)

	go s.workerLoop()

	go func() {
		s.processTick()

		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for range ticker.C {
			s.processTick()
		}
	}()
}

func (s *CampaignEmailScheduler) workerLoop() {
	for camp := range s.queue {
		s.queued.Delete(camp.Id)
		s.inFlight.Store(camp.Id, true)
		startedAt := time.Now()

		logger.Log.Info("email scheduler dispatch started",
			zap.Int64("campaign_id", camp.Id),
			zap.Int64("user_id", camp.UserId),
			zap.Int("queue_depth", len(s.queue)),
		)

		s.handler.sendCampaignEmailsInBackground(camp.Id, camp.UserId)

		logger.Log.Info("email scheduler dispatch finished",
			zap.Int64("campaign_id", camp.Id),
			zap.Int64("user_id", camp.UserId),
			zap.Duration("duration", time.Since(startedAt)),
			zap.Int("queue_depth", len(s.queue)),
		)

		s.inFlight.Delete(camp.Id)
	}
}

func (s *CampaignEmailScheduler) processTick() {
	campaigns, err := s.repo.ListEmailDispatchCandidates()
	if err != nil {
		logger.Log.Error("scheduler tick failed to list campaigns", zap.Error(err))
		return
	}

	enqueuedCount := 0
	skippedDevMode := 0
	skippedInFlight := 0
	skippedQueued := 0
	queueFullDrops := 0

	for _, camp := range campaigns {
		if camp.DevMode {
			skippedDevMode++
			continue
		}

		if _, running := s.inFlight.Load(camp.Id); running {
			skippedInFlight++
			continue
		}

		if _, alreadyQueued := s.queued.LoadOrStore(camp.Id, true); alreadyQueued {
			skippedQueued++
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
			enqueuedCount++
			logger.Log.Info("campaign enqueued for dispatch",
				zap.Int64("campaign_id", camp.Id),
				zap.Int64("user_id", camp.UserId),
				zap.String("dispatch_status", string(camp.EmailDispatchStatus)),
				zap.Int("queue_depth", len(s.queue)),
			)
		default:
			s.queued.Delete(camp.Id)
			queueFullDrops++
			logger.Log.Warn("email scheduler queue is full; campaign will be retried on next tick",
				zap.Int64("campaign_id", camp.Id),
				zap.Int("queue_depth", len(s.queue)),
			)
		}
	}

	logger.Log.Info("email scheduler tick processed",
		zap.Int("candidates", len(campaigns)),
		zap.Int("enqueued", enqueuedCount),
		zap.Int("skipped_dev_mode", skippedDevMode),
		zap.Int("skipped_in_flight", skippedInFlight),
		zap.Int("skipped_already_queued", skippedQueued),
		zap.Int("queue_full_drops", queueFullDrops),
		zap.Int("queue_depth", len(s.queue)),
	)
}
