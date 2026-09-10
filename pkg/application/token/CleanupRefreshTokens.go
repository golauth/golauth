package token

import (
	"context"
	"time"

	"github.com/golauth/golauth/pkg/domain/repository"
	"github.com/sirupsen/logrus"
)

// DefaultCleanupInterval is how often expired refresh-token rows are purged when
// REFRESH_TOKEN_CLEANUP_INTERVAL is unset.
const DefaultCleanupInterval = time.Hour

// StartRefreshTokenCleanup purges expired refresh-token rows once immediately and
// then every interval until ctx is cancelled. An unbounded table is an
// operational bug waiting to happen; a deployment that prefers an external cron
// can set the interval very long and run `DELETE FROM golauth_refresh_token
// WHERE expires_at < now()` on its own schedule.
func StartRefreshTokenCleanup(ctx context.Context, repo repository.RefreshTokenRepository, interval time.Duration) {
	if interval <= 0 {
		interval = DefaultCleanupInterval
	}
	purge := func() {
		n, err := repo.DeleteExpired(context.Background(), time.Now())
		if err != nil {
			logrus.Warnf("refresh token cleanup failed: %v", err)
			return
		}
		if n > 0 {
			logrus.Infof("refresh token cleanup removed %d expired row(s)", n)
		}
	}

	purge()
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				purge()
			}
		}
	}()
}
