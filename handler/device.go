package handler

import (
	"errors"
	"fmt"
	"mainPackage/model"
	"mainPackage/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// @summary Get Device IoT
// @tags Device Iot
// @security ApiKeyAuth
// @id Get Device IoT
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/devices [get]
func GetDeviceIoT(c *gin.Context) {
	logger := utils.GetLog()

	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	txtId := uuid.New().String()

	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	orgId := GetVariableFromToken(c, "orgId")
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

	query := `SELECT "orgId", "deviceId", "deviceType", "model", "firmwareVer",
	                 "latitude", "longitude", "ipAddress", "macAddress", 
	                 "createdAt", "updatedAt", "createdBy", "updatedBy"
	          FROM public.device_iot WHERE "orgId" = $1
	          LIMIT $2 OFFSET $3`

	rows, err := conn.Query(ctx, query, orgId, length, start)
	if err != nil {
		logger.Warn("Query failed", zap.Error(err))
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "DeviceIoT", "GetDeviceIoT", "",
			"search", -1, start_time, GetQueryParams(c), response, "Query : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()

	var devices []model.DeviceIoT
	found := false
	for rows.Next() {
		found = true
		var d model.DeviceIoT
		err := rows.Scan(
			&d.OrgID,
			&d.DeviceID,
			&d.DeviceType,
			&d.Model,
			&d.FirmwareVer,
			&d.Latitude,
			&d.Longitude,
			&d.IPAddress,
			&d.MacAddress,
			&d.CreatedAt,
			&d.UpdatedAt,
			&d.CreatedBy,
			&d.UpdatedBy,
		)
		if err != nil {
			logger.Warn("Scan failed", zap.Error(err))
			response := model.Response{
				Status: "-1",
				Msg:    "Failure",
				Desc:   err.Error(),
			}
			//=======AUDIT_START=====//
			_ = utils.InsertAuditLogs(
				c, conn, orgId.(string), username.(string),
				txtId, id, "DeviceIoT", "GetDeviceIoT", "",
				"search", -1, start_time, GetQueryParams(c), response, "Scan : "+err.Error(),
			)
			//=======AUDIT_END=====//
			c.JSON(http.StatusInternalServerError, response)
			return
		}
		devices = append(devices, d)
	}

	if !found {
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   "Not found",
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "DeviceIoT", "GetDeviceIoT", "",
			"search", -1, start_time, GetQueryParams(c), response, "Not Found",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   devices,
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, id, "DeviceIoT", "GetDeviceIoT", "",
		"search", 0, start_time, GetQueryParams(c), response, "GetDeviceIoT Success",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}

// @summary Get Device IoT By ID
// @tags Device Iot
// @security ApiKeyAuth
// @id Get Device IoT By ID
// @accept json
// @produce json
// @Param id path string true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/devices/{id} [get]
func GetDeviceIoTById(c *gin.Context) {
	logger := utils.GetLog()
	deviceId := c.Param("id") // path param like /device-iot/:id

	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	txtId := uuid.New().String()

	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	orgId := GetVariableFromToken(c, "orgId")
	query := `SELECT
				"orgId",
				"deviceId",
				"deviceType",
				"model",
				"firmwareVer",
				"latitude",
				"longitude",
				"ipAddress",
				"macAddress",
				"createdAt",
				"updatedAt",
				"createdBy",
				"updatedBy"
			  FROM public.device_iot
			  WHERE "deviceId"=$1 AND "orgId"=$2`

	var device model.DeviceIoT
	logger.Debug("Executing query", zap.String("query", query), zap.String("deviceId", deviceId))

	err := conn.QueryRow(ctx, query, deviceId, orgId).Scan(
		&device.OrgID,
		&device.DeviceID,
		&device.DeviceType,
		&device.Model,
		&device.FirmwareVer,
		&device.Latitude,
		&device.Longitude,
		&device.IPAddress,
		&device.MacAddress,
		&device.CreatedAt,
		&device.UpdatedAt,
		&device.CreatedBy,
		&device.UpdatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			response := model.Response{
				Status: "-1",
				Msg:    "Not Found",
				Desc:   fmt.Sprintf("No device found with deviceId %s", deviceId),
			}
			//=======AUDIT_START=====//
			_ = utils.InsertAuditLogs(
				c, conn, orgId.(string), username.(string),
				txtId, deviceId, "DeviceIoT", "GetDeviceIoTById", "",
				"search", -1, start_time, GetQueryParams(c), response, "Not Found : "+err.Error(),
			)
			//=======AUDIT_END=====//
			c.JSON(http.StatusNotFound, response)
			return
		}
		logger.Warn("Query failed", zap.Error(err))
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, deviceId, "DeviceIoT", "GetDeviceIoTById", "",
			"search", -1, start_time, GetQueryParams(c), response, "Query failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   device,
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, deviceId, "DeviceIoT", "GetDeviceIoTById", "",
		"search", 0, start_time, GetQueryParams(c), response, "GetDeviceIoTById Success",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}
