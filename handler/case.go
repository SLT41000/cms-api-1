package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"mainPackage/config"
	"mainPackage/model"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func Process(function string, key string, status string, input interface{}, output interface{}) string {
	// Marshal input to JSON
	inputJSON, err := json.Marshal(input)
	if err != nil {
		inputJSON = []byte(`"marshal input error"`)
	}

	// Marshal output to JSON
	outputJSON, err := json.Marshal(output)
	if err != nil {
		outputJSON = []byte(`"marshal output error"`)
	}

	// Format final log string
	return fmt.Sprintf("[%s][%s][%s][%s][%s]", function, key, status, string(inputJSON), string(outputJSON))
}

func genCaseID() string {
	currentTime := time.Now()
	year := currentTime.Format("06")                                 // "25" for 2025
	month := fmt.Sprintf("%02d", int(currentTime.Month()))           // "06" for June
	day := fmt.Sprintf("%02d", currentTime.Day())                    // "10" for 10th
	hour := fmt.Sprintf("%02d", currentTime.Hour())                  // "15" for 3 PM
	minute := fmt.Sprintf("%02d", currentTime.Minute())              // "04"
	second := fmt.Sprintf("%02d", currentTime.Second())              // "05"
	millisecond := fmt.Sprintf("%07d", currentTime.Nanosecond()/1e3) // "1234567" (nanoseconds → microseconds)

	// Combine into DYYMMDDHHMMSSNNNNNNN format
	timestamp := fmt.Sprintf("D%s%s%s%s%s%s%s",
		year,
		month,
		day,
		hour,
		minute,
		second,
		millisecond,
	)
	return timestamp
}

