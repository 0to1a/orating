// Package cron is a minimal in-process scheduler. Each job runs once at
// Start, then every Interval until ctx is cancelled.
package cron

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type Job struct {
	Name     string
	Interval time.Duration
	Run      func(ctx context.Context) error
}

type Scheduler struct {
	jobs []Job
	wg   sync.WaitGroup
	log  *slog.Logger
}

// New membuat scheduler baru. logger boleh nil, default slog.Default.
func New(logger *slog.Logger) *Scheduler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Scheduler{log: logger}
}

func (s *Scheduler) Register(j Job) {
	s.jobs = append(s.jobs, j)
}

func (s *Scheduler) Start(ctx context.Context) {
	for _, j := range s.jobs {
		s.wg.Add(1)
		go s.run(ctx, j)
	}
}

// Wait blocks until every job goroutine returns after ctx cancellation.
func (s *Scheduler) Wait() {
	s.wg.Wait()
}

func (s *Scheduler) run(ctx context.Context, j Job) {
	defer s.wg.Done()

	exec := func() {
		if err := j.Run(ctx); err != nil {
			s.log.Error("cron job failed", "job", j.Name, "error", err)
		}
	}

	exec()

	t := time.NewTicker(j.Interval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			exec()
		}
	}
}
