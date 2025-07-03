package handler

import (
	"encoding/json"
	"mainPackage/config"
	"mainPackage/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// ListCase godoc
// @summary Get Form
// @tags Form
// @security ApiKeyAuth
// @id Get Form
// @accept json
// @produce json
// @Param id path string true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/forms/{id} [get]
func GetForm(c *gin.Context) {
	logger := config.GetLog()
	id := c.Param("id")

	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	query := `SELECT form_builder."formId",form_builder."formName",form_builder."formColSpan",form_elements."eleData" FROM public.form_builder INNER JOIN public.form_elements ON form_builder."formId"=form_elements."formId" WHERE form_builder."formId" = $1`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	logger.Debug(`id :` + id)
	rows, err := conn.Query(ctx, query, id)
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
	var formFields []map[string]interface{}
	var form model.FormGetOptModel
	for rows.Next() {
		var rawJSON []byte
		err := rows.Scan(&form.FormId, &form.FormName, &form.FormColSpan, &rawJSON)
		if err != nil {
			logger.Warn("Scan failed", zap.Error(err))
			continue
		}

		var field map[string]interface{}
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
		c.JSON(http.StatusInternalServerError, response)
	} else {
		response := model.Response{
			Status: "0",
			Msg:    "Success",
			Data:   form,
			Desc:   "",
		}
		c.JSON(http.StatusOK, response)
	}
}