// @summary List Cases
// @tags Cases
// @security ApiKeyAuth
// @id ListCase
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @Param start_date query string false "start_date"
// @Param end_date query string false "end_date"
// @Param caseType query string false "caseType"
// @Param caseSType query string false "caseSType"
// @Param detail query string false "detail"
// @Param countryId query string false "countryId"
// @Param provId query string false "provId"
// @Param distId query string false "distId"
// @Param category query string false "category"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/case [get]
func ListCase(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
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

	caseType := c.Query("caseType")
	caseSType := c.Query("caseSType")
	detail := c.Query("detail")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	category := c.Query("category")
	countryId := c.Query("countryId")
	provId := c.Query("provId")
	distId := c.Query("distId")

	// Dynamic query builder
	baseQuery := `SELECT id, "orgId", "caseId", "caseVersion", "referCaseId", "caseTypeId", "caseSTypeId", 
       priority, "wfId", source, "deviceId", "phoneNo", "phoneNoHide", "caseDetail", 
       "extReceive", "statusId", "caseLat", "caseLon", "caselocAddr", "caselocAddrDecs", 
       "countryId", "provId", "distId", "caseDuration", "createdAt", "startedDate", 
       "commandedDate", "receivedDate", "arrivedDate", "closedDate", usercreate, 
       usercommand, userreceive, userarrive, userclose, "resId", "resDetail",  
       "scheduleFlag", "scheduleDate", "updatedAt", "createdBy", "updatedBy"
	FROM public.tix_cases WHERE "orgId" = $1 `

	params := []interface{}{orgId}
	paramIndex := 2 // start at $2 because $1 is already used for orgId

	// Add conditions dynamically
	if caseType != "" {
		baseQuery += fmt.Sprintf(" AND \"caseTypeId\" = $%d", paramIndex)
		params = append(params, caseType)
		paramIndex++
	}

	if caseSType != "" {
		baseQuery += fmt.Sprintf(" AND \"caseSTypeId\" = $%d", paramIndex)
		params = append(params, caseSType)
		paramIndex++
	}

	if detail != "" {
		baseQuery += fmt.Sprintf(" AND \"caseDetail\" ILIKE $%d", paramIndex)
		params = append(params, "%"+detail+"%")
		paramIndex++
	}

	if startDate != "" {
		baseQuery += fmt.Sprintf(" AND \"createdAt\" >= $%d", paramIndex)
		params = append(params, startDate)
		paramIndex++
	}

	if endDate != "" {
		baseQuery += fmt.Sprintf(" AND \"createdAt\" <= $%d", paramIndex)
		params = append(params, endDate)
		paramIndex++
	}

	if category != "" {
		baseQuery += fmt.Sprintf(" AND \"statusId\" = $%d", paramIndex)
		params = append(params, category)
		paramIndex++
	}

	// New filters
	if countryId != "" {
		baseQuery += fmt.Sprintf(" AND \"countryId\" = $%d", paramIndex)
		params = append(params, countryId)
		paramIndex++
	}

	if provId != "" {
		baseQuery += fmt.Sprintf(" AND \"provId\" = $%d", paramIndex)
		params = append(params, provId)
		paramIndex++
	}

	if distId != "" {
		baseQuery += fmt.Sprintf(" AND \"distId\" = $%d", paramIndex)
		params = append(params, distId)
		paramIndex++
	}

	// Add pagination
	baseQuery += fmt.Sprintf(" ORDER BY priority ASC, \"createdAt\" DESC LIMIT $%d OFFSET $%d", paramIndex, paramIndex+1)
	params = append(params, length, start)

	rows, err := conn.Query(ctx, baseQuery, params...)
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

	var caseLists []model.Case
	found := false
	for rows.Next() {
		var cusCase model.Case
		err := rows.Scan(
			&cusCase.ID,
			&cusCase.OrgID,
			&cusCase.CaseID,
			&cusCase.CaseVersion,
			&cusCase.ReferCaseID,
			&cusCase.CaseTypeID,
			&cusCase.CaseSTypeID,
			&cusCase.Priority,
			&cusCase.WfID,
			&cusCase.Source,
			&cusCase.DeviceID,
			&cusCase.PhoneNo,
			&cusCase.PhoneNoHide,
			&cusCase.CaseDetail,
			&cusCase.ExtReceive,
			&cusCase.StatusID,
			&cusCase.CaseLat,
			&cusCase.CaseLon,
			&cusCase.CaseLocAddr,
			&cusCase.CaseLocAddrDecs,
			&cusCase.CountryID,
			&cusCase.ProvID,
			&cusCase.DistID,
			&cusCase.CaseDuration,
			&cusCase.CreatedAt, // "createdAt"
			&cusCase.StartedDate,
			&cusCase.CommandedDate,
			&cusCase.ReceivedDate,
			&cusCase.ArrivedDate,
			&cusCase.ClosedDate,
			&cusCase.UserCreate,
			&cusCase.UserCommand,
			&cusCase.UserReceive,
			&cusCase.UserArrive,
			&cusCase.UserClose,
			&cusCase.ResID,
			&cusCase.ResDetail,
			&cusCase.ScheduleFlag,
			&cusCase.ScheduleDate,
			&cusCase.UpdatedAt, // "updatedAt"
			&cusCase.CreatedBy, // "createdBy"
			&cusCase.UpdatedBy, // "updatedBy"
		)

		if err != nil {
			logger.Warn("Query failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.Response{
				Status: "-1",
				Msg:    "Failed",
				Desc:   err.Error(),
			})
			return
		}

		caseLists = append(caseLists, cusCase)
		found = true
	}

	if !found {
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   "Not found",
		})
		return
	}

	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   caseLists,
	}
	c.JSON(http.StatusOK, response)

	paramQuery := c.Request.URL.RawQuery
	logStr := Process("ListCase", paramQuery, response.Status, paramQuery, response)
	logger.Info(logStr)
}

