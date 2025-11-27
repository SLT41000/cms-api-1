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
	ctx := context.Background()

	// 👇 ดูค่าแต่ไม่ลด
	globalStatus, err := RateLimiterInstance.Peek(ctx, "global:rate")
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 👇 ลดเฉพาะของ IP
	ip := c.ClientIP()
	ipStatus, err := RateLimiterInstance.Get(ctx, fmt.Sprintf("ip:%s", ip))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"limit":     RateLimit,
		"remaining": globalStatus.Remaining,
		"reset":     time.Unix(globalStatus.Reset, 0).UTC(),
		"reached":   globalStatus.Reached,
		"per_ip": map[string]interface{}{
			"ip":        ip,
			"limit":     RateLimit,
			"remaining": ipStatus.Remaining,
			"reset":     time.Unix(ipStatus.Reset, 0).UTC(),
			"reached":   ipStatus.Reached,
		},
	})
}
