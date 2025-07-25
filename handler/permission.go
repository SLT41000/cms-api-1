package handler

import (
	"mainPackage/config"
	"mainPackage/model"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// ListPermission godoc
// @summary Get Permission
// @tags Permission
// @security ApiKeyAuth
// @id Get Permission
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/permission [get]
func GetPermission(c *gin.Context) {
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
	query := `SELECT  id, "groupName", "permId", "permName", active, "createdAt", "updatedAt", "createdBy", "updatedBy" 
	FROM public.um_permissions LIMIT $1 OFFSET $2`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	rows, err = conn.Query(ctx, query, length, start)
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
	var Permission model.Permission
	var PermissionList []model.Permission
	rowIndex := 0
	for rows.Next() {
		rowIndex++
		err := rows.Scan(&Permission.ID, &Permission.GroupName, &Permission.PermID, &Permission.PermName,
			&Permission.Active, &Permission.CreatedAt, &Permission.UpdatedAt,
			&Permission.CreatedBy, &Permission.UpdatedBy)
		if err != nil {
			logger.Warn("Scan failed", zap.Error(err))
			response := model.Response{
				Status: "-1",
				Msg:    "Failed",
				Desc:   errorMsg,
			}
			c.JSON(http.StatusInternalServerError, response)
		}
		PermissionList = append(PermissionList, Permission)
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
			Data:   PermissionList,
			Desc:   "",
		}
		c.JSON(http.StatusOK, response)
	}
}

// ListPermission godoc
// @summary Get Permission by id
// @tags Permission
// @security ApiKeyAuth
// @id Get Permission by id
// @accept json
// @produce json
// @Param permId path string true "permId"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/permission/{permId} [get]
func GetPermissionById(c *gin.Context) {
	logger := config.GetLog()
	id := c.Param("permId")
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	query := `SELECT  id, "groupName", "permId", "permName", active, "createdAt", "updatedAt", "createdBy", "updatedBy" 
	FROM public.um_permissions WHERE "permId"=$1`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
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
	var Permission model.Permission
	err = rows.Scan(&Permission.ID, &Permission.GroupName, &Permission.PermID, &Permission.PermName,
		&Permission.Active, &Permission.CreatedAt, &Permission.UpdatedAt,
		&Permission.CreatedBy, &Permission.UpdatedBy)
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

	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   Permission,
		Desc:   "",
	}
	c.JSON(http.StatusOK, response)

}

// @summary Create Permission
// @id Create Permission
// @security ApiKeyAuth
// @tags Permission
// @accept json
// @produce json
// @param Body body model.PermissionInsert true "Create Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/permission/add [post]
func InsertPermission(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	var req model.PermissionInsert
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
	uuid := uuid.New()
	now := time.Now()
	var id int
	query := `
	INSERT INTO public."um_permissions"(
	"groupName", "permId", "permName",active, "createdAt", "updatedAt", "createdBy", "updatedBy")
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	RETURNING id ;
	`

	err := conn.QueryRow(ctx, query,
		req.GroupName, uuid, req.PermName, req.Active, now,
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

// @summary Update Permission
// @id Update Permission
// @security ApiKeyAuth
// @accept json
// @produce json
// @tags Permission
// @Param permId path string true "permId"
// @param Body body model.PermissionUpdate true "Update data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/permission/{permId} [patch]
func UpdatePermission(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	id := c.Param("id")

	var req model.PermissionUpdate
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
	query := `UPDATE public."um_permissions"
	SET "groupName"=$2, "permName"=$3,active=$4,
	 "updatedAt"=$5, "updatedBy"=$6
	WHERE "permId" = $1 `
	_, err := conn.Exec(ctx, query,
		id, req.GroupName, req.PermName, req.Active, now, username)
	logger.Debug("Update Case SQL Args",
		zap.String("query", query),
		zap.Any("Input", []any{
			id, req.GroupName, req.PermName, req.Active, now, username,
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

// @summary Delete Permission
// @id Delete Permission
// @security ApiKeyAuth
// @accept json
// @tags Permission
// @produce json
// @Param permId path string true "permId"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/permission/{permId} [delete]
func DeletePermission(c *gin.Context) {

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
	query := `DELETE FROM public."um_permissions" WHERE "permId" = $1`
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