// @summary Cases By Id
// @tags Cases
// @security ApiKeyAuth
// @id Case By Id
// @accept json
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/case/{id} [get]
func CaseById(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	orgId := GetVariableFromToken(c, "orgId")
	id := c.Param("id")

	query := `SELECT id, "orgId", "caseId", "caseVersion", "referCaseId", "caseTypeId", "caseSTypeId", priority, "wfId", "versions", source, "deviceId", "phoneNo", "phoneNoHide", "caseDetail", "extReceive", "statusId", "caseLat", "caseLon", "caselocAddr", "caselocAddrDecs", "countryId", "provId", "distId", "caseDuration", "createdDate", "startedDate", "commandedDate", "receivedDate", "arrivedDate", "closedDate", usercreate, usercommand, userreceive, userarrive, userclose, "resId", "resDetail", "ScheduleFlag", "scheduleDate", "createdAt", "updatedAt", "createdBy", "updatedBy"
	FROM public.tix_cases WHERE "orgId"=$1 AND id=$2`
	logger.Debug(`Query`, zap.String("query", query))
	var cusCase model.Case
	err := conn.QueryRow(ctx, query, orgId, id).Scan(
		&cusCase.ID,
		&cusCase.OrgID,
		&cusCase.CaseID,
		&cusCase.CaseVersion,
		&cusCase.ReferCaseID,
		&cusCase.CaseTypeID,
		&cusCase.CaseSTypeID,
		&cusCase.Priority,
		&cusCase.WfID,
		&cusCase.WfVersions,
		&cusCase.Source,
		&cusCase.DeviceID,
		&cusCase.PhoneNo,
		&cusCase.PhoneNoHide,
		&cusCase.CaseDetail,
		&cusCase.ExtReceive,
		&cusCase.StatusID,
		&cusCase.CaseLat,
		&cusCase.CaseLon,
		&cusCase.CaseLocAddr,
		&cusCase.CaseLocAddrDecs,
		&cusCase.CountryID,
		&cusCase.ProvID,
		&cusCase.DistID,
		&cusCase.CaseDuration,
		&cusCase.CreatedDate,
		&cusCase.StartedDate,
		&cusCase.CommandedDate,
		&cusCase.ReceivedDate,
		&cusCase.ArrivedDate,
		&cusCase.ClosedDate,
		&cusCase.UserCreate,
		&cusCase.UserCommand,
		&cusCase.UserReceive,
		&cusCase.UserArrive,
		&cusCase.UserClose,
		&cusCase.ResID,
		&cusCase.ResDetail,
		&cusCase.ScheduleFlag,
		&cusCase.ScheduleDate,
		&cusCase.CreatedAt,
		&cusCase.UpdatedAt,
		&cusCase.CreatedBy,
		&cusCase.UpdatedBy,
	)
	if err != nil {
		logger.Warn("Query failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		return
	}

	// Final JSON
	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   cusCase,
		Desc:   "",
	}
	c.JSON(http.StatusOK, response)

	paramQuery := c.Request.URL.RawQuery
	logStr := Process("ListCase", paramQuery, response.Status, paramQuery, response)
	logger.Info(logStr)
}

