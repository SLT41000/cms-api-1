package handler

import (
	"encoding/json"
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

// @summary Get Stations
// @tags Organization
// @security ApiKeyAuth
// @id Get Stations
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/stations [get]
func GetStation(c *gin.Context) {
	txtId := uuid.New().String()
	start_time := time.Now()
	logger := utils.GetLog()

	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")

	conn, ctx, cancel := utils.ConnectDB()
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

	//forLogs := fmt.Sprintf("start=%d,length=%d", start, length)

	query := `SELECT  id,"orgId", "deptId", "commId", "stnId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy" 
    FROM public.sec_stations WHERE "orgId"=$1 LIMIT $2 OFFSET $3`

	var rows pgx.Rows
	logger.Debug(`GetStation : Query`, zap.String("query", query))
	rows, err = conn.Query(ctx, query, orgId, length, start)
	if err != nil {
		///logger.Warn("GetStation :  Query failed", zap.Error(err))
		utils.WriteConsole("error", "GetStation", "Query failed", zap.Error(err))
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   err.Error(),
		}

		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, "", "Station", "GetStation", "",
			"search", -1, start_time, GetQueryParams(c), response, "Query failed = "+err.Error(),
		)

		//=======AUDIT_END=====//

		c.JSON(http.StatusInternalServerError, response)
		return

	}
	defer rows.Close()
	var errorMsg string
	var Station model.Station
	var StationList []model.Station
	rowIndex := 0
	for rows.Next() {
		rowIndex++
		err := rows.Scan(&Station.ID, &Station.OrgID, &Station.DeptID, &Station.CommID, &Station.StnID, &Station.En, &Station.Th,
			&Station.Active, &Station.CreatedAt, &Station.UpdatedAt, &Station.CreatedBy, &Station.UpdatedBy)
		if err != nil {
			//logger.Warn("GetStation : Scan failed", zap.Error(err))
			utils.WriteConsole("error", "GetStation", "Scan failed", zap.Error(err))
			response := model.Response{
				Status: "-2",
				Msg:    "Failed",
				Desc:   err.Error(),
			}

			//=======AUDIT_START=====//
			_ = utils.InsertAuditLogs(
				c, conn, orgId.(string), username.(string),
				txtId, "", "Station", "GetStation", "",
				"search", -2, start_time, GetQueryParams(c), response, "Scan failed = "+err.Error(),
			)
			//=======AUDIT_END=====//

			c.JSON(http.StatusInternalServerError, response)
			return
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
		return
	} else {
		response := model.Response{
			Status: "0",
			Msg:    "Success",
			Data:   StationList,
			Desc:   "",
		}

		response_2 := model.Response{
			Status: "0",
			Msg:    "Success",
			Data:   len(StationList),
			Desc:   "",
		}
		//=======AUDIT_START=====//
		responseJSON, err := json.Marshal(response)
		if err != nil {
			//logger.Error("GetStation : Failed to marshal response", zap.Error(err))
			utils.WriteConsole("error", "GetStation", "Failed response", zap.Error(err))
		} else {
			utils.WriteConsole("Info", "GetStation", "Success", zap.String("response", string(responseJSON)))
			//logger.Info("GetStation : Success", zap.String("response", string(responseJSON)))
		}

		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, "", "Station", "GetStation", "",
			"search", 0, start_time, GetQueryParams(c), response_2, "Get station list successfully",
		)
		//=======AUDIT_END=====//

		c.JSON(http.StatusOK, response)
		return
	}

}

// @summary Get Stations Command Department
// @tags Organization
// @security ApiKeyAuth
// @id Get Stations Command Department
// @accept json
// @produce json
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/department_command_stations [get]
func GetDepartmentCommandStation(c *gin.Context) {
	txtId := uuid.New().String()
	start_time := time.Now()
	//logger := utils.GetLog()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")

	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	// เรียก service layer
	stations, err := utils.GetDepartmentCommandStationOrLoad(ctx, conn, orgId.(string))
	if err != nil {
		utils.WriteConsole("error", "GetDepartmentCommandStation", "Query failed", zap.Error(err))
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   err.Error(),
		}
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, "", "Station", "GetDepartmentCommandStation", "",
			"search", -1, start_time, GetQueryParams(c), response, "Query failed = "+err.Error(),
		)
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   stations,
		Desc:   "",
	}

	response_2 := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   len(stations),
		Desc:   "",
	}

	_, err = json.Marshal(response)
	if err != nil {
		utils.WriteConsole("error", "GetDepartmentCommandStation", "Failed marshal response", zap.Error(err))
	}
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, "", "Station", "GetDepartmentCommandStation", "",
		"search", 0, start_time, GetQueryParams(c), response_2, "Get Department, Command, Station list successfully",
	)

	c.JSON(http.StatusOK, response)
}

