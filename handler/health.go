package handler

import (
	"fmt"
	"mainPackage/model"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

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
