package handler

import (
	"mainPackage/config"
	"mainPackage/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// @summary Get Audit Log
// @tags Audit Log
// @security ApiKeyAuth
// @id Get Audit Log
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/audit_log [get]
func GetAuditlog(c *gin.Context) {
	logger := config.GetLog()

	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	startStr := c.DefaultQuery("start", "0")
	start, err := strconv.Atoi(startStr)
	if err != nil {
		start = 0
	}
	lengthStr := c.DefaultQuery("length", "1000")
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		length = 1000
	}
	query := `SELECT id, "orgId", username, "txId", "uniqueId", "mainFunc", "subFunc", "nameFunc", action, status, duration, "newData", "oldData", "resData", message, "createdAt"
	FROM public.audit_logs LIMIT $1 OFFSET $2`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	rows, err = conn.Query(ctx, query, length, start)
	if err != nil {
		logger.Warn("Query failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		return
	}
	defer rows.Close()
	var errorMsg string
	var AuditLog model.AuditLog
	var AuditLogList []model.AuditLog
	found := false
	for rows.Next() {

		err := rows.Scan(
			&AuditLog.ID,
			&AuditLog.OrgID,
			&AuditLog.Username,
			&AuditLog.TxID,
			&AuditLog.UniqueId,
			&AuditLog.MainFunc,
			&AuditLog.SubFunc,
			&AuditLog.NameFunc,
			&AuditLog.Action,
			&AuditLog.Status,
			&AuditLog.Duration,
			&AuditLog.NewData,
			&AuditLog.OldData,
			&AuditLog.ResData,
			&AuditLog.Message,
			&AuditLog.CreatedAt,
		)
		if err != nil {
			logger.Warn("Scan failed", zap.Error(err))
			response := model.Response{
				Status: "-1",
				Msg:    "Failed",
				Desc:   errorMsg,
			}
			c.JSON(http.StatusInternalServerError, response)
			return
		}
		AuditLogList = append(AuditLogList, AuditLog)
		found = true
	}
	if !found {
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   errorMsg,
		}
		c.JSON(http.StatusInternalServerError, response)
	} else {
		response := model.Response{
			Status: "0",
			Msg:    "Success",
			Data:   AuditLogList,
			Desc:   "",
		}
		c.JSON(http.StatusOK, response)
	}
}

// @summary Get Audit Log By Username
// @tags Audit Log
// @security ApiKeyAuth
// @id Get Audit Log By Username
// @accept json
// @produce json
// @Param username path string true "username"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/audit_log/{username} [get]
func GetAuditlogByUsername(c *gin.Context) {
	logger := config.GetLog()

	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	orgId := GetVariableFromToken(c, "orgId")
	username := c.Param("username")

	query := `SELECT id, "orgId", username, "txId", "uniqueId", "mainFunc", "subFunc", "nameFunc", action, status, duration, "newData", "oldData", "resData", message, "createdAt"
	FROM public.audit_logs WHERE "orgId"=$1 AND username=$2`
	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	rows, err := conn.Query(ctx, query, orgId, username)
	if err != nil {
		logger.Warn("Query failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		return
	}
	defer rows.Close()
	var errorMsg string
	var AuditLog model.AuditLog
	var AuditLogList []model.AuditLog
	found := false
	for rows.Next() {

		err := rows.Scan(
			&AuditLog.ID,
			&AuditLog.OrgID,
			&AuditLog.Username,
			&AuditLog.TxID,
			&AuditLog.UniqueId,
			&AuditLog.MainFunc,
			&AuditLog.SubFunc,
			&AuditLog.NameFunc,
			&AuditLog.Action,
			&AuditLog.Status,
			&AuditLog.Duration,
			&AuditLog.NewData,
			&AuditLog.OldData,
			&AuditLog.ResData,
			&AuditLog.Message,
			&AuditLog.CreatedAt,
		)
		if err != nil {
			logger.Warn("Scan failed", zap.Error(err))
			response := model.Response{
				Status: "-1",
				Msg:    "Failed",
				Desc:   errorMsg,
			}
			c.JSON(http.StatusInternalServerError, response)
			return
		}
		AuditLogList = append(AuditLogList, AuditLog)
		found = true
	}
	if !found {
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   errorMsg,
		}
		c.JSON(http.StatusInternalServerError, response)
	} else {
		response := model.Response{
			Status: "0",
			Msg:    "Success",
			Data:   AuditLogList,
			Desc:   "",
		}
		c.JSON(http.StatusOK, response)
	}
}