// @summary Create Case
// @id Create Case
// @security ApiKeyAuth
// @tags Cases
// @accept json
// @produce json
// @param Body body model.CaseInsert true "Create Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/case/add [post]
func InsertCase(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	var req model.CaseInsert
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
	var caseId string
	if req.CaseId == nil || *req.CaseId == "" || *req.CaseId == "null" {
		caseId = genCaseID()
	} else {
		caseId = *req.CaseId
	}
	var id int
	orgId := GetVariableFromToken(c, "orgId")
	query := `
	INSERT INTO public."tix_cases"(
	"orgId", "caseId", "caseVersion", "referCaseId", "caseTypeId", "caseSTypeId", priority, "wfId", "versions",source, "deviceId",
	"phoneNo", "phoneNoHide", "caseDetail", "extReceive", "statusId", "caseLat", "caseLon", "caselocAddr",
	"caselocAddrDecs", "countryId", "provId", "distId", "caseDuration", "createdDate", "startedDate",
	"commandedDate", "receivedDate", "arrivedDate", "closedDate", usercreate, usercommand, userreceive,
	userarrive, userclose, "resId", "resDetail", "scheduleFlag", "scheduleDate", "createdAt", "updatedAt", "createdBy", "updatedBy")
	VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
		$11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
		$21, $22, $23, $24, $25, $26, $27, $28, $29, $30,
		$31, $32, $33, $34, $35, $36, $37, $38, $39 , $40, $41, $42, $43
	) RETURNING id ;
	`

	logger.Debug(`Query`, zap.String("query", query), zap.Any("req", req))
	err := conn.QueryRow(ctx, query,
		orgId, caseId, req.CaseVersion, req.ReferCaseID, req.CaseTypeID, req.CaseSTypeID, req.Priority, req.WfID, req.WfVersions,
		req.Source, req.DeviceID, req.PhoneNo, req.PhoneNoHide, req.CaseDetail, req.ExtReceive, req.StatusID,
		req.CaseLat, req.CaseLon, req.CaseLocAddr, req.CaseLocAddrDecs, req.CountryID, req.ProvID, req.DistID,
		req.CaseDuration, req.CreatedDate, req.StartedDate, req.CommandedDate, req.ReceivedDate, req.ArrivedDate,
		req.ClosedDate, req.UserCreate, req.UserCommand, req.UserReceive, req.UserArrive, req.UserClose, req.ResID,
		req.ResDetail, req.ScheduleFlag, req.ScheduleDate, now, now, username, username).Scan(&id)

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

	fmt.Printf("=======CurrentStage========")
	fmt.Printf("%s", req.NodeID)
	if req.NodeID != "" {
		var data = model.CustomCaseCurrentStage{
			CaseID:   caseId,
			WfID:     req.WfID,
			NodeID:   req.NodeID,
			StatusID: req.StatusID,
		}
		fmt.Printf("=======yyy========")
		err = CaseCurrentStageInsert(conn, ctx, c, data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, model.Response{
				Status: "-1",
				Msg:    "Failure",
				Desc:   err.Error(),
			})
			return
		}
	}
	fmt.Printf("=======AnswerForm========")
	if req.FormData.FormId == "" {

	} else {
		err = InsertFormAnswer(conn, ctx, orgId.(string), caseId, *req.FormData, username.(string))
		if err != nil {
			log.Fatal("Insert error:", err)
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, model.Response{
				Status: "-1",
				Msg:    "Failure",
				Desc:   err.Error(),
			})
			return
		}
	}

	//Noti Custom
	data := []model.Data{
		{Key: "delay", Value: "0"}, //0=white, 1=yellow , 2=red
	}
	recipients := []model.Recipient{
		{Type: "provId", Value: req.ProvID},
	}
	genNotiCustom(c, orgId.(string), username.(string), username.(string), "/case/"+caseId, "Create", data, "สร้าง Case สำเร็จ : "+caseId, recipients, "", "User")

	c.JSON(http.StatusOK, model.ResponseCreateCase{
		Status: "0",
		Msg:    "Success",
		Desc:   "Create successfully",
		CaseID: caseId,
	})

}

