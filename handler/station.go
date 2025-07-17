package handler

import (
	"mainPackage/config"
	"mainPackage/model"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// Stations godoc
// @summary Get Stations
// @tags Dispatch
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
	query := `SELECT  id,"orgId", "deptId", "commId", "stnId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy" 
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
		err := rows.Scan(&Station.ID, &Station.DeptID, &Station.OrgID, &Station.CommID, &Station.StnID, &Station.En, &Station.Th,
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
// @tags Dispatch
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
	query := `SELECT  id,"orgId", "deptId", "commId", "stnId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy" 
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
	err = rows.Scan(&Station.ID, &Station.DeptID, &Station.OrgID, &Station.CommID, &Station.StnID, &Station.En, &Station.Th,
		&Station.Active, &Station.CreatedAt, &Station.UpdatedAt, &Station.CreatedBy, &Station.UpdatedBy)
	if err != nil {
		logger.Warn("Scan failed", zap.Error(err))
		errorMsg = err.Error()
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   errorMsg,
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	if errorMsg != "" {
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   errorMsg,
		}
		c.JSON(http.StatusInternalServerError, response)
		return
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

// @summary Create Stations
// @id Create Stations
// @security ApiKeyAuth
// @tags Dispatch
// @accept json
// @produce json
// @param Body body model.StationInsert true "Create Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/stations/add [post]
func InsertStations(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	var req model.StationInsert
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}
	username := GetVariableFromToken(c, "username")
	now := time.Now()
	var id int
	query := `
	INSERT INTO public."sec_stations"(
	"orgId", "deptId", "commId", "stnId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy")
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	RETURNING id ;
	`

	err := conn.QueryRow(ctx, query,
		req.DeptID, req.OrgID, req.DeptID, req.StnID, req.En, req.Th, req.Active, now,
		now, username, username).Scan(&id)

	if err != nil {
		// log.Printf("Insert failed: %v", err)
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	// Continue logic...
	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Create successfully",
	})

}

// @summary Update Stations
// @id Update Stations
// @security ApiKeyAuth
// @accept json
// @produce json
// @tags Dispatch
// @Param id path int true "id"
// @param Body body model.StationUpdate true "Update data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/stations/{id} [patch]
func UpdateStations(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	id := c.Param("id")

	var req model.StationUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Update failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		return
	}
	now := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	query := `UPDATE public."sec_departments"
	SET "deptId"=$3, "commId"=$4, "stnId"=$5, en=$6, th=$7, active=$8,
	 "updatedAt"=$9, "updatedBy"=$10
	WHERE id = $1 AND "orgId"=$2`
	_, err := conn.Exec(ctx, query,
		id, orgId, req.DeptID, req.CommID, req.StnID, req.En, req.Th, req.Active,
		now, username,
	)
	logger.Debug("Update Case SQL Args",
		zap.String("query", query),
		zap.Any("Input", []any{
			id, orgId, req.DeptID, req.CommID, req.StnID, req.En, req.Th, req.Active,
			now, username,
		}))
	if err != nil {
		// log.Printf("Insert failed: %v", err)
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Update failed", zap.Error(err))
		return
	}

	// Continue logic...
	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Update successfully",
	})
}

// @summary Delete Stations
// @id Delete Stations
// @security ApiKeyAuth
// @accept json
// @tags Dispatch
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/stations/{id} [delete]
func DeleteStations(c *gin.Context) {

	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()
	orgId := GetVariableFromToken(c, "orgId")
	id := c.Param("id")
	query := `DELETE FROM public."sec_stations" WHERE id = $1 AND "orgId"=$2`
	logger.Debug("Query", zap.String("query", query), zap.Any("id", id))
	_, err := conn.Exec(ctx, query, id, orgId)
	if err != nil {
		// log.Printf("Insert failed: %v", err)
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Update failed", zap.Error(err))
		return
	}

	// Continue logic...
	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Delete successfully",
	})
}
