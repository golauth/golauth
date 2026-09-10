package token

import "time"

// LockoutPolicy configures the per-account lockout applied by generateToken
// after repeated failures. Zero values fall back to the defaults, so a caller
// that passes an empty struct still gets sane behaviour.
type LockoutPolicy struct {
	// Threshold is the number of consecutive failures that must accumulate
	// before any lock is applied.
	Threshold int
	// BaseDelay is the lock duration at exactly Threshold failures. Each
	// further failure doubles it.
	BaseDelay time.Duration
	// MaxDelay caps the doubling.
	MaxDelay time.Duration
}

// DefaultLockoutPolicy is used when NewGenerateToken is given a zero policy.
var DefaultLockoutPolicy = LockoutPolicy{
	Threshold: 5,
	BaseDelay: 1 * time.Minute,
	MaxDelay:  15 * time.Minute,
}

func (p LockoutPolicy) withDefaults() LockoutPolicy {
	if p.Threshold <= 0 {
		p.Threshold = DefaultLockoutPolicy.Threshold
	}
	if p.BaseDelay <= 0 {
		p.BaseDelay = DefaultLockoutPolicy.BaseDelay
	}
	if p.MaxDelay < p.BaseDelay {
		p.MaxDelay = DefaultLockoutPolicy.MaxDelay
	}
	if p.MaxDelay < p.BaseDelay {
		p.MaxDelay = p.BaseDelay
	}
	return p
}

// lockDuration is the back-off for a given number of consecutive failures:
// nothing below the threshold, BaseDelay at the threshold, doubling per extra
// failure, capped at MaxDelay.
func (p LockoutPolicy) lockDuration(failures int) time.Duration {
	if failures < p.Threshold {
		return 0
	}
	d := p.BaseDelay
	for i := p.Threshold; i < failures; i++ {
		d *= 2
		if d >= p.MaxDelay {
			return p.MaxDelay
		}
	}
	if d > p.MaxDelay {
		return p.MaxDelay
	}
	return d
}
