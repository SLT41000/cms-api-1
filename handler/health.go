package handler

import (
	"context"
	"fmt"
	"mainPackage/model"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
)

var RateLimiterInstance *limiter.Limiter
var RateLimit int64
var RatePeriod time.Duration

// @summary AS Health
// @tags AS Health
// @security ApiKeyAuth
// @id Health
// @accept json
// @produce json
// @response 200 {object} model.Response "OK - Request successful"
// @Router /health [get]
func Health(c *gin.Context) {
	currentTime := time.Now().Format("06/01/02 15:04:05")
	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   fmt.Sprintf("HealthCheck OK - %s", currentTime),
	})

}

// RateLimitHandler returns the current rate limit status
// @summary Rate Limit
// @tags AS Health
// @security ApiKeyAuth
// @accept json
// @produce json
// @response 200 {object} model.Response "OK - Request successful"
// @Router /rate_limit [get]
func Ratelimit(c *gin.Context) {
	if RateLimiterInstance == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "rate limiter not initialized"})
		return
	}

	ctx := context.Background()

	// Global rate (decrement)
	globalStatus, err := RateLimiterInstance.Get(ctx, "global:rate")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	globalReset := time.Unix(globalStatus.Reset, 0).UTC()

	// Per-IP rate (decrement)
	ip := c.ClientIP()
	ipStatus, err := RateLimiterInstance.Get(ctx, fmt.Sprintf("ip:%s", ip))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ipReset := time.Unix(ipStatus.Reset, 0).UTC()

	c.JSON(http.StatusOK, gin.H{
		"limit":     RateLimit, // from config
		"remaining": globalStatus.Remaining,
		"reset":     globalReset.Format(time.RFC3339),
		"reached":   globalStatus.Reached,
		"per_ip": map[string]interface{}{
			"ip":        ip,
			"limit":     RateLimit,
			"remaining": ipStatus.Remaining,
			"reset":     ipReset.Format(time.RFC3339),
			"reached":   ipStatus.Reached,
		},
	})
}
