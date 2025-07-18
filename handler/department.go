package handler

import (
	"mainPackage/config"
	"mainPackage/model"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// ListDepartment godoc
// @summary Get Department
// @tags Dispatch
// @security ApiKeyAuth
// @id Get Department
// @accept json
// @produce json
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/departments [get]
func GetDepartment(c *gin.Context) {
	logger := config.GetLog()
	orgId := GetVariableFromToken(c, "orgId")
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	query := `SELECT id,"deptId", "orgId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy" 
	FROM public.sec_departments WHERE "orgId"=$1 LIMIT 1000`

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
	var Department model.Department
	var DepartmentList []model.Department
	rowIndex := 0
	for rows.Next() {
		rowIndex++
		err := rows.Scan(&Department.ID, &Department.DeptID, &Department.OrgID, &Department.En, &Department.Th,
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

// ListDepartment godoc
// @summary Get Department by ID
// @tags Dispatch
// @security ApiKeyAuth
// @id Get Department by ID
// @accept json
// @produce json
// @Param id path string true "id" "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/departments/{id} [get]
func GetDepartmentbyId(c *gin.Context) {
	logger := config.GetLog()
	id := c.Param("id")
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	orgId := GetVariableFromToken(c, "orgId")
	query := `SELECT id,"deptId", "orgId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy" 
	FROM public.sec_departments WHERE "deptId" = $1 AND "orgId"=$2`

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
	var Department model.Department
	err = rows.Scan(&Department.ID, &Department.DeptID, &Department.OrgID, &Department.En, &Department.Th,
		&Department.Active, &Department.CreatedAt, &Department.UpdatedAt, &Department.CreatedBy, &Department.UpdatedBy)
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

// @summary Create Department
// @id Create Department
// @security ApiKeyAuth
// @tags Dispatch
// @accept json
// @produce json
// @param Body body model.DepartmentInsert true "Create Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/departments/add [post]
func InsertDepartment(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	var req model.DepartmentInsert
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
	orgId := GetVariableFromToken(c, "orgId")
	now := time.Now()
	uuid := uuid.New()
	var id int
	query := `
	INSERT INTO public."sec_departments"(
	"deptId", "orgId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy")
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

// @summary Update Department
// @id Update Department
// @security ApiKeyAuth
// @accept json
// @produce json
// @tags Dispatch
// @Param id path int true "id"
// @param Body body model.DepartmentUpdate true "Update data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/departments/{id} [patch]
func UpdateDepartment(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	id := c.Param("id")

	var req model.DepartmentUpdate
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

// @summary Delete Department
// @id Delete Department
// @security ApiKeyAuth
// @accept json
// @tags Dispatch
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/departments/{id} [delete]
func DeleteDepartment(c *gin.Context) {

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
	query := `DELETE FROM public."sec_departments" WHERE id = $1 AND "orgId"=$2`
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
