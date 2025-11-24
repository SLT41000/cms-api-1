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

	key := "global:rate"                                      // or c.ClientIP() for per-IP
	ctx := context.TODO()                                     // do not pass nil
	limiterContext, err := RateLimiterInstance.Peek(ctx, key) // Peek does NOT decrement
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resetTime := time.Unix(limiterContext.Reset, 0).UTC()

	resp := map[string]interface{}{
		"limit":     limiterContext.Limit,
		"remaining": limiterContext.Remaining,
		"reset":     resetTime.Format(time.RFC3339),
		"reached":   limiterContext.Reached,
	}

	c.JSON(http.StatusOK, resp)
}
