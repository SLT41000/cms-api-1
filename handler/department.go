package handler

import (
	"mainPackage/config"
	"mainPackage/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// ListDepartment godoc
// @summary Get Department
// @tags Department
// @security ApiKeyAuth
// @id Get Department
// @accept json
// @produce json
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/departments [get]
func GetDepartment(c *gin.Context) {
	logger := config.GetLog()

	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	query := `SELECT "deptId", "orgId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy" FROM public.sec_departments`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	rows, err := conn.Query(ctx, query)
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
	var Department model.Department
	var DepartmentList []model.Department
	rowIndex := 0
	for rows.Next() {
		rowIndex++
		err := rows.Scan(&Department.DeptID, &Department.OrgID, &Department.En, &Department.Th,
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