// @summary Get Stations by id
// @tags Organization
// @security ApiKeyAuth
// @id Get Stations  by id
// @accept json
// @produce json
// @Param id path string true "id" "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/stations/{id} [get]
func GetStationbyId(c *gin.Context) {
	txtId := uuid.New().String()
	start_time := time.Now()
	//logger := utils.GetLog()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	id := c.Param("id")

	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	query := `SELECT id, "orgId", "deptId", "commId", "stnId", en, th, active,
	                 "createdAt", "updatedAt", "createdBy", "updatedBy"
	          FROM public.sec_stations
	          WHERE "stnId" = $1 AND "orgId" = $2`

	//logger.Debug(`GetStationById : Query`, zap.String("query", query))
	utils.WriteConsole("debug", "GetStationById", "Query", zap.String("query", query))
	var station model.Station
	err := conn.QueryRow(ctx, query, id, orgId).Scan(
		&station.ID, &station.OrgID, &station.DeptID, &station.CommID,
		&station.StnID, &station.En, &station.Th, &station.Active,
		&station.CreatedAt, &station.UpdatedAt, &station.CreatedBy, &station.UpdatedBy,
	)
	if err != nil {
		//logger.Warn("GetStationById : QueryRow failed", zap.Error(err))
		utils.WriteConsole("error", "GetStationById", "QueryRow failed", zap.Error(err))
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   err.Error(),
		}

		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Station", "GetStationById", "",
			"view", -1, start_time, GetQueryParams(c), response, "Query failed = "+err.Error(),
		)
		//=======AUDIT_END=====//

		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Success case
	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   station,
		Desc:   "",
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		//logger.Error("GetDepartmentCommandStation : Failed to marshal response", zap.Error(err))
		utils.WriteConsole("error", "GetStationById", "Failed response", zap.Error(err))
	} else {
		//logger.Info("GetDepartmentCommandStation : Success", zap.String("response", string(responseJSON)))
		utils.WriteConsole("Info", "GetStationById", "Success", zap.String("response", string(responseJSON)))
	}

	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, id, "Station", "GetStationById", "",
		"view", 0, start_time, GetQueryParams(c), response, "Get Station by Id successfully",
	)
	//=======AUDIT_END=====//

	c.JSON(http.StatusOK, response)
}

// @summary Create Stations
// @id Create Stations
// @security ApiKeyAuth
// @tags Organization
// @accept json
// @produce json
// @param Body body model.StationInsert true "Create Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/stations/add [post]
func InsertStations(c *gin.Context) {
	txtId := uuid.New().String()
	start_time := time.Now()
	//logger := utils.GetLog()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	var req model.StationInsert
	if err := c.ShouldBindJSON(&req); err != nil {

		//logger.Warn("Insert failed", zap.Error(err))
		utils.WriteConsole("error", "InsertStations", "Insert failed", zap.Error(err))
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   err.Error(),
		}

		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, "", "Station", "InsertStations", "",
			"create", -1, start_time, req, response, "Insert failed = "+err.Error(),
		)
		//=======AUDIT_END=====//

		c.JSON(http.StatusBadRequest, response)
		return
	}

	if username == "" {

		response := model.Response{
			Status: "-2",
			Msg:    "Failure",
			Desc:   "Missing username in token",
		}

		responseJSON, err := json.Marshal(response)
		if err != nil {
			//logger.Error("GetDepartmentCommandStation : Failed to marshal response", zap.Error(err))
			utils.WriteConsole("error", "InsertStations", "Failed response", zap.Error(err))
		} else {
			//logger.Info("GetDepartmentCommandStation : Success", zap.String("response", string(responseJSON)))
			utils.WriteConsole("Info", "InsertStations", "Success", zap.String("response", string(responseJSON)))
		}

		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, "", "Station", "InsertStations", "",
			"create", -2, start_time, req, response, "Insert failed = Missing username in token",
		)
		//=======AUDIT_END=====//

		c.JSON(http.StatusBadRequest, response)
		return
	}

	now := time.Now()
	stnUUID := uuid.New()
	var id int

	query := `
    INSERT INTO public."sec_stations"(
    "orgId", "deptId", "commId", "stnId", en, th, active,
    "createdAt", "updatedAt", "createdBy", "updatedBy"
    )
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    RETURNING id;
    `

	err := conn.QueryRow(ctx, query,
		orgId,
		req.DeptID,
		req.CommID,
		stnUUID,
		req.En,
		req.Th,
		req.Active,
		now,
		now,
		username,
		username,
	).Scan(&id)

	if err != nil {
		response := model.Response{
			Status: "-3",
			Msg:    "Failure",
			Desc:   err.Error(),
		}

		utils.WriteConsole("error", "InsertStations", "Insert failed", zap.Error(err))
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, "", "Station", "InsertStations", "",
			"create", -3, start_time, req, response, "Insert failed : "+err.Error(),
		)
		//=======AUDIT_END=====//

		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   fmt.Sprintf("Station created successfully (ID: %d)", id),
	}

	//utils.WriteConsole("error", "InsertStations", "Success", response)
	responseJSON, err := json.Marshal(response)
	if err != nil {
		//logger.Error("GetDepartmentCommandStation : Failed to marshal response", zap.Error(err))
		utils.WriteConsole("error", "InsertStations", "Failed response", zap.Error(err))
	} else {
		//logger.Info("GetDepartmentCommandStation : Success", zap.String("response", string(responseJSON)))
		utils.WriteConsole("Info", "InsertStations", "Success", zap.String("response", string(responseJSON)))
	}

	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, stnUUID.String(), "Station", "InsertStations", "",
		"create", 0, start_time, req, response, "Get Station by Id successfully",
	)
	//=======AUDIT_END=====//

	c.JSON(http.StatusOK, response)

}

