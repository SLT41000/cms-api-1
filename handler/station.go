package handler

import (
	"mainPackage/config"
	"mainPackage/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// Stations godoc
// @summary Get Stations
// @tags Stations
// @security ApiKeyAuth
// @id Get Stations
// @accept json
// @produce json
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/stations [get]
func GetStation(c *gin.Context) {
	logger := config.GetLog()

	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	orgId := GetVariableFromToken(c, "orgId")
	query := `SELECT  "orgId", "deptId", "commId", "stnId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy" 
	FROM public.sec_stations WHERE "orgId"=$1 LIMIT 1000`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	rows, err := conn.Query(ctx, query, orgId)
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
	var Station model.Station
	var StationList []model.Station
	rowIndex := 0
	for rows.Next() {
		rowIndex++
		err := rows.Scan(&Station.DeptID, &Station.OrgID, &Station.CommID, &Station.StnID, &Station.En, &Station.Th,
			&Station.Active, &Station.CreatedAt, &Station.UpdatedAt, &Station.CreatedBy, &Station.UpdatedBy)
		if err != nil {
			logger.Warn("Scan failed", zap.Error(err))
			response := model.Response{
				Status: "-1",
				Msg:    "Failed",
				Desc:   errorMsg,
			}
			c.JSON(http.StatusInternalServerError, response)
		}
		StationList = append(StationList, Station)
	}
	if errorMsg != "" {
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
			Data:   StationList,
			Desc:   "",
		}
		c.JSON(http.StatusOK, response)
	}
}

// Stations godoc
// @summary Get Stations by id
// @tags Stations
// @security ApiKeyAuth
// @id Get Stations  by id
// @accept json
// @produce json
// @Param id path string true "id" "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/stations/{id} [get]
func GetStationbyId(c *gin.Context) {
	logger := config.GetLog()
	id := c.Param("id")
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	orgId := GetVariableFromToken(c, "orgId")
	query := `SELECT  "orgId", "deptId", "commId", "stnId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy" 
	FROM public.sec_stations WHERE "stnId"=$1 AND "orgId"=$2`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	rows, err := conn.Query(ctx, query, id, orgId)
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
	var Station model.Station
	rowIndex := 0
	for rows.Next() {
		rowIndex++
		err := rows.Scan(&Station.DeptID, &Station.OrgID, &Station.CommID, &Station.StnID, &Station.En, &Station.Th,
			&Station.Active, &Station.CreatedAt, &Station.UpdatedAt, &Station.CreatedBy, &Station.UpdatedBy)
		if err != nil {
			logger.Warn("Scan failed", zap.Error(err))
			response := model.Response{
				Status: "-1",
				Msg:    "Failed",
				Desc:   errorMsg,
			}
			c.JSON(http.StatusInternalServerError, response)
		}
	}
	if errorMsg != "" {
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
			Data:   Station,
			Desc:   "",
		}
		c.JSON(http.StatusOK, response)
	}
}
