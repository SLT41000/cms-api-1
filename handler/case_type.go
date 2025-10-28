package handler

import (
	"mainPackage/model"
	"mainPackage/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// @summary List Cases Type with Sub type
// @tags Case-Types and Case-SubTypes
// @security ApiKeyAuth
// @id List Cases Type with Sub type
// @accept json
// @produce json
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/casetypes_with_subtype [get]
func ListCaseTypeWithSubtype(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()

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
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Cases", "ListCaseTypeWithSubType", "",
			"search", -1, start_time, GetQueryParams(c), response, "Query failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
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
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, id, "Cases", "ListCaseTypeWithSubType", "",
		"search", 0, start_time, GetQueryParams(c), response, "GetListCaseTypeWitjSubType Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)

	paramQuery := c.Request.URL.RawQuery
	logStr := Process("ListCase", paramQuery, response.Status, paramQuery, response)
	logger.Info(logStr)
}

// @summary List Cases Type
// @tags Case-Types and Case-SubTypes
// @security ApiKeyAuth
// @id ListCaseTypes
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/casetypes [get]
func ListCaseType(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
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
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Cases", "ListCaseType", "",
			"search", -1, start_time, GetQueryParams(c), response, "Failure : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
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
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, id, "Cases", "ListCaseType", "",
		"search", 0, start_time, GetQueryParams(c), response, "Get ListCaseType Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)

	paramQuery := c.Request.URL.RawQuery
	logStr := Process("ListCase", paramQuery, response.Status, paramQuery, response)
	logger.Info(logStr)
}

// @summary Create CaseType
// @id Create CaseType
// @security ApiKeyAuth
// @tags Case-Types and Case-SubTypes
// @accept json
// @produce json
// @param Body body model.CaseTypeInsert true "Create Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/casetypes/add [post]
func InsertCaseType(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	uuid := uuid.New().String()
	var req model.CaseTypeInsert
	if err := c.ShouldBindJSON(&req); err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			uuid, uuid, "Cases", "InsertCaseType", "",
			"create", -1, start_time, req, response, "Failure : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}
	now := time.Now()
	var id int

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
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			uuid, uuid, "Cases", "InsertCaseType", "",
			"create", -1, start_time, req, response, "Failure : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}
	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Create successfully",
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		uuid, uuid, "Cases", "InsertCaseType", "",
		"create", 0, start_time, req, response, "Insert Case Success.",
	)
	//=======AUDIT_END=====//
	// Continue logic...
	c.JSON(http.StatusOK, response)

}

// @summary Update CaseType
// @id Update CaseType
// @security ApiKeyAuth
// @accept json
// @produce json
// @tags Case-Types and Case-SubTypes
// @Param id path string true "id"
// @param Body body model.CaseTypeUpdate true "Update data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/casetypes/{id} [patch]
func UpdateCaseType(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	id := c.Param("id")
	now := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()

	var req model.CaseTypeUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Update failed", zap.Error(err))
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Cases", "UpdateCaseType", "",
			"update", -1, now, req, response, "Failure : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		return
	}

	query := `UPDATE public."case_types"
	SET en=$2, th=$3, active=$4,
	 "updatedAt"=$5, "updatedBy"=$6
	WHERE "typeId" = $1 AND "orgId"=$7`
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
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Cases", "UpdateCaseType", "",
			"update", -1, now, req, response, "Failure : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Update failed", zap.Error(err))
		return
	}
	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Update successfully",
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, id, "Cases", "UpdateCaseType", "",
		"update", 0, now, req, response, "UpdateCaseType Success.",
	)
	//=======AUDIT_END=====//
	// Continue logic...
	c.JSON(http.StatusOK, response)
}

// @summary Delete CaseType
// @id Delete CaseType
// @security ApiKeyAuth
// @accept json
// @tags Case-Types and Case-SubTypes
// @produce json
// @Param id path string true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/casetypes/{id} [delete]
func DeleteCaseType(c *gin.Context) {

	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()
	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	query := `DELETE FROM public."case_types" WHERE "typeId" = $1 AND "orgId"=$2`
	logger.Debug("Query", zap.String("query", query), zap.Any("id", id))
	_, err := conn.Exec(ctx, query, id, orgId)
	if err != nil {
		// log.Printf("Insert failed: %v", err)
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Cases", "DeleteCaseType", "",
			"delete", -1, start_time, GetQueryParams(c), response, "Failure : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Update failed", zap.Error(err))
		return
	}
	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Delete successfully",
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, id, "Cases", "DeleteCaseType", "",
		"delete", 0, start_time, GetQueryParams(c), response, "Delete CaseType Success.",
	)
	//=======AUDIT_END=====//
	// Continue logic...
	c.JSON(http.StatusOK, response)
}

