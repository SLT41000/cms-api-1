package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
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

// @summary Get Form
// @tags Form and Workflow
// @security ApiKeyAuth
// @id Get Form
// @accept json
// @produce json
// @Param id query string true "id"
// @Param version query string true "version"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/forms [get]
func GetForm(c *gin.Context) {
	logger := utils.GetLog()
	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	version := c.Query("version")
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	query := `SELECT form_builder."formId",form_builder."formName",form_builder."formColSpan",form_elements."eleData" 
	FROM public.form_builder INNER JOIN public.form_elements ON form_builder."formId"=form_elements."formId" 
	WHERE form_builder."formId" = $1 AND form_builder."orgId"=$2 AND form_elements."versions"=$3`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	logger.Debug(`id :` + id)
	rows, err := conn.Query(ctx, query, id, orgId, version)
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
			txtId, id, "Form", "GetForm", "",
			"search", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()
	var errorMsg string
	var formFields []model.IndividualFormField
	var form model.Form
	for rows.Next() {
		var rawJSON []byte
		err := rows.Scan(&form.FormId, &form.FormName, &form.FormColSpan, &rawJSON)
		if err != nil {
			logger.Warn("Scan failed", zap.Error(err))
			continue
		}

		var field model.IndividualFormField
		if err := json.Unmarshal(rawJSON, &field); err != nil {
			logger.Warn("unmarshal field failed", zap.Error(err))
			continue
		}

		formFields = append(formFields, field)
		form.FormFieldJson = formFields
	}
	if errorMsg != "" {
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Data:   form,
			Desc:   errorMsg,
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Form", "GetForm", "",
			"search", -1, start_time, GetQueryParams(c), response, "Failed : "+errorMsg,
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
	} else {
		response := model.Response{
			Status: "0",
			Msg:    "Success",
			Data:   form,
			Desc:   "",
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Form", "GetForm", "",
			"search", 0, start_time, GetQueryParams(c), response, "GetForm Success",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusOK, response)
	}
}

