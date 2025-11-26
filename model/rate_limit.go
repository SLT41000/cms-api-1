package model

import (
	"time"

	"github.com/ulule/limiter/v3"
)

type Handler struct {
	RateLimiterInstance *limiter.Limiter
	RateLimit           int64
	RatePeriod          time.Duration
}