// @summary List CasesSubType
// @tags Case-Types and Case-SubTypes
// @security ApiKeyAuth
// @id ListCaseSubTypes
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/casesubtypes [get]
func ListCaseSubType(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
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

	query := `SELECT id, "typeId", "sTypeId", "sTypeCode", "orgId", en, th, "wfId", "caseSla", priority, "userSkillList", "unitPropLists",
	 active, "createdAt", "updatedAt", "createdBy", "updatedBy", "mDeviceType", "mWorkOrderType" FROM public.case_sub_types WHERE "orgId"=$1 ORDER BY "sTypeCode" ASC  LIMIT $2 OFFSET $3`
	logger.Debug(`Query`, zap.String("query", query))

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
			txtId, id, "Cases", "ListCaseSubType", "",
			"search", -1, start_time, GetQueryParams(c), response, "Failure : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()

	var caseLists []model.CaseSubType
	var errorMsg string
	for rows.Next() {
		var cusCase model.CaseSubType
		err := rows.Scan(&cusCase.Id, &cusCase.TypeID, &cusCase.STypeID, &cusCase.STypeCode, &cusCase.OrgID, &cusCase.EN, &cusCase.TH, &cusCase.WFID,
			&cusCase.CaseSLA, &cusCase.Priority, &cusCase.UserSkillList, &cusCase.UnitPropLists, &cusCase.Active,
			&cusCase.CreatedAt, &cusCase.UpdatedAt, &cusCase.CreatedBy, &cusCase.UpdatedBy, &cusCase.MDeviceType, &cusCase.MWorkOrderType)
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
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, id, "Cases", "ListCaseSubType", "",
		"search", 0, start_time, GetQueryParams(c), response, "Get ListCaseSubType Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)

	paramQuery := c.Request.URL.RawQuery
	logStr := Process("ListCase", paramQuery, response.Status, paramQuery, response)
	logger.Info(logStr)
}

// @summary Create CaseSubType
// @id Create CaseSubType
// @security ApiKeyAuth
// @tags Case-Types and Case-SubTypes
// @accept json
// @produce json
// @param Body body model.CaseSubTypeInsert true "Create Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/casesubtypes/add [post]
func InsertCaseSubType(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()
	now := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	uuid := uuid.New().String()
	var req model.CaseSubTypeInsert
	if err := c.ShouldBindJSON(&req); err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			uuid, uuid, "Cases", "InsertCaseTypeSubType", "",
			"create", -1, now, req, response, "Failure : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}
	var id int

	query := `
	INSERT INTO public."case_sub_types"(
	"typeId", "sTypeId", "sTypeCode", "orgId", en, th, "wfId", "caseSla", priority,
	 "userSkillList", "unitPropLists", active, "createdAt", "updatedAt", "createdBy", "updatedBy", "mDeviceType", "mWorkOrderType")
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	RETURNING id ;
	`

	err := conn.QueryRow(ctx, query, req.TypeID, uuid, req.STypeCode, orgId, req.EN, req.TH,
		req.WFID, req.CaseSLA, req.Priority, req.UserSkillList, req.UnitPropLists, req.Active, now,
		now, username, username, req.MDeviceType, req.MWorkOrderType).Scan(&id)

	if err != nil {
		// log.Printf("Insert failed: %v", err)
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			uuid, uuid, "Cases", "InsertCaseTypeSubType", "",
			"create", -1, now, req, response, "Failure : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}
	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Create successfully",
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		uuid, uuid, "Cases", "InsertCaseTypeSubType", "",
		"create", 0, now, req, response, "InsertCaseSubAType Success/",
	)
	//=======AUDIT_END=====//
	// Continue logic...
	c.JSON(http.StatusOK, response)

}

// @summary Update CaseSubType
// @id Update CaseSubType
// @security ApiKeyAuth
// @accept json
// @produce json
// @tags Case-Types and Case-SubTypes
// @Param id path string true "id"
// @param Body body model.CaseSubTypeUpdate true "Update data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/casesubtypes/{id} [patch]
func UpdateCaseSubType(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()

	var req model.CaseSubTypeUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Update failed", zap.Error(err))
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Cases", "UpdateCaseSubType", "",
			"update", -1, start_time, req, response, "Failure : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		return
	}
	now := time.Now()
	query := `UPDATE public."case_sub_types"
	SET "sTypeCode"=$3, en=$4, th=$5, "wfId"=$6, "caseSla"=$7,
	 priority=$8, "userSkillList"=$9, "unitPropLists"=$10, active=$11, "updatedAt"=$12,
	  "updatedBy"=$13, "mDeviceType"=$14 , "mWorkOrderType"=$15, "typeId"=$16
	WHERE "sTypeId" = $1 AND "orgId"=$2`
	_, err := conn.Exec(ctx, query,
		id, orgId, req.STypeCode, req.EN, req.TH, req.WFID, req.CaseSLA, req.Priority, req.UserSkillList, req.UnitPropLists, req.Active,
		now, username, req.MDeviceType, req.MWorkOrderType, req.TypeID,
	)
	logger.Debug("Update Case SQL Args",
		zap.String("query", query),
		zap.Any("Input", []any{
			id, orgId, req.STypeCode, req.EN, req.TH, req.WFID, req.CaseSLA, req.Priority, req.UserSkillList, req.UnitPropLists, req.Active,
			now, username,
		}))
	if err != nil {
		// log.Printf("Insert failed: %v", err)
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Cases", "UpdateCaseSubType", "",
			"update", -1, start_time, req, response, "Failure : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Update failed", zap.Error(err))
		return
	}
	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Update successfully",
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, id, "Cases", "UpdateCaseSubType", "",
		"update", 0, start_time, req, response, "UpdateCaseSubType Success.",
	)
	//=======AUDIT_END=====//
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
// @tags Case-Types and Case-SubTypes
// @produce json
// @Param id path string true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/casesubtypes/{id} [delete]
func DeleteCaseSubType(c *gin.Context) {

	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()
	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	query := `DELETE FROM public."case_sub_types" WHERE "sTypeId" = $1 AND "orgId"=$2`
	logger.Debug("Query", zap.String("query", query), zap.Any("id", id))
	_, err := conn.Exec(ctx, query, id, orgId)
	if err != nil {
		// log.Printf("Insert failed: %v", err)
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Cases", "DeleteCaseSubType", "",
			"delete", -1, start_time, GetQueryParams(c), response, "Failure : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Update failed", zap.Error(err))
		return
	}
	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Delete successfully",
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, id, "Cases", "DeleteCaseSubType", "",
		"delete", 0, start_time, GetQueryParams(c), response, "DeleteCaseSubType success.",
	)
	//=======AUDIT_END=====//
	// Continue logic...
	c.JSON(http.StatusOK, response)
}
