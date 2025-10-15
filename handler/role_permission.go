package handler

import (
	"mainPackage/model"
	"mainPackage/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// ListRolePermission godoc
// @summary Get RolePermission
// @tags Role
// @security ApiKeyAuth
// @id Get RolePermission
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/role_permission [get]
func GetRolePermission(c *gin.Context) {
	logger := utils.GetLog()
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
	query := `SELECT id, "orgId", "roleId", "permId", active, "createdAt", "updatedAt", "createdBy", "updatedBy"
	FROM public.um_role_with_permissions WHERE "orgId"=$1 LIMIT $2 OFFSET $3`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	rows, err = conn.Query(ctx, query, orgId, length, start)
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
	var RolePermission model.RolePermission
	var RolePermissionList []model.RolePermission
	rowIndex := 0
	for rows.Next() {
		rowIndex++
		err := rows.Scan(&RolePermission.ID, &RolePermission.OrgID, &RolePermission.RoleID, &RolePermission.PermID,
			&RolePermission.Active, &RolePermission.CreatedAt, &RolePermission.UpdatedAt, &RolePermission.CreatedBy, &RolePermission.UpdatedBy)
		if err != nil {
			logger.Warn("Scan failed", zap.Error(err))
			response := model.Response{
				Status: "-1",
				Msg:    "Failed",
				Desc:   errorMsg,
			}
			c.JSON(http.StatusInternalServerError, response)
		}
		RolePermissionList = append(RolePermissionList, RolePermission)
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
			Data:   RolePermissionList,
			Desc:   "",
		}
		c.JSON(http.StatusOK, response)
	}
}

// ListRolePermission godoc
// @summary Get RolePermission by ID
// @tags Role
// @security ApiKeyAuth
// @id Get RolePermission by ID
// @accept json
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/role_permission/{id} [get]
func GetRolePermissionbyId(c *gin.Context) {
	logger := utils.GetLog()
	id := c.Param("id")
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	orgId := GetVariableFromToken(c, "orgId")

	query := `SELECT id, "orgId", "roleId", "permId", active, "createdAt", "updatedAt", "createdBy", "updatedBy"
	FROM public.um_role_with_permissions WHERE id = $1 AND "orgId" = $2`

	logger.Debug("Query", zap.String("query", query),
		zap.Any("Input", []any{id, orgId}),
	)

	var RolePermission model.RolePermission
	err := conn.QueryRow(ctx, query, id, orgId).Scan(
		&RolePermission.ID,
		&RolePermission.OrgID,
		&RolePermission.RoleID,
		&RolePermission.PermID,
		&RolePermission.Active,
		&RolePermission.CreatedAt,
		&RolePermission.UpdatedAt,
		&RolePermission.CreatedBy,
		&RolePermission.UpdatedBy,
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

	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   RolePermission,
		Desc:   "",
	}
	c.JSON(http.StatusOK, response)

}

// ListRolePermission godoc
// @summary Get RolePermission by roleID
// @tags Role
// @security ApiKeyAuth
// @id Get RolePermission by roleID
// @accept json
// @produce json
// @Param roleId path string true "roleId"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/role_permission/roleId/{roleId} [get]
func GetRolePermissionbyroleId(c *gin.Context) {
	logger := utils.GetLog()
	id := c.Param("roleId")
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	orgId := GetVariableFromToken(c, "orgId")

	query := `SELECT id, "orgId", "roleId", "permId", active, "createdAt", "updatedAt", "createdBy", "updatedBy"
	FROM public.um_role_with_permissions WHERE "orgId"=$1 AND "roleId"=$2`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	rows, err := conn.Query(ctx, query, orgId, id)
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
	var RolePermission model.RolePermission
	var RolePermissionList []model.RolePermission
	rowIndex := 0
	found := false
	for rows.Next() {
		rowIndex++
		err := rows.Scan(&RolePermission.ID, &RolePermission.OrgID, &RolePermission.RoleID, &RolePermission.PermID,
			&RolePermission.Active, &RolePermission.CreatedAt, &RolePermission.UpdatedAt, &RolePermission.CreatedBy, &RolePermission.UpdatedBy)
		if err != nil {
			logger.Warn("Scan failed", zap.Error(err))
			response := model.Response{
				Status: "-1",
				Msg:    "Failed",
				Desc:   err.Error(),
			}
			c.JSON(http.StatusInternalServerError, response)
		}
		RolePermissionList = append(RolePermissionList, RolePermission)
		found = true
	}
	if !found {
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   "Not Found",
		}
		c.JSON(http.StatusInternalServerError, response)
	} else {
		response := model.Response{
			Status: "0",
			Msg:    "Success",
			Data:   RolePermissionList,
			Desc:   "",
		}
		c.JSON(http.StatusOK, response)
	}
}

