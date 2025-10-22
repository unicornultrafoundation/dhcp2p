package jobs

import (
	"context"
	"time"

	"github.com/duchuongnguyen/dhcp2p/internal/app/domain/ports"
	"github.com/duchuongnguyen/dhcp2p/internal/app/infrastructure/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type NonceCleanerJob struct {
	repo     ports.NonceRepository
	interval time.Duration
	logger   *zap.Logger

	stopCh chan struct{}
}

var _ ports.NonceCleaner = &NonceCleanerJob{}

func NewNonceCleanerJob(lc fx.Lifecycle, cfg *config.AppConfig, repo ports.NonceRepository, logger *zap.Logger) *NonceCleanerJob {
	j := &NonceCleanerJob{repo, time.Duration(cfg.NonceCleanerInterval) * time.Minute, logger.With(zap.String("job", "nonce_cleaner")), make(chan struct{})}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return j.Run(ctx)
		},
		OnStop: func(ctx context.Context) error {
			close(j.stopCh)
			return nil
		},
	})

	return j
}

func (j *NonceCleanerJob) Run(ctx context.Context) error {
	go func() {
		runCtx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ticker := time.NewTicker(j.interval)
		defer ticker.Stop()

		// Delete expired nonces on start
		j.run(runCtx)

		for {
			select {
			case <-j.stopCh:
				return
			case <-ticker.C:
				j.run(runCtx)
			}
		}
	}()

	return nil
}

func (j *NonceCleanerJob) run(ctx context.Context) {
	err := j.repo.DeleteExpiredNonces(ctx)
	if err != nil {
		j.logger.Error("Failed to delete expired nonces", zap.Error(err))
		return
	}

	j.logger.Info("Deleted expired nonces")
}
