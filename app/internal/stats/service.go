package stats

import (
	"context"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"FeatherProxy/app/internal/database"
	"FeatherProxy/app/internal/database/schema"
)

const (
	defaultBatchSize     = 50
	defaultFlushInterval = 5 * time.Second
	defaultChannelCap    = 1000
	defaultRetentionDays = 30
)

// Config holds stats service configuration (from env or defaults).
type Config struct {
	BatchSize     int
	FlushInterval time.Duration
	ChannelCap    int
	RetentionDays int
}

// ConfigFromEnv returns config from environment (STATS_BATCH_SIZE, STATS_FLUSH_INTERVAL, etc.).
func ConfigFromEnv() Config {
	c := Config{
		BatchSize:     defaultBatchSize,
		FlushInterval: defaultFlushInterval,
		ChannelCap:    defaultChannelCap,
		RetentionDays: defaultRetentionDays,
	}
	if v := os.Getenv("STATS_BATCH_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			c.BatchSize = n
		}
	}
	if v := os.Getenv("STATS_FLUSH_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			c.FlushInterval = d
		}
	}
	if v := os.Getenv("STATS_CHANNEL_CAP"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			c.ChannelCap = n
		}
	}
	if v := os.Getenv("STATS_RETENTION_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			c.RetentionDays = n
		}
	}
	return c
}

// Service runs the async stats worker (batch insert) and periodic vacuum.
// It implements Recorder: Record() sends to a channel and does not block.
type Service struct {
	repo   database.Repository
	config Config
	ch     chan schema.ProxyStat
	wg     sync.WaitGroup
}

// NewService creates a stats service that will use the given repository for persistence.
func NewService(repo database.Repository, config Config) *Service {
	if config.BatchSize <= 0 {
		config.BatchSize = defaultBatchSize
	}
	if config.FlushInterval <= 0 {
		config.FlushInterval = defaultFlushInterval
	}
	if config.ChannelCap <= 0 {
		config.ChannelCap = defaultChannelCap
	}
	if config.RetentionDays <= 0 {
		config.RetentionDays = defaultRetentionDays
	}
	return &Service{
		repo:   repo,
		config: config,
		ch:     make(chan schema.ProxyStat, config.ChannelCap),
	}
}

// Record sends the stat to the worker channel. Non-blocking; drops and logs if channel full.
func (s *Service) Record(stat schema.ProxyStat) {
	select {
	case s.ch <- stat:
	default:
		log.Printf("stats: channel full, dropping stat for %s %s", stat.Method, stat.Path)
	}
}

// Run starts the worker and vacuum loop. Blocks until ctx is cancelled.
// On shutdown, flushes any remaining batch before returning.
func (s *Service) Run(ctx context.Context) {
	// Worker: read from channel, batch, flush on N or T
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.runWorker(ctx)
	}()

	// Vacuum ticker: once per day (or proportional to retention)
	vacuumInterval := 24 * time.Hour
	if s.config.RetentionDays > 0 {
		vacuumInterval = time.Duration(s.config.RetentionDays) * 24 * time.Hour / 2 // run at least twice per retention window
		if vacuumInterval > 24*time.Hour {
			vacuumInterval = 24 * time.Hour
		}
	}
	ticker := time.NewTicker(vacuumInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.wg.Wait()
			return
		case <-ticker.C:
			until := time.Now().Add(-time.Duration(s.config.RetentionDays) * 24 * time.Hour)
			if err := s.repo.DeleteProxyStatsOlderThan(until); err != nil {
				log.Printf("stats: vacuum failed: %v", err)
			}
		}
	}
}

func (s *Service) runWorker(ctx context.Context) {
	batch := make([]schema.ProxyStat, 0, s.config.BatchSize*2)
	flushTimer := time.NewTimer(s.config.FlushInterval)
	defer flushTimer.Stop()
	flush := func() {
		if len(batch) == 0 {
			return
		}
		toInsert := make([]schema.ProxyStat, len(batch))
		copy(toInsert, batch)
		batch = batch[:0]
		if err := s.repo.CreateProxyStats(toInsert); err != nil {
			log.Printf("stats: batch insert failed: %v", err)
		}
		if !flushTimer.Stop() {
			select {
			case <-flushTimer.C:
			default:
			}
		}
		flushTimer.Reset(s.config.FlushInterval)
	}

	for {
		select {
		case stat, ok := <-s.ch:
			if !ok {
				flush()
				return
			}
			batch = append(batch, stat)
			if len(batch) >= s.config.BatchSize {
				flush()
			}
		case <-flushTimer.C:
			flush()
			flushTimer.Reset(s.config.FlushInterval)
		case <-ctx.Done():
			// Drain channel best-effort, then flush
			for {
				select {
				case stat, ok := <-s.ch:
					if !ok {
						flush()
						return
					}
					batch = append(batch, stat)
					if len(batch) >= s.config.BatchSize {
						flush()
					}
				default:
					flush()
					return
				}
			}
		}
	}
}