// @summary Create RolePermission
// @id Create RolePermission
// @security ApiKeyAuth
// @tags Role
// @accept json
// @produce json
// @param Body body model.RolePermissionInsert true "Create Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/role_permission/add [post]
func InsertRolePermission(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	now := time.Now()

	var req model.RolePermissionInsert
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	for i, item := range req.PermID {
		logger.Debug("eleNumber", zap.Int("i", i+1))
		logger.Debug("JsonArray", zap.Any("Json", item))
		// logger.Debug("JsonArray", zap.Any("active", item.Active))
		// logger.Debug("JsonArray", zap.Any("permId", item.PermID))

		var id int
		query := `
		INSERT INTO public."um_role_with_permissions"(
		"orgId", "roleId", "permId", active, "createdAt", "updatedAt", "createdBy", "updatedBy")
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id ;
		`
		logger.Debug(`Query`, zap.String("query", query),
			zap.Any("Input", []any{
				req, item,
			}))

		err := conn.QueryRow(ctx, query,
			orgId, req.RoleID, item.PermID, item.Active, now, now, username, username).Scan(&id)

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

	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Create successfully",
	})

}

// @summary Update RolePermission
// @id Update RolePermission
// @security ApiKeyAuth
// @accept json
// @produce json
// @tags Role
// @Param roleId path string true "roleId"
// @param Body body model.RolePermissionUpdate true "Update data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/role_permission/{roleId} [patch]
func UpdateRolePermission(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	id := c.Param("roleId")

	var req model.RolePermissionUpdate
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
	query := `DELETE FROM public."um_role_with_permissions" WHERE "roleId" = $1 AND "orgId"=$2`

	logger.Debug(`Query`, zap.String("query", query),
		zap.Any("Input", []any{
			id, orgId,
		}))

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
	logger.Debug("Update Case SQL Args",
		zap.String("query", query),
		zap.Any("Input", []any{id, orgId}))

	for i, item := range req.PermID {
		logger.Debug("eleNumber", zap.Int("i", i+1))
		logger.Debug("JsonArray", zap.Any("Json", item))
		// logger.Debug("JsonArray", zap.Any("active", item.Active))
		// logger.Debug("JsonArray", zap.Any("permId", item.PermID))

		var id int
		query := `
		INSERT INTO public."um_role_with_permissions"(
		"orgId", "roleId", "permId", active, "createdAt", "updatedAt", "createdBy", "updatedBy")
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id ;
		`
		logger.Debug(`Query`, zap.String("query", query),
			zap.Any("Input", []any{
				req, item,
			}))

		err := conn.QueryRow(ctx, query,
			orgId, id, item.PermID, item.Active, now, now, username, username).Scan(&id)

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

	// Continue logic...
	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Update successfully",
	})
}

// @summary Update Multi RolePermission
// @id Update Multi RolePermission
// @security ApiKeyAuth
// @accept json
// @produce json
// @tags Role
// @param Body body model.MultiRolePermissionUpdate true "Update data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/role_permission/multi [patch]
func UpdateMultiRolePermission(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	var req model.MultiRolePermissionUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Update failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		return
	}

	for _, item := range req.Body {
		logger.Debug("JsonArray", zap.Any("roleId", item.RoleID))
		id := item.RoleID
		now := time.Now()
		username := GetVariableFromToken(c, "username")
		orgId := GetVariableFromToken(c, "orgId")
		query := `DELETE FROM public."um_role_with_permissions" WHERE "roleId" = $1 AND "orgId"=$2`

		logger.Debug(`Query`, zap.String("query", query),
			zap.Any("Input", []any{
				id, orgId,
			}))
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
		logger.Debug("Update Case SQL Args",
			zap.String("query", query),
			zap.Any("Input", []any{id, orgId}))

		for _, items := range item.PermID {
			logger.Debug("JsonArray", zap.Any("PermID", items.PermID))

			query := `
				INSERT INTO public."um_role_with_permissions"(
				"orgId", "roleId", "permId", active, "createdAt", "updatedAt", "createdBy", "updatedBy")
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				`
			logger.Debug(`Query`, zap.String("query", query),
				zap.Any("Input", []any{
					req, items,
				}))

			_, err := conn.Exec(ctx, query,
				orgId, id, items.PermID, items.Active, now, now, username, username)

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
	}

	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Update successfully",
	})
}

// @summary Delete RolePermission
// @id Delete RolePermission
// @security ApiKeyAuth
// @accept json
// @tags Role
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/role_permission/{id} [delete]
func DeleteRolePermission(c *gin.Context) {

	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()
	orgId := GetVariableFromToken(c, "orgId")
	id := c.Param("id")
	query := `DELETE FROM public."um_role_with_permissions" WHERE id = $1 AND "orgId"=$2`
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
