package handler

import (
	"mainPackage/config"
	"mainPackage/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// ListCommands godoc
// @summary Get Commands
// @tags Commands
// @security ApiKeyAuth
// @id Get Commands
// @accept json
// @produce json
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/commands [get]
func GetCommand(c *gin.Context) {
	logger := config.GetLog()

	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	orgId := GetVariableFromToken(c, "orgId")
	query := `SELECT  "deptId", "orgId", "commId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy" 
	FROM public.sec_commands WHERE "orgId"=$1 LIMIT 1000`

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
	var Department model.Command
	var DepartmentList []model.Command
	rowIndex := 0
	for rows.Next() {
		rowIndex++
		err := rows.Scan(&Department.DeptID, &Department.OrgID, &Department.CommID, &Department.En, &Department.Th,
			&Department.Active, &Department.CreatedAt, &Department.UpdatedAt, &Department.CreatedBy, &Department.UpdatedBy)
		if err != nil {
			logger.Warn("Scan failed", zap.Error(err))
			response := model.Response{
				Status: "-1",
				Msg:    "Failed",
				Desc:   errorMsg,
			}
			c.JSON(http.StatusInternalServerError, response)
		}
		DepartmentList = append(DepartmentList, Department)
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
			Data:   DepartmentList,
			Desc:   "",
		}
		c.JSON(http.StatusOK, response)
	}
}

// ListCommands godoc
// @summary Get Commands by id
// @tags Commands
// @security ApiKeyAuth
// @id Get Commands by id
// @accept json
// @produce json
// @Param id path string true "id" "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/commands/{id} [get]
func GetCommandById(c *gin.Context) {
	logger := config.GetLog()
	id := c.Param("id")
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	orgId := GetVariableFromToken(c, "orgId")
	query := `SELECT  "deptId", "orgId", "commId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy" 
	FROM public.sec_commands WHERE "commId"=$1 AND "orgId"=$2`

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
	var Department model.Command
	rowIndex := 0
	for rows.Next() {
		rowIndex++
		err := rows.Scan(&Department.DeptID, &Department.OrgID, &Department.CommID, &Department.En, &Department.Th,
			&Department.Active, &Department.CreatedAt, &Department.UpdatedAt, &Department.CreatedBy, &Department.UpdatedBy)
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
			Data:   Department,
			Desc:   "",
		}
		c.JSON(http.StatusOK, response)
	}
}