// @summary Get All Form
// @tags Form and Workflow
// @security ApiKeyAuth
// @id Get All Form
// @accept json
// @produce json
// @response 200 {object} model.ResponseDataFormList
// @Router /api/v1/forms/getAllForms [get]
func GetAllForm(c *gin.Context) {
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

	if orgId == "" {
		response := model.Response{
			Status: "-1",
			Msg:    "Invalid token",
			Desc:   "orgId not found in token",
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Form", "GetAllForm", "",
			"search", -1, start_time, GetQueryParams(c), response, "Invalid Token",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		return
	}
	query := `SELECT DISTINCT ON (fb."formId") 
    fb."formId", fe."versions", fb."active", fb."publish", fb."formName",
    fb."locks", fe."eleData", fe."createdBy", fe."createdAt", fe."updatedAt", fe."updatedBy"
	FROM public.form_builder AS fb 
	INNER JOIN public.form_elements AS fe ON fb."formId" = fe."formId"
	WHERE fb."orgId" = $1
	ORDER BY fb."formId" ASC, CAST(fe."versions" AS INTEGER) DESC;`

	logger.Debug("Query", zap.String("query", query))

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
			txtId, id, "Form", "GetAllForm", "",
			"search", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()

	var forms []model.FormsManager

	for rows.Next() {
		var rawJSON []byte
		var form model.FormsManager

		err := rows.Scan(
			&form.FormId,
			&form.Versions,
			&form.Active,
			&form.Publish,
			&form.FormName,
			&form.Locks,
			&rawJSON,
			&form.CreatedBy,
			&form.CreatedAt,
			&form.UpdatedAt,
			&form.UpdatedBy,
		)

		if err != nil {
			logger.Warn("Scan failed", zap.Error(err))
			continue
		}

		var eleData model.Form
		if err := json.Unmarshal(rawJSON, &eleData); err != nil {
			logger.Warn("Unmarshal eleData failed", zap.Error(err))
			continue
		}

		form.FormFieldJson = eleData.FormFieldJson
		form.FormColSpan = eleData.FormColSpan
		forms = append(forms, form)
		// logger.Debug("Row data", zap.Any("form", form))
	}

	if len(forms) == 0 {
		response := model.Response{
			Status: "-1",
			Msg:    "No data found",
			Data:   nil,
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Form", "GetAllForm", "",
			"search", -1, start_time, GetQueryParams(c), response, "Not Found",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusNotFound, response)
		return
	}

	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   forms,
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, id, "Form", "GetAllForm", "",
		"search", 0, start_time, GetQueryParams(c), response, "GetAllForm Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}

// old_db_stuct
// func ___GetAllForm(c *gin.Context) {
// 	logger := utils.GetLog()
// 	conn, ctx, cancel := utils.ConnectDB()
// 	if conn == nil {
// 		return
// 	}
// 	defer cancel()
// 	defer conn.Close(ctx)

// 	id := c.Param("id")
// 	start_time := time.Now()
// 	username := GetVariableFromToken(c, "username")
// 	orgId := GetVariableFromToken(c, "orgId")
// 	txtId := uuid.New().String()

// 	if orgId == "" {
// 		response := model.Response{
// 			Status: "-1",
// 			Msg:    "Invalid token",
// 			Desc:   "orgId not found in token",
// 		}
// 		//=======AUDIT_START=====//
// 		_ = utils.InsertAuditLogs(
// 			c, conn, orgId.(string), username.(string),
// 			txtId, id, "Form", "GetAllForm", "",
// 			"search", -1, start_time, GetQueryParams(c), response, "Invalid Token",
// 		)
// 		//=======AUDIT_END=====//
// 		c.JSON(http.StatusBadRequest, response)
// 		return
// 	}
// 	query := `SELECT form_builder."formId", form_builder."versions",form_builder."active",form_builder."publish",
//     form_builder."formName",form_builder."locks", form_builder."formColSpan", form_elements."eleData" , form_elements."createdBy"
//               ,form_elements."createdAt", form_elements."updatedAt",  form_elements."updatedBy"
//               FROM public.form_builder
//               INNER JOIN public.form_elements
//               ON form_builder."formId" = form_elements."formId"
//               WHERE form_builder."orgId" = $1
//               ORDER BY public.form_elements."eleNumber" ASC`

// 	logger.Debug("Query", zap.String("query", query))

// 	rows, err := conn.Query(ctx, query, orgId)
// 	if err != nil {
// 		logger.Warn("Query failed", zap.Error(err))
// 		response := model.Response{
// 			Status: "-1",
// 			Msg:    "Failure",
// 			Desc:   err.Error(),
// 		}
// 		//=======AUDIT_START=====//
// 		_ = utils.InsertAuditLogs(
// 			c, conn, orgId.(string), username.(string),
// 			txtId, id, "Form", "GetAllForm", "",
// 			"search", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
// 		)
// 		//=======AUDIT_END=====//
// 		c.JSON(http.StatusInternalServerError, response)
// 		return
// 	}
// 	defer rows.Close()

// 	var forms []model.FormsManager

// 	for rows.Next() {
// 		var rawJSON []byte
// 		var form model.FormsManager

// 		err := rows.Scan(
// 			&form.FormId,
// 			&form.Versions,
// 			&form.Active,
// 			&form.Publish,
// 			&form.FormName,
// 			&form.Locks,
// 			&form.FormColSpan,
// 			&rawJSON,
// 			&form.CreatedBy,
// 			&form.CreatedAt,
// 			&form.UpdatedAt,
// 			&form.UpdatedBy,
// 		)

// 		if err != nil {
// 			logger.Warn("Scan failed", zap.Error(err))
// 			continue
// 		}

// 		var field model.IndividualFormField
// 		if err := json.Unmarshal(rawJSON, &field); err != nil {
// 			logger.Warn("Unmarshal field failed", zap.Error(err))
// 			continue
// 		}

// 		form.FormFieldJson = []model.IndividualFormField{field}
// 		forms = append(forms, form)
// 		// logger.Debug("Row data", zap.Any("form", form))
// 	}

// 	var result []model.FormsManager
// 	formMap := make(map[string]int)
// 	for _, form := range forms {
// 		if idx, exists := formMap[*form.FormName]; exists {
// 			result[idx].FormFieldJson = append(result[idx].FormFieldJson, form.FormFieldJson...)
// 		} else {
// 			result = append(result, form)
// 			formMap[*form.FormName] = len(result) - 1
// 			logger.Debug("DEBUG", zap.Any("formMap", formMap))
// 		}
// 	}

// 	if len(forms) == 0 {
// 		response := model.Response{
// 			Status: "-1",
// 			Msg:    "No data found",
// 			Data:   nil,
// 		}
// 		//=======AUDIT_START=====//
// 		_ = utils.InsertAuditLogs(
// 			c, conn, orgId.(string), username.(string),
// 			txtId, id, "Form", "GetAllForm", "",
// 			"search", -1, start_time, GetQueryParams(c), response, "Not Found",
// 		)
// 		//=======AUDIT_END=====//
// 		c.JSON(http.StatusNotFound, response)
// 		return
// 	}

// 	response := model.Response{
// 		Status: "0",
// 		Msg:    "Success",
// 		Data:   result,
// 	}
// 	//=======AUDIT_START=====//
// 	_ = utils.InsertAuditLogs(
// 		c, conn, orgId.(string), username.(string),
// 		txtId, id, "Form", "GetAllForm", "",
// 		"search", 0, start_time, GetQueryParams(c), response, "GetAllForm Success.",
// 	)
// 	//=======AUDIT_END=====//
// 	c.JSON(http.StatusOK, response)
// }

// @summary Create Form
// @tags Form and Workflow
// @security ApiKeyAuth
// @id Create Form
// @accept json
// @produce json
// @param Case body model.FormInsert true "Created Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/forms [post]
// @summary Create Form
// @tags Form and Workflow
// @security ApiKeyAuth
// @id Create Form
// @accept json
// @produce json
// @param Case body model.FormInsert true "Created Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/forms [post]
func FormInsert(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()

	var req model.FormInsert
	if err := c.ShouldBindJSON(&req); err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, "", "Form", "FormInsert", "",
			"create", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	uuid := uuid.New()
	now := time.Now()
	var id int
	var exists bool
	uuidStr := uuid.String()
	checkQuery := `
	SELECT EXISTS (
		SELECT 1 FROM public."form_builder" 
		WHERE "orgId" = $1 AND "formName" = $2
	);
`
	err := conn.QueryRow(ctx, checkQuery, orgId, req.FormName).Scan(&exists)
	if err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, "", "Form", "FormInsert", "",
			"create", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}
	if exists {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   "form name already exists",
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, "", "Form", "FormInsert", "",
			"create", -1, start_time, GetQueryParams(c), response, "form name already exists.",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Insert failed form name already exists")
		return
	}
	query := `
	INSERT INTO public."form_builder"(
	"orgId", "formId", "formName", active, publish, versions, locks, "createdAt", "updatedAt", "createdBy", "updatedBy")
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	RETURNING id ;
	`
	err = conn.QueryRow(ctx, query,
		orgId, uuid, req.FormName, req.Active, req.Publish, "1", req.Locks, now,
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
			txtId, "", "Form", "FormInsert", "",
			"create", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	form := model.Form{
		FormName:      req.FormName,
		FormId:        &uuidStr,
		FormColSpan:   req.FormColSpan,
		FormFieldJson: req.FormFieldJson,
	}

	err = InsertFormElement(conn, ctx, form, orgId.(string), username.(string), "1", uuid.String(), nil, nil)

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
			txtId, uuid.String(), "Form", "FormInsert", "",
			"create", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
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
		Data: map[string]interface{}{
			"formId":  uuid,
			"version": 1,
		},
	}

	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, uuid.String(), "Form", "FormInsert", "",
		"create", 0, start_time, GetQueryParams(c), response, "FormInsert Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}

// old db stuct
// func ___FormInsert(c *gin.Context) {
// 	logger := utils.GetLog()
// 	conn, ctx, cancel := utils.ConnectDB()
// 	if conn == nil {
// 		return
// 	}
// 	defer cancel()
// 	defer conn.Close(ctx)

// 	start_time := time.Now()
// 	username := GetVariableFromToken(c, "username")
// 	orgId := GetVariableFromToken(c, "orgId")
// 	txtId := uuid.New().String()

// 	var req model.FormInsert
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		response := model.Response{
// 			Status: "-1",
// 			Msg:    "Failure",
// 			Desc:   err.Error(),
// 		}
// 		//=======AUDIT_START=====//
// 		_ = utils.InsertAuditLogs(
// 			c, conn, orgId.(string), username.(string),
// 			txtId, "", "Form", "FormInsert", "",
// 			"create", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
// 		)
// 		//=======AUDIT_END=====//
// 		c.JSON(http.StatusBadRequest, response)
// 		logger.Warn("Insert failed", zap.Error(err))
// 		return
// 	}

// 	uuid := uuid.New()
// 	now := time.Now()
// 	var id int
// 	var exists bool
// 	checkQuery := `
// 	SELECT EXISTS (
// 		SELECT 1 FROM public."form_builder"
// 		WHERE "orgId" = $1 AND "formName" = $2
// 	);
// `
// 	err := conn.QueryRow(ctx, checkQuery, orgId, req.FormName).Scan(&exists)
// 	if err != nil {
// 		response := model.Response{
// 			Status: "-1",
// 			Msg:    "Failure",
// 			Desc:   err.Error(),
// 		}
// 		//=======AUDIT_START=====//
// 		_ = utils.InsertAuditLogs(
// 			c, conn, orgId.(string), username.(string),
// 			txtId, "", "Form", "FormInsert", "",
// 			"create", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
// 		)
// 		//=======AUDIT_END=====//
// 		c.JSON(http.StatusInternalServerError, response)
// 		logger.Warn("Insert failed", zap.Error(err))
// 		return
// 	}
// 	if exists {
// 		response := model.Response{
// 			Status: "-1",
// 			Msg:    "Failure",
// 			Desc:   "form name already exists",
// 		}
// 		//=======AUDIT_START=====//
// 		_ = utils.InsertAuditLogs(
// 			c, conn, orgId.(string), username.(string),
// 			txtId, "", "Form", "FormInsert", "",
// 			"create", -1, start_time, GetQueryParams(c), response, "form name already exists.",
// 		)
// 		//=======AUDIT_END=====//
// 		c.JSON(http.StatusInternalServerError, response)
// 		logger.Warn("Insert failed form name already exists")
// 		return
// 	}
// 	query := `
// 	INSERT INTO public."form_builder"(
// 	"orgId", "formId", "formName", "formColSpan", active, publish, versions, locks, "createdAt", "updatedAt", "createdBy", "updatedBy")
// 	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
// 	RETURNING id ;
// 	`
// 	err = conn.QueryRow(ctx, query,
// 		orgId, uuid, req.FormName, req.FormColSpan, req.Active, req.Publish, "draft", req.Locks, now,
// 		now, username, username).Scan(&id)

// 	if err != nil {
// 		// log.Printf("Insert failed: %v", err)
// 		response := model.Response{
// 			Status: "-1",
// 			Msg:    "Failure",
// 			Desc:   err.Error(),
// 		}
// 		//=======AUDIT_START=====//
// 		_ = utils.InsertAuditLogs(
// 			c, conn, orgId.(string), username.(string),
// 			txtId, "", "Form", "FormInsert", "",
// 			"create", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
// 		)
// 		//=======AUDIT_END=====//
// 		c.JSON(http.StatusInternalServerError, response)
// 		logger.Warn("Insert failed", zap.Error(err))
// 		return
// 	}

// 	for i, item := range req.FormFieldJson {
// 		logger.Debug("eleNumber", zap.Int("i", i+1))
// 		logger.Debug("JsonArray", zap.Any("Json", item))
// 		query := `
// 				INSERT INTO public.form_elements(
// 				"orgId", "formId", versions, "eleNumber", "eleData", "createdAt", "updatedAt", "createdBy", "updatedBy")
// 				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
// 				RETURNING id ;
// 				`
// 		err = conn.QueryRow(ctx, query,
// 			orgId, uuid, "draft", i+1, item, now, now, username, username).Scan(&id)

// 		if err != nil {
// 			// log.Printf("Insert failed: %v", err)
// 			response := model.Response{
// 				Status: "-1",
// 				Msg:    "Failure",
// 				Desc:   err.Error(),
// 			}
// 			//=======AUDIT_START=====//
// 			_ = utils.InsertAuditLogs(
// 				c, conn, orgId.(string), username.(string),
// 				txtId, uuid.String(), "Form", "FormInsert", "",
// 				"create", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
// 			)
// 			//=======AUDIT_END=====//
// 			c.JSON(http.StatusInternalServerError, response)
// 			logger.Warn("Insert failed", zap.Error(err))
// 			return
// 		}
// 	}

// 	response := model.Response{
// 		Status: "0",
// 		Msg:    "Success",
// 		Desc:   "Create successfully",
// 		Data:   uuid,
// 	}

// 	//=======AUDIT_START=====//
// 	_ = utils.InsertAuditLogs(
// 		c, conn, orgId.(string), username.(string),
// 		txtId, uuid.String(), "Form", "FormInsert", "",
// 		"create", 0, start_time, GetQueryParams(c), response, "FormInsert Success.",
// 	)
// 	//=======AUDIT_END=====//
// 	c.JSON(http.StatusOK, response)
// }

// @summary Update Form
// @tags Form and Workflow
// @security ApiKeyAuth
// @id Update Form
// @accept json
// @produce json
// @Param uuid path string true "uuid"
// @param Case body model.FormUpdate true "Update Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/forms/{uuid} [patch]
func FormUpdate(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	uuid := c.Param("uuid")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")

	var req model.FormUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			uuid, "", "Form", "FormUpdate", "",
			"Update", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		c.JSON(http.StatusBadRequest, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	now := time.Now()
	var id int
	var formElementsLastVersion int
	var currentVersion string

	err := conn.QueryRow(ctx, `
		SELECT versions FROM public.form_builder
		WHERE "formId"=$1 AND "orgId"=$2
	`, uuid, orgId).Scan(&currentVersion)

	println("form current Version :" + currentVersion)

	if err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			uuid, strconv.Itoa(id), "Form", "FormUpdate", "",
			"Update", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	var createBy string
	var createAt time.Time

	err = conn.QueryRow(ctx, `
		SELECT CAST(versions AS INTEGER) AS max_version,
			"createdAt", "createdBy"
		FROM public.form_elements
		WHERE "formId" = $1 AND "orgId" = $2
		ORDER BY CAST(versions AS INTEGER) DESC
		LIMIT 1
	`, uuid, orgId).Scan(&formElementsLastVersion, &createAt, &createBy)

	if err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			uuid, strconv.Itoa(id), "Form", "FormUpdate", "",
			"Update", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	// Update main form
	query := `
	UPDATE public.form_builder
	SET "formName"=$2, active=$3, publish=$4,
	    locks=$5, "updatedAt"=$6, "updatedBy"=$7
	WHERE "formId"=$1 AND "orgId"=$8;
	`
	_, err = conn.Exec(ctx, query, uuid,
		req.FormName, req.Active, req.Publish, req.Locks,
		now, username, orgId)

	if err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			uuid, "", "Form", "FormUpdate", "",
			"Update", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	// Convert version string to int
	currentVersionInt, _ := strconv.Atoi(currentVersion)

	if currentVersionInt == formElementsLastVersion {
		form := model.Form{
			FormName:      req.FormName,
			FormId:        &uuid,
			FormColSpan:   req.FormColSpan,
			FormFieldJson: req.FormFieldJson,
		}
		newVersion := strconv.Itoa(currentVersionInt + 1)
		err = InsertFormElement(conn, ctx, form, orgId.(string), username.(string), newVersion, uuid, &createAt, &createBy)
	} else {
		form := model.Form{
			FormName:      req.FormName,
			FormColSpan:   req.FormColSpan,
			FormFieldJson: req.FormFieldJson,
			FormId:        &uuid,
		}

		eleData, _ := json.Marshal(form)
		updateEleQuery := `
			UPDATE public.form_elements
			SET "eleData"=$3, "updatedAt"=$4, "updatedBy"=$5
			WHERE "formId"=$1 AND "orgId"=$2 AND versions=$6;
		`
		_, err = conn.Exec(ctx, updateEleQuery, uuid, orgId, eleData, now, username, strconv.Itoa(formElementsLastVersion))
	}

	if err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			uuid, strconv.Itoa(id), "Form", "FormUpdate", "",
			"Update", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Update successfully",
	}
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		uuid, strconv.Itoa(id), "Form", "FormUpdate", "",
		"Update", 0, start_time, GetQueryParams(c), response, "FormUpdate Success.",
	)
	c.JSON(http.StatusOK, response)
}

//old db stuct
// func FormUpdate(c *gin.Context) {
// 	logger := utils.GetLog()
// 	conn, ctx, cancel := utils.ConnectDB()
// 	if conn == nil {
// 		return
// 	}
// 	defer cancel()
// 	defer conn.Close(ctx)
// 	uuid := c.Param("uuid")
// 	start_time := time.Now()
// 	username := GetVariableFromToken(c, "username")
// 	orgId := GetVariableFromToken(c, "orgId")

// 	var req model.FormUpdate
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		response := model.Response{
// 			Status: "-1",
// 			Msg:    "Failure",
// 			Desc:   err.Error(),
// 		}
// 		//=======AUDIT_START=====//
// 		_ = utils.InsertAuditLogs(
// 			c, conn, orgId.(string), username.(string),
// 			uuid, "", "Form", "FormUpdate", "",
// 			"Update", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
// 		)
// 		//=======AUDIT_END=====//
// 		c.JSON(http.StatusBadRequest, response)
// 		logger.Warn("Insert failed", zap.Error(err))
// 		return
// 	}

// 	now := time.Now()
// 	var id int

// 	query := `
// 	UPDATE public.form_builder
// 	SET "formName"=$2, "formColSpan"=$3, active=$4, publish=$5,
// 	 versions=$6, locks=$7, "updatedAt"=$8,"updatedBy"=$9
// 	WHERE "formId"=$1 AND "orgId"=$10;
// 	`
// 	_, err := conn.Exec(ctx, query, uuid,
// 		req.FormName, req.FormColSpan, req.Active, req.Publish, "draft", req.Locks,
// 		now, username, orgId)

// 	if err != nil {
// 		// log.Printf("Insert failed: %v", err)
// 		response := model.Response{
// 			Status: "-1",
// 			Msg:    "Failure",
// 			Desc:   err.Error(),
// 		}
// 		//=======AUDIT_START=====//
// 		_ = utils.InsertAuditLogs(
// 			c, conn, orgId.(string), username.(string),
// 			uuid, "", "Form", "FormUpdate", "",
// 			"Update", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
// 		)
// 		//=======AUDIT_END=====//
// 		c.JSON(http.StatusInternalServerError, response)
// 		logger.Warn("Insert failed", zap.Error(err))
// 		return
// 	}

// 	query = `
// 	DELETE FROM public.form_elements
// 	WHERE "formId"=$1;
// 	`
// 	_, err = conn.Exec(ctx, query, uuid)

// 	if err != nil {
// 		// log.Printf("Insert failed: %v", err)
// 		response := model.Response{
// 			Status: "-1",
// 			Msg:    "Failure",
// 			Desc:   err.Error(),
// 		}
// 		//=======AUDIT_START=====//
// 		_ = utils.InsertAuditLogs(
// 			c, conn, orgId.(string), username.(string),
// 			uuid, strconv.Itoa(id), "Form", "FormUpdate", "",
// 			"Update", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
// 		)
// 		//=======AUDIT_END=====//
// 		c.JSON(http.StatusInternalServerError, response)
// 		logger.Warn("Insert failed", zap.Error(err))
// 		return
// 	}

// 	for i, item := range req.FormFieldJson {
// 		now := time.Now()
// 		logger.Debug("eleNumber", zap.Int("i", i+1))
// 		logger.Debug("JsonArray", zap.Any("Json", item))
// 		query := `
// 				INSERT INTO public.form_elements(
// 				"orgId", "formId", versions, "eleNumber", "eleData", "createdAt", "updatedAt", "createdBy", "updatedBy")
// 				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
// 				RETURNING id ;
// 				`
// 		err = conn.QueryRow(ctx, query,
// 			orgId, uuid, "draft", i+1, item, now, now, username, username).Scan(&id)

// 		if err != nil {
// 			// log.Printf("Insert failed: %v", err)
// 			response := model.Response{
// 				Status: "-1",
// 				Msg:    "Failure",
// 				Desc:   err.Error(),
// 			}
// 			//=======AUDIT_START=====//
// 			_ = utils.InsertAuditLogs(
// 				c, conn, orgId.(string), username.(string),
// 				uuid, strconv.Itoa(id), "Form", "FormUpdate", "",
// 				"Update", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
// 			)
// 			//=======AUDIT_END=====//
// 			c.JSON(http.StatusInternalServerError, response)
// 			logger.Warn("Insert failed", zap.Error(err))
// 			return
// 		}
// 	}

// 	response := model.Response{
// 		Status: "0",
// 		Msg:    "Success",
// 		Desc:   "Update successfully",
// 	}
// 	//=======AUDIT_START=====//
// 	_ = utils.InsertAuditLogs(
// 		c, conn, orgId.(string), username.(string),
// 		uuid, strconv.Itoa(id), "Form", "FormUpdate", "",
// 		"Update", 0, start_time, GetQueryParams(c), response, "FormUpdate Success.",
// 	)
// 	//=======AUDIT_END=====//
// 	c.JSON(http.StatusOK, response)
// }

// @summary Update Form Publish
// @tags Form and Workflow
// @security ApiKeyAuth
// @id Update Form Publish
// @accept json
// @produce json
// @param Case body model.FormPublish true "Update Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/forms/publish [patch]
func FormPublish(c *gin.Context) {

	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	var req model.FormPublish
	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	var formElementsLastVersion int
	if err := c.ShouldBindJSON(&req); err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Form", "FormPubilsh", "",
			"update", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	now := time.Now()
	err := conn.QueryRow(ctx, `
		SELECT CAST(versions AS INTEGER) AS max_version
		FROM public.form_elements
		WHERE "formId" = $1 AND "orgId" = $2
		ORDER BY CAST(versions AS INTEGER) DESC
		LIMIT 1
	`, req.FormID, orgId).Scan(&formElementsLastVersion)

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
			txtId, id, "Form", "FormPubilsh", "",
			"update", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	query := `
	UPDATE public.form_builder
	SET publish=$2, "updatedAt"=$3,"updatedBy"=$4, versions =$6
	WHERE "formId"=$1 AND "orgId"=$5;
	`
	_, err = conn.Exec(ctx, query, req.FormID, req.Publish, now, username, orgId, strconv.Itoa(formElementsLastVersion))
	logger.Debug("Update Case SQL Args",
		zap.String("query", query),
		zap.Any("Input", []any{
			req.FormID, req.Publish, now, username, orgId,
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
			txtId, id, "Form", "FormPubilsh", "",
			"update", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Insert failed", zap.Error(err))
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
		txtId, id, "Form", "FormPubilsh", "",
		"update", -1, start_time, GetQueryParams(c), response, "Update FormPublish Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}

// @summary Update Form Lock
// @tags Form and Workflow
// @security ApiKeyAuth
// @id Update Form Lock
// @accept json
// @produce json
// @param Case body model.FormLock true "Update Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/forms/lock [patch]
func FormLock(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	var req model.FormLock
	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	if err := c.ShouldBindJSON(&req); err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Form", "FormLock", "",
			"update", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	now := time.Now()

	query := `
	UPDATE public.form_builder
	SET locks=$2, "updatedAt"=$3,"updatedBy"=$4
	WHERE "formId"=$1 AND "orgId"=$5;
	`
	_, err := conn.Exec(ctx, query, req.FormID, req.Locks, now, username, orgId)
	logger.Debug("Update Case SQL Args",
		zap.String("query", query),
		zap.Any("Input", []any{
			req.FormID, req.Locks, now, username, orgId,
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
			txtId, id, "Form", "FormLock", "",
			"update", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Insert failed", zap.Error(err))
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
		txtId, id, "Form", "FormLock", "",
		"update", 0, start_time, GetQueryParams(c), response, "Update FormLock Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}

// @summary Update Form Status
// @tags Form and Workflow
// @security ApiKeyAuth
// @id Update Form Status
// @accept json
// @produce json
// @param Case body model.FormActive true "Update Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/forms/active [patch]
func FormActive(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	var req model.FormActive
	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	if err := c.ShouldBindJSON(&req); err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Form", "FormActive", "",
			"update", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	now := time.Now()

	query := `
	UPDATE public.form_builder
	SET active=$2, "updatedAt"=$3,"updatedBy"=$4
	WHERE "formId"=$1 AND "orgId"=$5;
	`
	_, err := conn.Exec(ctx, query, req.FormID, req.Active, now, username, orgId)
	logger.Debug("Update Case SQL Args",
		zap.String("query", query),
		zap.Any("Input", []any{
			req.FormID, req.Active, now, username, orgId,
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
			txtId, id, "Form", "FormActive", "",
			"update", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Insert failed", zap.Error(err))
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
		txtId, id, "Form", "FormActive", "",
		"update", 0, start_time, GetQueryParams(c), response, "Update FormActive Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}

// @summary Delete Form
// @id Delete Form
// @security ApiKeyAuth
// @accept json
// @tags Form and Workflow
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/forms/{id} [delete]
func DeleteForm(c *gin.Context) {}

// @summary Get Workflow
// @tags Form and Workflow
// @security ApiKeyAuth
// @id Get Workflow
// @accept json
// @produce json
// @Param id path string true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/workflows/{id} [get]
func GetWorkFlow(c *gin.Context) {
	logger := utils.GetLog()
	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()

	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	query := `SELECT "section","data",title,"desc",wf_definitions."versions",wf_definitions."createdAt",wf_definitions."updatedAt" 
	FROM public.wf_definitions Inner join public.wf_nodes
	ON wf_definitions."wfId" = wf_nodes."wfId" WHERE wf_definitions."wfId" = $1 AND wf_nodes."orgId"=$2`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	logger.Debug(`id :` + id)
	rows, err := conn.Query(ctx, query, id, orgId)
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
			txtId, id, "WorkFlow", "GetWorkFlow", "",
			"search", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()
	var errorMsg string
	var NodesArray []map[string]interface{}
	var ConnectionArray []map[string]interface{}
	var workflow model.WorkFlow
	var workflowMetaData model.WorkFlowMetadata
	rowIndex := 0
	for rows.Next() {
		rowIndex++
		var rawJSON []byte
		var rowsType string
		err := rows.Scan(&rowsType, &rawJSON, &workflowMetaData.Title, &workflowMetaData.Desc,
			&workflowMetaData.Status, &workflowMetaData.CreatedAt, &workflowMetaData.UpdatedAt)
		if err != nil {
			logger.Warn("Scan failed", zap.Error(err))
			response := model.Response{
				Status: "-1",
				Msg:    "Failed",
				Data:   workflow,
				Desc:   errorMsg,
			}
			//=======AUDIT_START=====//
			_ = utils.InsertAuditLogs(
				c, conn, orgId.(string), username.(string),
				txtId, id, "WorkFlow", "GetWorkFlow", "",
				"search", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
			)
			//=======AUDIT_END=====//
			c.JSON(http.StatusInternalServerError, response)
		}

		switch rowsType {
		case "nodes":
			field, err := unmarshalToMap(rawJSON)
			if err != nil {
				logger.Warn("Unmarshal nodes failed", zap.Error(err))
				continue
			}
			NodesArray = append(NodesArray, field)
		case "connections":
			fields, err := unmarshalToSliceOfMaps(rawJSON)
			if err != nil {
				logger.Warn("Unmarshal connections failed", zap.Error(err))
				continue
			}
			ConnectionArray = append(ConnectionArray, fields...)
		default:
			logger.Warn("Unknown rowsType", zap.String("rowsType", rowsType))
			continue
		}
		workflow.Nodes = NodesArray
		workflow.Connections = ConnectionArray
		workflow.MetaData = workflowMetaData
	}
	if errorMsg != "" {
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Data:   workflow,
			Desc:   errorMsg,
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "WorkFlow", "GetWorkFlow", "",
			"search", -1, start_time, GetQueryParams(c), response, "Failed : "+errorMsg,
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
	} else {
		response := model.Response{
			Status: "0",
			Msg:    "Success",
			Data:   workflow,
			Desc:   "",
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "WorkFlow", "GetWorkFlow", "",
			"search", 0, start_time, GetQueryParams(c), response, "GetWorkFlow Success",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusOK, response)
	}
}

// @summary Get Workflow List
// @tags Form and Workflow
// @security ApiKeyAuth
// @id Get Workflow List
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/workflows [get]
func GetWorkFlowList(c *gin.Context) {
	logger := utils.GetLog()

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
	if err != nil || length > 100 {
		length = 1000
	}
	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	query := `SELECT id, "orgId", "wfId", title, "desc", active, publish, locks, versions, "createdAt", "updatedAt", "createdBy", "updatedBy"
	FROM public.wf_definitions WHERE "orgId"=$1 LIMIT $2 OFFSET $3`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query),
		zap.Any("Input", []any{orgId, length, start}))
	rows, err = conn.Query(ctx, query, orgId, length, start)
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
			txtId, id, "WorkFlow", "GetWorkFlowList", "",
			"search", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()
	var Workflow model.WorkflowModel
	var Data []model.WorkflowModel
	found := false
	for rows.Next() {
		err := rows.Scan(&Workflow.ID, &Workflow.OrgID, &Workflow.WfID, &Workflow.Title, &Workflow.Desc, &Workflow.Active,
			&Workflow.Publish, &Workflow.Locks, &Workflow.Versions, &Workflow.CreatedAt, &Workflow.UpdatedAt, &Workflow.CreatedBy, &Workflow.UpdatedBy)
		if err != nil {
			logger.Warn("Scan failed", zap.Error(err))
			response := model.Response{
				Status: "-1",
				Msg:    "Failed",
				Data:   nil,
				Desc:   err.Error(),
			}
			//=======AUDIT_START=====//
			_ = utils.InsertAuditLogs(
				c, conn, orgId.(string), username.(string),
				txtId, id, "WorkFlow", "GetWorkFlowList", "",
				"search", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
			)
			//=======AUDIT_END=====//
			c.JSON(http.StatusInternalServerError, response)
			return
		}
		// Store metadata (assuming same metadata per workflowId)
		Data = append(Data, Workflow)
		found = true
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
			txtId, id, "WorkFlow", "GetWorkFlowList", "",
			"search", -1, start_time, GetQueryParams(c), response, "Not Found.",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   Data,
		Desc:   "",
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, id, "WorkFlow", "GetWorkFlowList", "",
		"search", 0, start_time, GetQueryParams(c), response, "GetWorkFlowList Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)

}

func GetWorkFlowListOld(c *gin.Context) {
	logger := utils.GetLog()

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
	if err != nil || length > 100 {
		length = 1000
	}
	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	query := `SELECT t1."id",t1."wfId","section","data",title,"desc",t1."versions",t1."createdAt",t1."updatedAt" 
	FROM public.wf_definitions t1
	Inner join public.wf_nodes t2
	ON t1."wfId" = t2."wfId" WHERE t2."orgId"=$1 LIMIT $2 OFFSET $3`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query),
		zap.Any("Input", []any{orgId, length, start}))
	rows, err = conn.Query(ctx, query, orgId, length, start)
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
			txtId, id, "WorkFlow", "GetWorkFlowListOld", "",
			"search", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()
	var errorMsg string
	var WfId string
	var workflowList []model.WorkFlow
	var workflowMetaData model.WorkFlowMetadata
	rowIndex := 0

	// Temporary maps to store nodes/connections grouped by WfId
	workflowNodesMap := make(map[string][]map[string]interface{})
	workflowConnectionsMap := make(map[string][]map[string]interface{})
	workflowMetaMap := make(map[string]model.WorkFlowMetadata)
	found := false
	for rows.Next() {
		rowIndex++
		var rawJSON []byte
		var rowsType string
		err := rows.Scan(&workflowMetaData.Id, &WfId, &rowsType, &rawJSON, &workflowMetaData.Title, &workflowMetaData.Desc,
			&workflowMetaData.Status, &workflowMetaData.CreatedAt, &workflowMetaData.UpdatedAt)
		if err != nil {
			logger.Warn("Scan failed", zap.Error(err))
			response := model.Response{
				Status: "-1",
				Msg:    "Failed",
				Data:   nil,
				Desc:   errorMsg,
			}
			//=======AUDIT_START=====//
			_ = utils.InsertAuditLogs(
				c, conn, orgId.(string), username.(string),
				txtId, id, "WorkFlow", "GetWorkFlowListOld", "",
				"search", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
			)
			//=======AUDIT_END=====//
			c.JSON(http.StatusInternalServerError, response)
			return
		}
		switch rowsType {
		case "nodes":
			field, err := unmarshalToMap(rawJSON)
			if err != nil {
				logger.Warn("Unmarshal nodes failed", zap.Error(err))
				continue
			}
			workflowNodesMap[WfId] = append(workflowNodesMap[WfId], field)
		case "connections":
			fields, err := unmarshalToSliceOfMaps(rawJSON)
			if err != nil {
				logger.Warn("Unmarshal connections failed", zap.Error(err))
				continue
			}
			workflowConnectionsMap[WfId] = append(workflowConnectionsMap[WfId], fields...)
		default:
			logger.Warn("Unknown rowsType", zap.String("rowsType", rowsType))
			continue
		}

		// Store metadata (assuming same metadata per workflowId)
		workflowMetaMap[WfId] = workflowMetaData
		found = true
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
			txtId, id, "WorkFlow", "GetWorkFlowListOld", "",
			"search", -1, start_time, GetQueryParams(c), response, "Not Found.",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	uniqueWfIDs := make(map[string]bool)
	for id := range workflowNodesMap {
		uniqueWfIDs[id] = true
	}
	for id := range workflowConnectionsMap {
		uniqueWfIDs[id] = true
	}
	for id := range workflowMetaMap {
		uniqueWfIDs[id] = true
	}
	for wfId := range uniqueWfIDs {
		workflow := model.WorkFlow{
			// Nodes:       workflowNodesMap[wfId],
			// Connections: workflowConnectionsMap[wfId],
			MetaData: workflowMetaMap[wfId],
		}
		workflowList = append(workflowList, workflow)

	}

	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   workflowList,
		Desc:   "",
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, id, "WorkFlow", "GetWorkFlowListOld", "",
		"search", 0, start_time, GetQueryParams(c), response, "GetWorkFlowListOld Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)

}

// @summary Create Workflow
// @tags Form and Workflow
// @security ApiKeyAuth
// @id Create Workflow
// @accept json
// @produce json
// @param Body body model.WorkFlowInsert true "Create Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/workflows [post]
func WorkFlowInsert(c *gin.Context) {
	logger := utils.GetLog()

	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	var req model.WorkFlowInsert
	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	if err := c.ShouldBindJSON(&req); err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "WorkFlow", "WorkFlowInsert", "",
			"create", 0, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}
	uuid := uuid.New()
	now := time.Now()
	query := `INSERT INTO public.wf_definitions(
	"orgId", "wfId", title, "desc", active, publish, locks, versions, "createdAt", "updatedAt", "createdBy", "updatedBy")
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	logger.Debug(`Query`, zap.String("query", query), zap.Any("req", req))
	_, err := conn.Exec(ctx, query, orgId, uuid, req.MetaData.Title, req.MetaData.Desc,
		true, true, true, "draft", now, now, username, username)
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
			txtId, id, "WorkFlow", "WorkFlowInsert", "",
			"create", 0, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	for i, item := range req.Nodes {
		now = time.Now()
		logger.Debug("eleNumber", zap.Int("i", i+1))
		logger.Debug("JsonArray", zap.Any("Json", item))
		var pic, group, formID interface{}
		if item.Data != nil && item.Data.Config != nil {
			config := *item.Data.Config

			if val, ok := config["pic"].(string); ok {
				pic = val
			}
			if val, ok := config["group"].(string); ok {
				group = val
			}
			if val, ok := config["formId"].(string); ok {
				formID = val
			}
		}

		query := `
		INSERT INTO public.wf_nodes(
	"orgId", "wfId", "nodeId", versions, type, section, data,pic,"group","formId", "createdAt", "updatedAt", "createdBy", "updatedBy")
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14);
		`
		logger.Debug(`Query`, zap.String("query", query),
			zap.Any("Input", []any{
				req, item,
			}))

		_, err := conn.Exec(ctx, query,
			orgId, uuid, item.Id, "draft", item.Type, "nodes", item, pic, group, formID, now, now, username, username)

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
				txtId, id, "WorkFlow", "WorkFlowInsert", "",
				"create", 0, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
			)
			//=======AUDIT_END=====//
			c.JSON(http.StatusInternalServerError, response)
			logger.Warn("Insert failed", zap.Error(err))
			return
		}
	}

	query = `
		INSERT INTO public.wf_nodes(
	"orgId", "wfId", "nodeId", versions, type, section, data, "createdAt", "updatedAt", "createdBy", "updatedBy")
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);
		`
	logger.Debug(`Query`, zap.String("query", query),
		zap.Any("Input", []any{
			req, req.Connections,
		}))

	_, err = conn.Exec(ctx, query,
		orgId, uuid, "", "draft", "", "connections", req.Connections, now, now, username, username)

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
			txtId, id, "WorkFlow", "WorkFlowInsert", "",
			"create", 0, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Insert failed", zap.Error(err))
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
		txtId, id, "WorkFlow", "WorkFlowInsert", "",
		"create", 0, start_time, GetQueryParams(c), response, "GetWorkFlowInsert Success.",
	)
	//=======AUDIT_END=====//

	c.JSON(http.StatusOK, response)

}

// @summary Update Workflow
// @tags Form and Workflow
// @security ApiKeyAuth
// @id Update Workflow
// @accept json
// @produce json
// @Param uuid path string true "uuid"
// @param Body body model.WorkFlowInsert true "Create Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/workflows/{uuid} [patch]
func WorkFlowUpdate(c *gin.Context) {
	logger := utils.GetLog()

	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	var req model.WorkFlowInsert
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}
	uuid := c.Param("uuid")
	id := c.Param("id")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	now := time.Now()

	query := `UPDATE public.wf_definitions
	SET title=$3, "desc"=$4, active=$5, publish=$6, locks=$7, versions=$8, "updatedAt"=$9,"updatedBy"=$10
		WHERE "wfId"=$1 AND "orgId"=$2;`

	logger.Debug(`Query`, zap.String("query", query), zap.Any("req", req))
	_, err := conn.Exec(ctx, query,
		uuid, orgId, req.MetaData.Title, req.MetaData.Desc,
		true, true, true, "draft", now, username)
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
			uuid, id, "WorkFlow", "WorkFlowUpdate", "",
			"update", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	query = `
	DELETE FROM public.wf_nodes
	WHERE "wfId"=$1;
	`
	_, err = conn.Exec(ctx, query, uuid)

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
			uuid, id, "WorkFlow", "WorkFlowUpdate", "",
			"update", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	for i, item := range req.Nodes {
		now = time.Now()
		logger.Debug("eleNumber", zap.Int("i", i+1))
		logger.Debug("JsonArray", zap.Any("Json", item))
		var pic, group, formID interface{}
		if item.Data != nil && item.Data.Config != nil {
			config := *item.Data.Config

			if val, ok := config["pic"].(string); ok {
				pic = val
			}
			if val, ok := config["group"].(string); ok {
				group = val
			}
			if val, ok := config["formId"].(string); ok {
				formID = val
			}
		}

		query := `
		INSERT INTO public.wf_nodes(
	"orgId", "wfId", "nodeId", versions, type, section, data,pic,"group","formId", "createdAt", "updatedAt", "createdBy", "updatedBy")
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14);
		`
		logger.Debug(`Query`, zap.String("query", query),
			zap.Any("Input", []any{
				req, item,
			}))

		_, err := conn.Exec(ctx, query,
			orgId, uuid, item.Id, "draft", item.Type, "nodes", item, pic, group, formID, now, now, username, username)

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
	}

	query = `
		INSERT INTO public.wf_nodes(
	"orgId", "wfId", "nodeId", versions, type, section, data, "createdAt", "updatedAt", "createdBy", "updatedBy")
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);
		`
	logger.Debug(`Query`, zap.String("query", query),
		zap.Any("Input", []any{
			req, req.Connections,
		}))

	_, err = conn.Exec(ctx, query,
		orgId, uuid, "", "draft", "", "connections", req.Connections, now, now, username, username)

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
			uuid, id, "WorkFlow", "WorkFlowUpdate", "",
			"update", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Insert failed", zap.Error(err))
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
		uuid, id, "WorkFlow", "WorkFlowUpdate", "",
		"update", 0, start_time, GetQueryParams(c), response, "WorkFlowUpdate Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)

}

// @summary Delete Workflow
// @tags Form and Workflow
// @security ApiKeyAuth
// @id Delete Workflow
// @accept json
// @produce json
// @Param uuid path string true "uuid"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/workflows/{uuid} [delete]
func WorkflowDelete(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	id := c.Param("uuid")
	start_time := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	query := `DELETE FROM public."wf_definitions" WHERE "wfId" = $1 AND "orgId"=$2`
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
			txtId, id, "WorkFlow", "WorkflowDlete", "",
			"delete", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Update failed", zap.Error(err))
		return
	}
	query = `DELETE FROM public."wf_nodes" WHERE "wfId" = $1 AND "orgId"=$2`
	logger.Debug("Query", zap.String("query", query), zap.Any("id", id))
	_, err = conn.Exec(ctx, query, id, orgId)
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
			txtId, id, "WorkFlow", "WorkflowDlete", "",
			"delete", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Update failed", zap.Error(err))
		return
	}

	// Continue logic...
	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Delete successfully",
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, id, "WorkFlow", "WorkflowDlete", "",
		"delete", 0, start_time, GetQueryParams(c), response, "WorkflowDelete Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}

// @summary Get Form by Casesubtype
// @tags Form and Workflow
// @security ApiKeyAuth
// @id Get Form by Casesubtype
// @accept json
// @produce json
// @param Case body model.FormByCasesubtype true "Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/forms/casesubtype [post]
func GetFormByCaseSubType(c *gin.Context) {
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
	var req model.FormByCasesubtype
	if err := c.ShouldBindJSON(&req); err != nil {
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Form", "GetFormByCaseSubType", "",
			"search", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}
	var wfId string
	// username := GetVariableFromToken(c, "username")
	query := `SELECT "wfId" FROM public.case_sub_types WHERE "orgId"=$1 AND "sTypeId"=$2`
	logger.Debug(`Query`, zap.String("query", query),
		zap.Any("Input", []any{
			orgId, req.CaseSubType,
		}))
	err := conn.QueryRow(ctx, query, orgId, req.CaseSubType).Scan(&wfId)
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
			txtId, id, "Form", "GetFormByCaseSubType", "",
			"search", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	query = `SELECT 
    t2."section",
    t2."data",
    t2."nodeId",
    t1."versions"
FROM public.wf_definitions t1
INNER JOIN public.wf_nodes t2
    ON t1."wfId" = t2."wfId" AND t1."versions" = t2."versions"
WHERE t1."wfId" = $1
  AND t1."orgId" = $2
  AND LOWER(t2."type") = 'process'  
  AND t2."data"->'data'->'config'->>'action' = 'S001'`
	logger.Debug(`Query`, zap.String("query", query), zap.Any("Input", []any{
		wfId, orgId,
	}))

	var rows pgx.Rows
	rows, err = conn.Query(ctx, query, wfId, orgId)
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
			txtId, id, "Form", "GetFormByCaseSubType", "",
			"search", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()
	var formId string
	var nodeId string
	var versions string
	for rows.Next() {
		var rawJSON []byte
		var rowsType string
		var tempNodeId, tempVersions string

		err := rows.Scan(&rowsType, &rawJSON, &tempNodeId, &tempVersions)
		if err != nil {
			logger.Warn("Scan failed", zap.Error(err))
			continue
		}

		// Only handle nodes
		if rowsType != "nodes" {
			continue
		}

		var nodeMap map[string]interface{}
		if err := json.Unmarshal(rawJSON, &nodeMap); err != nil {
			logger.Warn("Unmarshal failed", zap.Error(err))
			continue
		}

		// Access data -> config -> formId if action = S002
		if data, ok := nodeMap["data"].(map[string]interface{}); ok {
			if config, ok := data["config"].(map[string]interface{}); ok {
				if action, ok := config["action"].(string); ok && action == "S001" {
					if formVal, ok := config["formId"].(string); ok {
						formId = formVal
						nodeId = tempNodeId
						versions = tempVersions
						break // found, exit loop
					}
				}
			}
		}
	}
	logger.Debug(formId)
	rows.Close()
	query = `SELECT t1."formId",t1."formName",t2."eleData" 
	FROM public.form_builder t1
	INNER JOIN public.form_elements t2
	ON t1."formId"=t2."formId" AND t2."versions"=t1."versions"
	WHERE t1."formId" = $1 AND t1."orgId"=$2 `

	logger.Debug(`Query`, zap.String("query", query), zap.Any("Input", []any{
		formId, orgId,
	}))
	rows, err = conn.Query(ctx, query, formId, orgId)
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
			txtId, id, "Form", "GetFormByCaseSubType", "",
			"search", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()
	var form model.FormAnswerRequest
	found := false
	for rows.Next() {
		var rawJSON []byte
		err := rows.Scan(&form.FormId, &form.FormName, &rawJSON)
		if err != nil {
			logger.Warn("Scan failed", zap.Error(err))
			response := model.Response{
				Status: "-1",
				Msg:    "Failed",
				Desc:   err.Error(),
			}
			//=======AUDIT_START=====//
			_ = utils.InsertAuditLogs(
				c, conn, orgId.(string), username.(string),
				txtId, id, "Form", "GetFormByCaseSubType", "",
				"search", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
			)
			//=======AUDIT_END=====//
			c.JSON(http.StatusOK, response)
			return
		}

		var field model.Form
		if err := json.Unmarshal(rawJSON, &field); err != nil {
			logger.Warn("unmarshal field failed", zap.Error(err))
			response := model.Response{
				Status: "-1",
				Msg:    "Failed",
				Desc:   err.Error(),
			}
			//=======AUDIT_START=====//
			_ = utils.InsertAuditLogs(
				c, conn, orgId.(string), username.(string),
				txtId, id, "Form", "GetFormByCaseSubType", "",
				"search", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
			)
			//=======AUDIT_END=====//
			c.JSON(http.StatusOK, response)
			return
		}

		form.FormColSpan = field.FormColSpan
		form.FormFieldJson = field.FormFieldJson
		found = true
	}
	if !found {
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   "No form data found",
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Form", "GetFormByCaseSubType", "",
			"search", -1, start_time, GetQueryParams(c), response, "Not Found.",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	form.NextNodeId = nodeId
	form.Versions = versions
	form.WfId = wfId
	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   form,
		Desc:   "",
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, id, "Form", "GetFormByCaseSubType", "",
		"search", 0, start_time, GetQueryParams(c), response, "GetFormByCaseSubType Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)

}

// old db stuct
// func GetFormByCaseSubType(c *gin.Context) {
// 	logger := utils.GetLog()
// 	conn, ctx, cancel := utils.ConnectDB()
// 	if conn == nil {
// 		return
// 	}
// 	defer cancel()
// 	defer conn.Close(ctx)
// 	id := c.Param("id")
// 	start_time := time.Now()
// 	username := GetVariableFromToken(c, "username")
// 	orgId := GetVariableFromToken(c, "orgId")
// 	txtId := uuid.New().String()
// 	var req model.FormByCasesubtype
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		response := model.Response{
// 			Status: "-1",
// 			Msg:    "Failure",
// 			Desc:   err.Error(),
// 		}
// 		//=======AUDIT_START=====//
// 		_ = utils.InsertAuditLogs(
// 			c, conn, orgId.(string), username.(string),
// 			txtId, id, "Form", "GetFormByCaseSubType", "",
// 			"search", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
// 		)
// 		//=======AUDIT_END=====//
// 		c.JSON(http.StatusBadRequest, response)
// 		logger.Warn("Insert failed", zap.Error(err))
// 		return
// 	}
// 	var wfId string
// 	// username := GetVariableFromToken(c, "username")
// 	query := `SELECT "wfId" FROM public.case_sub_types WHERE "orgId"=$1 AND "sTypeId"=$2`
// 	logger.Debug(`Query`, zap.String("query", query),
// 		zap.Any("Input", []any{
// 			orgId, req.CaseSubType,
// 		}))
// 	err := conn.QueryRow(ctx, query, orgId, req.CaseSubType).Scan(&wfId)
// 	if err != nil {
// 		logger.Warn("Query failed", zap.Error(err))
// 		response := model.Response{
// 			Status: "-1",
// 			Msg:    "Failure",
// 			Desc:   err.Error(),
// 		}
// 		//=======AUDIT_START=====//
// 		_ = utils.InsertAuditLogs(
// 			c, conn, orgId.(string), username.(string),
// 			txtId, id, "Form", "GetFormByCaseSubType", "",
// 			"search", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
// 		)
// 		//=======AUDIT_END=====//
// 		c.JSON(http.StatusInternalServerError, response)
// 		return
// 	}

// 	query = `SELECT
//     t2."section",
//     t2."data",
//     t2."nodeId",
//     t1."versions"
// FROM public.wf_definitions t1
// INNER JOIN public.wf_nodes t2
//     ON t1."wfId" = t2."wfId" AND t1."versions" = t2."versions"
// WHERE t1."wfId" = $1
//   AND t1."orgId" = $2
//   AND LOWER(t2."type") = 'process'
//   AND t2."data"->'data'->'config'->>'action' = 'S001'`
// 	logger.Debug(`Query`, zap.String("query", query), zap.Any("Input", []any{
// 		wfId, orgId,
// 	}))

// 	var rows pgx.Rows
// 	rows, err = conn.Query(ctx, query, wfId, orgId)
// 	if err != nil {
// 		logger.Warn("Query failed", zap.Error(err))
// 		response := model.Response{
// 			Status: "-1",
// 			Msg:    "Failure",
// 			Desc:   err.Error(),
// 		}
// 		//=======AUDIT_START=====//
// 		_ = utils.InsertAuditLogs(
// 			c, conn, orgId.(string), username.(string),
// 			txtId, id, "Form", "GetFormByCaseSubType", "",
// 			"search", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
// 		)
// 		//=======AUDIT_END=====//
// 		c.JSON(http.StatusInternalServerError, response)
// 		return
// 	}
// 	defer rows.Close()
// 	var formId string
// 	var nodeId string
// 	var versions string
// 	for rows.Next() {
// 		var rawJSON []byte
// 		var rowsType string
// 		var tempNodeId, tempVersions string

// 		err := rows.Scan(&rowsType, &rawJSON, &tempNodeId, &tempVersions)
// 		if err != nil {
// 			logger.Warn("Scan failed", zap.Error(err))
// 			continue
// 		}

// 		// Only handle nodes
// 		if rowsType != "nodes" {
// 			continue
// 		}

// 		var nodeMap map[string]interface{}
// 		if err := json.Unmarshal(rawJSON, &nodeMap); err != nil {
// 			logger.Warn("Unmarshal failed", zap.Error(err))
// 			continue
// 		}

// 		// Access data -> config -> formId if action = S002
// 		if data, ok := nodeMap["data"].(map[string]interface{}); ok {
// 			if config, ok := data["config"].(map[string]interface{}); ok {
// 				if action, ok := config["action"].(string); ok && action == "S001" {
// 					if formVal, ok := config["formId"].(string); ok {
// 						formId = formVal
// 						nodeId = tempNodeId
// 						versions = tempVersions
// 						break // found, exit loop
// 					}
// 				}
// 			}
// 		}
// 	}
// 	logger.Debug(formId)
// 	rows.Close()
// 	query = `SELECT t1."formId",t1."formName",t1."formColSpan",t2."eleData"
// 	FROM public.form_builder t1
// 	INNER JOIN public.form_elements t2
// 	ON t1."formId"=t2."formId" AND t2."versions"=t1."versions"
// 	WHERE t1."formId" = $1 AND t1."orgId"=$2 `

// 	logger.Debug(`Query`, zap.String("query", query), zap.Any("Input", []any{
// 		formId, orgId,
// 	}))
// 	rows, err = conn.Query(ctx, query, formId, orgId)
// 	if err != nil {
// 		logger.Warn("Query failed", zap.Error(err))
// 		response := model.Response{
// 			Status: "-1",
// 			Msg:    "Failure",
// 			Desc:   err.Error(),
// 		}
// 		//=======AUDIT_START=====//
// 		_ = utils.InsertAuditLogs(
// 			c, conn, orgId.(string), username.(string),
// 			txtId, id, "Form", "GetFormByCaseSubType", "",
// 			"search", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
// 		)
// 		//=======AUDIT_END=====//
// 		c.JSON(http.StatusInternalServerError, response)
// 		return
// 	}
// 	defer rows.Close()
// 	var formFields []model.IndividualFormField
// 	var form model.FormAnswerRequest
// 	found := false
// 	for rows.Next() {
// 		var rawJSON []byte
// 		err := rows.Scan(&form.FormId, &form.FormName, &form.FormColSpan, &rawJSON)
// 		if err != nil {
// 			logger.Warn("Scan failed", zap.Error(err))
// 			response := model.Response{
// 				Status: "-1",
// 				Msg:    "Failed",
// 				Desc:   err.Error(),
// 			}
// 			//=======AUDIT_START=====//
// 			_ = utils.InsertAuditLogs(
// 				c, conn, orgId.(string), username.(string),
// 				txtId, id, "Form", "GetFormByCaseSubType", "",
// 				"search", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
// 			)
// 			//=======AUDIT_END=====//
// 			c.JSON(http.StatusOK, response)
// 			return
// 		}

// 		var field model.IndividualFormField
// 		if err := json.Unmarshal(rawJSON, &field); err != nil {
// 			logger.Warn("unmarshal field failed", zap.Error(err))
// 			response := model.Response{
// 				Status: "-1",
// 				Msg:    "Failed",
// 				Desc:   err.Error(),
// 			}
// 			//=======AUDIT_START=====//
// 			_ = utils.InsertAuditLogs(
// 				c, conn, orgId.(string), username.(string),
// 				txtId, id, "Form", "GetFormByCaseSubType", "",
// 				"search", -1, start_time, GetQueryParams(c), response, "Failed : "+err.Error(),
// 			)
// 			//=======AUDIT_END=====//
// 			c.JSON(http.StatusOK, response)
// 			return
// 		}

// 		formFields = append(formFields, field)
// 		form.FormFieldJson = formFields
// 		found = true
// 	}
// 	if !found {
// 		response := model.Response{
// 			Status: "-1",
// 			Msg:    "Failed",
// 			Desc:   "No form data found",
// 		}
// 		//=======AUDIT_START=====//
// 		_ = utils.InsertAuditLogs(
// 			c, conn, orgId.(string), username.(string),
// 			txtId, id, "Form", "GetFormByCaseSubType", "",
// 			"search", -1, start_time, GetQueryParams(c), response, "Not Found.",
// 		)
// 		//=======AUDIT_END=====//
// 		c.JSON(http.StatusInternalServerError, response)
// 		return
// 	}
// 	form.NextNodeId = nodeId
// 	form.Versions = versions
// 	form.WfId = wfId
// 	response := model.Response{
// 		Status: "0",
// 		Msg:    "Success",
// 		Data:   form,
// 		Desc:   "",
// 	}
// 	//=======AUDIT_START=====//
// 	_ = utils.InsertAuditLogs(
// 		c, conn, orgId.(string), username.(string),
// 		txtId, id, "Form", "GetFormByCaseSubType", "",
// 		"search", 0, start_time, GetQueryParams(c), response, "GetFormByCaseSubType Success.",
// 	)
// 	//=======AUDIT_END=====//
// 	c.JSON(http.StatusOK, response)

// }

func InsertFormAnswer(conn *pgx.Conn, ctx context.Context, orgId string, caseId string, fa model.FormAnswerRequest, user string) error {

	form := model.Form{
		FormName:      &fa.FormName,
		FormColSpan:   fa.FormColSpan,
		FormFieldJson: fa.FormFieldJson,
		FormId:        &fa.FormId,
	}
	var insertedID int64
	query := `
			INSERT INTO form_answers (
				"orgId", "caseId", "formId", "versions",
				 "eleData",
				"createdAt", "updatedAt", "createdBy", "updatedBy"
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
			RETURNING id
		`

	err := conn.QueryRow(ctx, query,
		orgId,
		caseId,
		fa.FormId,
		fa.Versions,
		form,
		time.Now(),
		time.Now(),
		user,
		user,
	).Scan(&insertedID)

	if err != nil {
		return err
	}

	return nil
}

// for old db stuct
// func __InsertFormAnswer(conn *pgx.Conn, ctx context.Context, orgId string, caseId string, fa model.FormAnswerRequest, user string) error {
// 	for i, field := range fa.FormFieldJson {
// 		eleDataJSON, err := json.Marshal(field)
// 		if err != nil {
// 			return err
// 		}

// 		var insertedID int64
// 		query := `
// 			INSERT INTO form_answers (
// 				"orgId", "caseId", "formId", "versions",
// 				"eleNumber", "eleData",
// 				"createdAt", "updatedAt", "createdBy", "updatedBy"
// 			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
// 			RETURNING id
// 		`

// 		err = conn.QueryRow(ctx, query,
// 			orgId,
// 			caseId,
// 			fa.FormId,
// 			fa.Versions,
// 			i+1,
// 			eleDataJSON,
// 			time.Now(),
// 			time.Now(),
// 			user,
// 			user,
// 		).Scan(&insertedID)

// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

func UpdateFormAnswer(conn *pgx.Conn, ctx context.Context, orgId string, caseId string, fa model.FormAnswerRequest, user string) error {
	oldform, err := GetFormAnswers(conn, ctx, orgId, caseId, fa.FormId, true)
	if err != nil {
		log.Fatal("query error:", err)
	}
	if len(oldform.FormFieldJson) != len(fa.FormFieldJson) {
		log.Fatal("form not match")
	}

	query := `
		UPDATE form_answers
		SET 
			"eleData" = $1,
			"updatedAt" = NOW(),
			"updatedBy" = $2
		WHERE 
			"id" = $3 
			
	`
	var uid = *oldform.UID

	if uid == "" {
		return errors.New("no form UID")
	}
	_, err = conn.Exec(ctx, query, fa, user, uid)
	if err != nil {
		return err
	}

	return nil
}

//old db stuct
// func __UpdateFormAnswer(conn *pgx.Conn, ctx context.Context, orgId string, caseId string, fa model.FormAnswerRequest, user string) error {
// 	oldform, err := GetFormAnswers(conn, ctx, orgId, caseId, fa.FormId, true)
// 	if err != nil {
// 		log.Fatal("query error:", err)
// 	}
// 	if len(oldform.FormFieldJson) != len(fa.FormFieldJson) {
// 		log.Fatal("form not match")
// 	}
// 	for i, field := range fa.FormFieldJson { // i = int, field = map[string]interface{}
// 		eleDataJSON, err := json.Marshal(field)
// 		if err != nil {
// 			return err
// 		}
// 		if field.ID != oldform.FormFieldJson[i].ID {
// 			return errors.New("form id does't not match")
// 		}

// 		oldEleDataJSON, err := json.Marshal(oldform.FormFieldJson[i])
// 		if bytes.Equal(oldEleDataJSON, eleDataJSON) {
// 			log.Printf("Skipping field - value unchanged ")
// 			continue
// 		}
// 		if oldform.FormFieldJson[i].UID == nil {
// 			log.Printf("⚠️ Skipping field because UID is empty at index %d", i)
// 			return errors.New("no form UID")
// 		}
// 		var uid = *oldform.FormFieldJson[i].UID

// 		if uid == "" {
// 			return errors.New("no form UID")
// 		}

// 		query := `
// 		UPDATE form_answers
// 		SET
// 			"eleData" = $1,
// 			"updatedAt" = NOW(),
// 			"updatedBy" = $2
// 		WHERE
// 			"id" = $3

// 	`

// 		_, err = conn.Exec(ctx, query, eleDataJSON, user, uid)
// 		if err != nil {
// 			return err
// 		}

// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

func InsertFormElement(conn *pgx.Conn, ctx context.Context, req model.Form, orgId string, user string, versions string, formId string, createAt *time.Time, createBy *string) error {

	now := time.Now()
	var id int
	form := model.Form{
		FormName:      req.FormName,
		FormColSpan:   req.FormColSpan,
		FormFieldJson: req.FormFieldJson,
		FormId:        &formId,
	}

	eleData, err := json.Marshal(form)
	if err != nil {
		return err
	}
	var query string
	if versions == "1" {
		query = `
				INSERT INTO public.form_elements(
				"orgId", "formId", versions, "eleData", "createdAt", "updatedAt", "createdBy", "updatedBy")
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				RETURNING id ;
				`

		err = conn.QueryRow(ctx, query,
			orgId, formId, versions, eleData, now, now, user, user).Scan(&id)
		if err != nil {
			return err
		}
	} else {

		query = `
				INSERT INTO public.form_elements(
				"orgId", "formId", versions, "eleData", "createdAt", "updatedAt", "createdBy", "updatedBy")
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				RETURNING id ;
				`

		err = conn.QueryRow(ctx, query,
			orgId, formId, versions, eleData, &createAt, now, &createBy, user).Scan(&id)
		if err != nil {
			return err
		}
	}
	return nil
}
