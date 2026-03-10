package simplefin

import (
	"context"
	"log/slog"
	"time"
)

type Scheduler struct {
	syncService *SyncService
	syncHour    int
	stop        chan struct{}
	done        chan struct{}
}

func NewScheduler(syncService *SyncService, syncHour int) *Scheduler {
	return &Scheduler{
		syncService: syncService,
		syncHour:    syncHour,
		stop:        make(chan struct{}),
		done:        make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	go s.run()
}

func (s *Scheduler) Stop() {
	close(s.stop)
	<-s.done
}

func (s *Scheduler) run() {
	defer close(s.done)

	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), s.syncHour, 0, 0, 0, now.Location())
		if now.After(next) {
			next = next.Add(24 * time.Hour)
		}

		delay := time.Until(next)
		slog.Info("SimpleFIN sync scheduled", slog.String("next_sync", next.Format(time.RFC3339)), slog.Duration("delay", delay))

		select {
		case <-s.stop:
			return
		case <-time.After(delay):
			s.executeSync()
		}
	}
}

func (s *Scheduler) executeSync() {
	slog.Info("starting scheduled SimpleFIN sync")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	results := s.syncService.SyncAll(ctx)
	for _, r := range results {
		if r.Error != nil {
			slog.Error("scheduled sync failed for user", slog.String("user_id", r.UserID), slog.Any("error", r.Error))
		} else {
			slog.Info("scheduled sync complete for user", slog.String("user_id", r.UserID), slog.Int("updated", r.Updated), slog.Int("errors", r.Errors))
		}
	}
}