// @summary Update Case
// @id Update Case
// @security ApiKeyAuth
// @accept json
// @produce json
// @tags Cases
// @Param id path int true "id"
// @param Body body model.CaseUpdate true "Update data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/case/{id} [patch]
func UpdateCase(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	id := c.Param("id")

	var req model.CaseUpdate
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
	query := `UPDATE public."tix_cases"
	SET "caseVersion"=$3, "referCaseId"=$4, "caseTypeId"=$5, "caseSTypeId"=$6,
	 priority=$7, source=$8, "deviceId"=$9, "phoneNo"=$10, "phoneNoHide"=$11, "caseDetail"=$12, "extReceive"=$13,
	  "statusId"=$14, "caseLat"=$15, "caseLon"=$16, "caselocAddr"=$17, "caselocAddrDecs"=$18, "countryId"=$19,
	   "provId"=$20, "distId"=$21, "caseDuration"=$22, "createdDate"=$23, "startedDate"=$24, "commandedDate"=$25,
	    "receivedDate"=$26, "arrivedDate"=$27, "closedDate"=$28, usercreate=$29, usercommand=$30, userreceive=$31,
		 userarrive=$32, userclose=$33, "resId"=$34, "resDetail"=$35, "scheduleFlag"=$36 , "scheduleDate"=$37, "updatedAt"=$38,"updatedBy"=$39 ,"wfId"=$40
	WHERE id = $1 AND "orgId"=$2`
	_, err := conn.Exec(ctx, query,
		id, orgId, req.CaseVersion, req.ReferCaseID, req.CaseTypeID, req.CaseSTypeID, req.Priority,
		req.Source, req.DeviceID, req.PhoneNo, req.PhoneNoHide, req.CaseDetail, req.ExtReceive, req.StatusID,
		req.CaseLat, req.CaseLon, req.CaseLocAddr, req.CaseLocAddrDecs, req.CountryID, req.ProvID, req.DistID,
		req.CaseDuration, req.CreatedDate, req.StartedDate, req.CommandedDate, req.ReceivedDate, req.ArrivedDate,
		req.ClosedDate, req.UserCreate, req.UserCommand, req.UserReceive, req.UserArrive, req.UserClose, req.ResID,
		req.ResDetail, req.ScheduleFlag, req.ScheduleDate, now, username, req.WfID)
	logger.Debug("Update Case SQL Args",
		zap.String("query", query),
		zap.Any("Input", []any{
			id, orgId, req.CaseVersion, req.ReferCaseID, req.CaseTypeID, req.CaseSTypeID, req.Priority,
			req.Source, req.DeviceID, req.PhoneNo, req.PhoneNoHide, req.CaseDetail, req.ExtReceive, req.StatusID,
			req.CaseLat, req.CaseLon, req.CaseLocAddr, req.CaseLocAddrDecs, req.CountryID, req.ProvID, req.DistID,
			req.CaseDuration, req.CreatedDate, req.StartedDate, req.CommandedDate, req.ReceivedDate, req.ArrivedDate,
			req.ClosedDate, req.UserCreate, req.UserCommand, req.UserReceive, req.UserArrive, req.UserClose, req.ResID,
			req.ResDetail, req.ScheduleFlag, req.ScheduleDate, now, username, req.WfID,
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

// @summary Delete Case
// @id Delete Case
// @security ApiKeyAuth
// @accept json
// @tags Cases
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/case/{id} [delete]
func DeleteCase(c *gin.Context) {

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
	query := `DELETE FROM public."tix_cases" WHERE id = $1 AND "orgId"=$2`
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

// @summary List Cases Type with Sub type
// @tags Cases
// @security ApiKeyAuth
// @id List Cases Type with Sub type
// @accept json
// @produce json
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/casetypes_with_subtype [get]
func ListCaseTypeWithSubtype(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	orgId := GetVariableFromToken(c, "orgId")

	query := `SELECT t1."typeId",t1."orgId",t1."en",t1."th",t1."active",t2."sTypeId",t2."sTypeCode",
	t2."en",t2."th",t2."wfId", t2."caseSla", t2.priority, t2."userSkillList", t2."unitPropLists", t2.active
	FROM public.case_types t1
	FULL JOIN public.case_sub_types t2
	ON t1."typeId" = t2."typeId"
	WHERE t1."orgId"=$1`
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

	var caseLists []model.CaseTypeWithSubType
	var errorMsg string
	for rows.Next() {
		var cusCase model.CaseTypeWithSubType
		err := rows.Scan(
			&cusCase.TypeID, &cusCase.OrgID, &cusCase.TypeEN, &cusCase.TypeTH, &cusCase.TypeActive,
			&cusCase.SubTypeID, &cusCase.SubTypeCode, &cusCase.SubTypeEN, &cusCase.SubTypeTH,
			&cusCase.WfID, &cusCase.CaseSla, &cusCase.Priority,
			&cusCase.UserSkillList, &cusCase.UnitPropLists, &cusCase.SubTypeActive,
		)
		if err != nil {
			logger.Warn("Query failed", zap.Error(err))
			errorMsg = err.Error()
			continue
		}

		caseLists = append(caseLists, cusCase)
	}

	// Final JSON
	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   caseLists,
		Desc:   errorMsg,
	}
	c.JSON(http.StatusOK, response)

	paramQuery := c.Request.URL.RawQuery
	logStr := Process("ListCase", paramQuery, response.Status, paramQuery, response)
	logger.Info(logStr)
}

// @summary List Cases
// @tags Cases
// @security ApiKeyAuth
// @id ListCaseTypes
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/casetypes [get]
func ListCaseType(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
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
	query := `SELECT id,"typeId", "orgId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy" 
	FROM public.case_types WHERE "orgId"=$1 LIMIT $2 OFFSET $3`
	logger.Debug(`Query`, zap.String("query", query))

	rows, err := conn.Query(ctx, query, orgId, length, start)
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

	var caseLists []model.CaseType
	var errorMsg string
	for rows.Next() {
		var cusCase model.CaseType
		err := rows.Scan(&cusCase.Id, &cusCase.TypeId, &cusCase.OrgId, &cusCase.En, &cusCase.Th, &cusCase.Active, &cusCase.CreatedAt,
			&cusCase.UpdatedAt, &cusCase.CreatedBy, &cusCase.UpdatedBy)
		if err != nil {
			logger.Warn("Query failed", zap.Error(err))
			errorMsg = err.Error()
			continue
		}

		caseLists = append(caseLists, cusCase)
	}

	// Final JSON
	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   caseLists,
		Desc:   errorMsg,
	}
	c.JSON(http.StatusOK, response)

	paramQuery := c.Request.URL.RawQuery
	logStr := Process("ListCase", paramQuery, response.Status, paramQuery, response)
	logger.Info(logStr)
}

// @summary Create CaseType
// @id Create CaseType
// @security ApiKeyAuth
// @tags Cases
// @accept json
// @produce json
// @param Body body model.CaseTypeInsert true "Create Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/casetypes/add [post]
func InsertCaseType(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	var req model.CaseTypeInsert
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
	orgId := GetVariableFromToken(c, "orgId")
	uuid := uuid.New()
	query := `
	INSERT INTO public."case_types"(
	"typeId", "orgId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy")
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id ;
	`

	err := conn.QueryRow(ctx, query,
		uuid, orgId, req.En, req.Th, req.Active, now,
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

// @summary Update CaseType
// @id Update CaseType
// @security ApiKeyAuth
// @accept json
// @produce json
// @tags Cases
// @Param id path int true "id"
// @param Body body model.CaseTypeUpdate true "Update data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/casetypes/{id} [patch]
func UpdateCaseType(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	id := c.Param("id")

	var req model.CaseTypeUpdate
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
	query := `UPDATE public."case_types"
	SET en=$2, th=$3, active=$4,
	 "updatedAt"=$5, "updatedBy"=$6
	WHERE id = $1 AND "orgId"=$7`
	_, err := conn.Exec(ctx, query,
		id, req.En, req.Th, req.Active,
		now, username, orgId,
	)
	logger.Debug("Update Case SQL Args",
		zap.String("query", query),
		zap.Any("Input", []any{
			id, req.En, req.Th, req.Active,
			now, username, orgId,
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

// @summary Delete CaseType
// @id Delete CaseType
// @security ApiKeyAuth
// @accept json
// @tags Cases
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/casetypes/{id} [delete]
func DeleteCaseType(c *gin.Context) {

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
	query := `DELETE FROM public."case_types" WHERE id = $1 AND "orgId"=$2`
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

// @summary List CasesSubType
// @tags Cases
// @security ApiKeyAuth
// @id ListCaseSubTypes
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/casesubtypes [get]
func ListCaseSubType(c *gin.Context) {
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
	orgId := GetVariableFromToken(c, "orgId")
	query := `SELECT id, "typeId", "sTypeId", "sTypeCode", "orgId", en, th, "wfId", "caseSla", priority, "userSkillList", "unitPropLists",
	 active, "createdAt", "updatedAt", "createdBy", "updatedBy" FROM public.case_sub_types WHERE "orgId"=$1 LIMIT $2 OFFSET $3`
	logger.Debug(`Query`, zap.String("query", query))

	rows, err := conn.Query(ctx, query, orgId, length, start)
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

	var caseLists []model.CaseSubType
	var errorMsg string
	for rows.Next() {
		var cusCase model.CaseSubType
		err := rows.Scan(&cusCase.Id, &cusCase.TypeID, &cusCase.STypeID, &cusCase.STypeCode, &cusCase.OrgID, &cusCase.EN, &cusCase.TH, &cusCase.WFID,
			&cusCase.CaseSLA, &cusCase.Priority, &cusCase.UserSkillList, &cusCase.UnitPropLists, &cusCase.Active,
			&cusCase.CreatedAt, &cusCase.UpdatedAt, &cusCase.CreatedBy, &cusCase.UpdatedBy)
		if err != nil {
			logger.Warn("Query failed", zap.Error(err))
			errorMsg = err.Error()
			continue
		}

		caseLists = append(caseLists, cusCase)
	}

	// Final JSON
	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   caseLists,
		Desc:   errorMsg,
	}
	c.JSON(http.StatusOK, response)

	paramQuery := c.Request.URL.RawQuery
	logStr := Process("ListCase", paramQuery, response.Status, paramQuery, response)
	logger.Info(logStr)
}

// @summary Create CaseSubType
// @id Create CaseSubType
// @security ApiKeyAuth
// @tags Cases
// @accept json
// @produce json
// @param Body body model.CaseSubTypeInsert true "Create Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/casesubtypes/add [post]
func InsertCaseSubType(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	var req model.CaseSubTypeInsert
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
	orgId := GetVariableFromToken(c, "orgId")
	uuid := uuid.New()
	query := `
	INSERT INTO public."case_sub_types"(
	"typeId", "sTypeId", "sTypeCode", "orgId", en, th, "wfId", "caseSla", priority,
	 "userSkillList", "unitPropLists", active, "createdAt", "updatedAt", "createdBy", "updatedBy")
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	RETURNING id ;
	`

	err := conn.QueryRow(ctx, query, req.TypeID, uuid, req.STypeCode, orgId, req.EN, req.TH,
		req.WFID, req.CaseSLA, req.Priority, req.UserSkillList, req.UnitPropLists, req.Active, now,
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

// @summary Update CaseSubType
// @id Update CaseSubType
// @security ApiKeyAuth
// @accept json
// @produce json
// @tags Cases
// @Param id path int true "id"
// @param Body body model.CaseSubTypeUpdate true "Update data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/casesubtypes/{id} [patch]
func UpdateCaseSubType(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	id := c.Param("id")

	var req model.CaseSubTypeUpdate
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
	query := `UPDATE public."case_sub_types"
	SET "sTypeCode"=$3, en=$4, th=$5, "wfId"=$6, "caseSla"=$7,
	 priority=$8, "userSkillList"=$9, "unitPropLists"=$10, active=$11, "updatedAt"=$12,
	  "updatedBy"=$13
	WHERE id = $1 AND "orgId"=$2`
	_, err := conn.Exec(ctx, query,
		id, orgId, req.STypeCode, req.EN, req.TH, req.WFID, req.CaseSLA, req.Priority, req.UserSkillList, req.UnitPropLists, req.Active,
		now, username,
	)
	logger.Debug("Update Case SQL Args",
		zap.String("query", query),
		zap.Any("Input", []any{
			id, orgId, req.STypeCode, req.EN, req.TH, req.WFID, req.CaseSLA, req.Priority, req.UserSkillList, req.UnitPropLists, req.Active,
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

// @summary Delete CaseSubType
// @id Delete CaseSubType
// @security ApiKeyAuth
// @accept json
// @tags Cases
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/casesubtypes/{id} [delete]
func DeleteCaseSubType(c *gin.Context) {

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
	query := `DELETE FROM public."case_sub_types" WHERE id = $1 AND "orgId"=$2`
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