// @summary Update Stations
// @id Update Stations
// @security ApiKeyAuth
// @accept json
// @produce json
// @tags Organization
// @Param id path string true "id"
// @param Body body model.StationUpdate true "Update data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/stations/{id} [patch]
func UpdateStations(c *gin.Context) {
	txtId := uuid.New().String()
	start_time := time.Now()
	logger := utils.GetLog()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")

	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	id := c.Param("id")

	var req model.StationUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		//logger.Warn("Update failed", zap.Error(err))
		utils.WriteConsole("error", "UpdateStations", "Failed response", zap.Error(err))
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   err.Error(),
		}

		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Station", "UpdateStations", "",
			"update", -1, start_time, req, response, "Insert failed = "+err.Error(),
		)
		//=======AUDIT_END=====//

		c.JSON(http.StatusBadRequest, response)
		return
	}
	now := time.Now()
	query := `UPDATE public."sec_stations"
    SET "deptId"=$3, "commId"=$4, en=$5, th=$6, active=$7,
     "updatedAt"=$8, "updatedBy"=$9
    WHERE "stnId" = $1 AND "orgId"=$2`
	_, err := conn.Exec(ctx, query,
		id, orgId, req.DeptID, req.CommID, req.En, req.Th, req.Active,
		now, username,
	)
	logger.Debug("Update Case SQL Args",
		zap.String("query", query),
		zap.Any("Input", []any{
			id, orgId, req.DeptID, req.CommID, req.En, req.Th, req.Active,
			now, username,
		}))
	if err != nil {
		// log.Printf("Insert failed: %v", err)
		response := model.Response{
			Status: "-2",
			Msg:    "Failure",
			Desc:   err.Error(),
		}

		//logger.Warn("Update failed", zap.Error(err))
		utils.WriteConsole("error", "UpdateStations", "Update failed", zap.Error(err))
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, "", "Station", "UpdateStations", "",
			"update", -1, start_time, req, response, "Update failed = "+err.Error(),
		)
		//=======AUDIT_END=====//

		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Update successfully",
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		//logger.Error("GetDepartmentCommandStation : Failed to marshal response", zap.Error(err))
		utils.WriteConsole("error", "UpdateStations", "Failed response", zap.Error(err))
	} else {
		//logger.Info("GetDepartmentCommandStation : Success", zap.String("response", string(responseJSON)))
		utils.WriteConsole("Info", "UpdateStations", "Success", zap.String("response", string(responseJSON)))
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, id, "Station", "UpdateStations", "",
		"update", 0, start_time, req, response, "Get Station by Id successfully",
	)
	//=======AUDIT_END=====//

	// Continue logic...
	c.JSON(http.StatusOK, response)
}

// @summary Delete Stations
// @id Delete Stations
// @security ApiKeyAuth
// @accept json
// @tags Organization
// @produce json
// @Param id path string true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/stations/{id} [delete]
func DeleteStations(c *gin.Context) {

	txtId := uuid.New().String()
	start_time := time.Now()
	logger := utils.GetLog()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()
	print("====")
	id := c.Param("id")
	print(id)

	query := `DELETE FROM public."sec_stations" WHERE "stnId" = $1 AND "orgId"=$2`
	logger.Debug("Query", zap.String("query", query), zap.Any("id", id))
	_, err := conn.Exec(ctx, query, id, orgId)
	print(GetQueryParams(c))

	if err != nil {
		// log.Printf("Insert failed: %v", err)
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//logger.Warn("Delete failed", zap.Error(err))
		utils.WriteConsole("error", "DeleteStations", "Delete failed", zap.Error(err))
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, "", "Station", "DeleteStations", "",
			"delete", -1, start_time, GetQueryParams(c), response, "Delete failed = "+err.Error(),
		)
		//=======AUDIT_END=====//

		c.JSON(http.StatusInternalServerError, response)

		return
	}

	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Delete successfully",
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		//logger.Error("GetDepartmentCommandStation : Failed to marshal response", zap.Error(err))
		utils.WriteConsole("error", "DeleteStations", "Failed response", zap.Error(err))
	} else {
		//logger.Info("GetDepartmentCommandStation : Success", zap.String("response", string(responseJSON)))
		utils.WriteConsole("Info", "DeleteStations", "Success", zap.String("response", string(responseJSON)))
	}

	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, id, "Station", "DeleteStations", "",
		"delete", 0, start_time, GetQueryParams(c), response, "Delete Station by Id successfully",
	)
	//=======AUDIT_END=====//
	// Continue logic...
	c.JSON(http.StatusOK, response)
}
