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

// @summary Get Mmd Properties
// @tags Mobile device management (Units)
// @security ApiKeyAuth
// @id Get Mmd Properties
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/mdm/properties [get]
func GetMmdProperty(c *gin.Context) {
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
	query := `SELECT  id, "propId", "orgId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy"
	FROM public.mdm_properties WHERE "orgId"=$1 LIMIT $2 OFFSET $3`

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
	var Property model.MmdProperty
	var PropertyList []model.MmdProperty
	rowIndex := 0
	for rows.Next() {
		rowIndex++
		err := rows.Scan(&Property.ID, &Property.PropID, &Property.OrgID, &Property.EN, &Property.TH,
			&Property.Active, &Property.CreatedAt, &Property.UpdatedAt, &Property.CreatedBy, &Property.UpdatedBy)
		if err != nil {
			logger.Warn("Scan failed", zap.Error(err))
			response := model.Response{
				Status: "-1",
				Msg:    "Failed",
				Desc:   errorMsg,
			}
			c.JSON(http.StatusInternalServerError, response)
			return
		}
		PropertyList = append(PropertyList, Property)
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
			Data:   PropertyList,
			Desc:   "",
		}
		c.JSON(http.StatusOK, response)
	}
}

// @summary Get Mmd Properties by Id
// @tags Mobile device management (Units)
// @security ApiKeyAuth
// @id Get Mmd Properties by Id
// @accept json
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/mdm/properties/{id} [get]
func GetMmdPropertyById(c *gin.Context) {
	logger := config.GetLog()
	id := c.Param("id")
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	orgId := GetVariableFromToken(c, "orgId")
	defer cancel()
	defer conn.Close(ctx)
	query := `SELECT  id, "propId", "orgId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy"
	FROM public.mdm_properties WHERE id=$1 AND "orgId"=$2`

	var Property model.MmdProperty
	logger.Debug(`Query`, zap.String("query", query))
	err := conn.QueryRow(ctx, query, id, orgId).Scan(&Property.ID, &Property.PropID, &Property.OrgID, &Property.EN, &Property.TH,
		&Property.Active, &Property.CreatedAt, &Property.UpdatedAt, &Property.CreatedBy, &Property.UpdatedBy)
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
		Data:   Property,
		Desc:   "",
	}
	c.JSON(http.StatusOK, response)

}

// @summary Create Mmd Properties
// @id Create Mmd Properties
// @security ApiKeyAuth
// @tags Mobile device management (Units)
// @accept json
// @produce json
// @param Body body model.MmdPropertyInsert true "Create Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/mdm/properties/add [post]
func InsertMmdProperty(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	var req model.MmdPropertyInsert
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
	INSERT INTO public."mdm_properties"(
	"propId", "orgId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy")
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id ;
	`

	err := conn.QueryRow(ctx, query,
		uuid, orgId, req.EN, req.TH, req.Active, now, now, username, username).Scan(&id)

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

// @summary Update Mmd Properties
// @id Update Mmd Properties
// @security ApiKeyAuth
// @accept json
// @produce json
// @tags Mobile device management (Units)
// @Param id path int true "id"
// @param Body body model.MmdPropertyUpdate true "Update data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/mdm/properties/{id} [patch]
func UpdateMmdProperty(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	id := c.Param("id")

	var req model.MmdPropertyUpdate
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
	query := `UPDATE public."mdm_properties"
	SET en=$2, th=$3, active=$4,
	 "updatedAt"=$5, "updatedBy"=$6
	WHERE id = $1 AND "orgId"=$7`
	_, err := conn.Exec(ctx, query,
		id, req.EN, req.TH, req.Active,
		now, username, orgId,
	)
	logger.Debug("Update Case SQL Args",
		zap.String("query", query),
		zap.Any("Input", []any{
			id, req.EN, req.TH, req.Active,
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

// @summary Delete Mmd Properties
// @id Delete Mmd Properties
// @security ApiKeyAuth
// @accept json
// @tags Mobile device management (Units)
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/mdm/properties/{id} [delete]
func DeleteMmdProperty(c *gin.Context) {

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
	query := `DELETE FROM public."mdm_properties" WHERE id = $1 AND "orgId"=$2`
	logger.Debug("Query", zap.String("query", query), zap.Any("id", id))
	_, err := conn.Exec(ctx, query, id, orgId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Update failed", zap.Error(err))
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Delete successfully",
	})
}

// @summary Get Mmd Unit Sources
// @tags Mobile device management (Units)
// @security ApiKeyAuth
// @id Get Mmd Unit Sources
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/mdm/sources [get]
func GetMmdUnitSources(c *gin.Context) {
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
	query := `SELECT  id, "unitSourceId", "orgId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy"
	FROM public.mdm_unit_sources WHERE "orgId"=$1 LIMIT $2 OFFSET $3`

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
	var Property model.MmdUnitSource
	var PropertyList []model.MmdUnitSource
	rowIndex := 0
	for rows.Next() {
		rowIndex++
		err := rows.Scan(&Property.ID, &Property.UnitSourceID, &Property.OrgID, &Property.EN, &Property.TH,
			&Property.Active, &Property.CreatedAt, &Property.UpdatedAt, &Property.CreatedBy, &Property.UpdatedBy)
		if err != nil {
			logger.Warn("Scan failed", zap.Error(err))
			response := model.Response{
				Status: "-1",
				Msg:    "Failed",
				Desc:   errorMsg,
			}
			c.JSON(http.StatusInternalServerError, response)
			return
		}
		PropertyList = append(PropertyList, Property)
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
			Data:   PropertyList,
			Desc:   "",
		}
		c.JSON(http.StatusOK, response)
	}
}

// @summary Get Mmd Unit Sources by Id
// @tags Mobile device management (Units)
// @security ApiKeyAuth
// @id Get Mmd Unit Sources by Id
// @accept json
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/mdm/sources/{id} [get]
func GetMmdUnitSourcesById(c *gin.Context) {
	logger := config.GetLog()
	id := c.Param("id")
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	orgId := GetVariableFromToken(c, "orgId")
	defer cancel()
	defer conn.Close(ctx)
	query := `SELECT  id, "unitSourceId", "orgId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy"
	FROM public.mdm_unit_sources WHERE id=$1 AND "orgId"=$2`

	var Property model.MmdUnitSource
	logger.Debug(`Query`, zap.String("query", query))
	err := conn.QueryRow(ctx, query, id, orgId).Scan(&Property.ID, &Property.UnitSourceID, &Property.OrgID, &Property.EN, &Property.TH,
		&Property.Active, &Property.CreatedAt, &Property.UpdatedAt, &Property.CreatedBy, &Property.UpdatedBy)
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
		Data:   Property,
		Desc:   "",
	}
	c.JSON(http.StatusOK, response)

}

// @summary Create Mmd Unit Sources
// @id Create Mmd Unit Sources
// @security ApiKeyAuth
// @tags Mobile device management (Units)
// @accept json
// @produce json
// @param Body body model.MmdUnitSourceInsert true "Create Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/mdm/sources/add [post]
func InsertMmdUnitSources(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	var req model.MmdUnitSourceInsert
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
	INSERT INTO public."mdm_unit_sources"(
	"unitSourceId", "orgId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy")
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id ;
	`

	err := conn.QueryRow(ctx, query,
		uuid, orgId, req.EN, req.TH, req.Active, now, now, username, username).Scan(&id)

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

// @summary Update Mmd Unit Sources
// @id Update Mmd Unit Sources
// @security ApiKeyAuth
// @accept json
// @produce json
// @tags Mobile device management (Units)
// @Param id path int true "id"
// @param Body body model.MmdUnitSourceUpdate true "Update data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/mdm/sources/{id} [patch]
func UpdateMmdUnitSources(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	id := c.Param("id")

	var req model.MmdUnitSourceUpdate
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
	query := `UPDATE public."mdm_unit_sources"
	SET en=$2, th=$3, active=$4,
	 "updatedAt"=$5, "updatedBy"=$6
	WHERE id = $1 AND "orgId"=$7`
	_, err := conn.Exec(ctx, query,
		id, req.EN, req.TH, req.Active,
		now, username, orgId,
	)
	logger.Debug("Update Case SQL Args",
		zap.String("query", query),
		zap.Any("Input", []any{
			id, req.EN, req.TH, req.Active,
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

// @summary Delete Mmd Unit Sources
// @id Delete Mmd Unit Sources
// @security ApiKeyAuth
// @accept json
// @tags Mobile device management (Units)
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/mdm/sources/{id} [delete]
func DeleteMmdUnitSources(c *gin.Context) {

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
	query := `DELETE FROM public."mdm_unit_sources" WHERE id = $1 AND "orgId"=$2`
	logger.Debug("Query", zap.String("query", query), zap.Any("id", id))
	_, err := conn.Exec(ctx, query, id, orgId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Update failed", zap.Error(err))
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Delete successfully",
	})
}

// @summary Get Mmd Unit Types
// @tags Mobile device management (Units)
// @security ApiKeyAuth
// @id Get Mmd Unit Types
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/mdm/types [get]
func GetMmdUnitType(c *gin.Context) {
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
	query := `SELECT  id, "unitTypeId", "orgId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy"
	FROM public.mdm_unit_types WHERE "orgId"=$1 LIMIT $2 OFFSET $3`

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
	var Property model.MmdUnitType
	var PropertyList []model.MmdUnitType
	found := false
	for rows.Next() {
		err := rows.Scan(&Property.ID, &Property.UnitTypeId, &Property.OrgID, &Property.EN, &Property.TH,
			&Property.Active, &Property.CreatedAt, &Property.UpdatedAt, &Property.CreatedBy, &Property.UpdatedBy)
		if err != nil {
			logger.Warn("Scan failed", zap.Error(err))
			response := model.Response{
				Status: "-1",
				Msg:    "Failed",
				Desc:   errorMsg,
			}
			c.JSON(http.StatusInternalServerError, response)
			return
		}
		PropertyList = append(PropertyList, Property)
		found = true
	}
	if !found {
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
			Data:   PropertyList,
			Desc:   "",
		}
		c.JSON(http.StatusOK, response)
	}
}

// @summary Get Mmd Unit Types by Id
// @tags Mobile device management (Units)
// @security ApiKeyAuth
// @id Get Mmd Unit Types by Id
// @accept json
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/mdm/types/{id} [get]
func GetMmdUnitTypeById(c *gin.Context) {
	logger := config.GetLog()
	id := c.Param("id")
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	orgId := GetVariableFromToken(c, "orgId")
	defer cancel()
	defer conn.Close(ctx)
	query := `SELECT  id, "unitTypeId", "orgId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy"
	FROM public.mdm_unit_types WHERE id=$1 AND "orgId"=$2`

	var Property model.MmdUnitType
	logger.Debug(`Query`, zap.String("query", query))
	err := conn.QueryRow(ctx, query, id, orgId).Scan(&Property.ID, &Property.UnitTypeId, &Property.OrgID, &Property.EN, &Property.TH,
		&Property.Active, &Property.CreatedAt, &Property.UpdatedAt, &Property.CreatedBy, &Property.UpdatedBy)
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
		Data:   Property,
		Desc:   "",
	}
	c.JSON(http.StatusOK, response)

}

// @summary Create Mmd Unit Types
// @id Create Mmd Unit Types
// @security ApiKeyAuth
// @tags Mobile device management (Units)
// @accept json
// @produce json
// @param Body body model.MmdUnitTypeInsert true "Create Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/mdm/types/add [post]
func InsertMmdUnitType(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	var req model.MmdUnitTypeInsert
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
	INSERT INTO public."mdm_unit_types"(
	"unitTypeId", "orgId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy")
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id ;
	`

	err := conn.QueryRow(ctx, query,
		uuid, orgId, req.EN, req.TH, req.Active, now, now, username, username).Scan(&id)

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

// @summary Update Mmd Types Sources
// @id Update Mmd Types Sources
// @security ApiKeyAuth
// @accept json
// @produce json
// @tags Mobile device management (Units)
// @Param id path int true "id"
// @param Body body model.MmdUnitTypeUpdate true "Update data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/mdm/types/{id} [patch]
func UpdateMmdUnitType(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	id := c.Param("id")

	var req model.MmdUnitTypeUpdate
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
	query := `UPDATE public."mdm_unit_sources"
	SET en=$2, th=$3, active=$4,
	 "updatedAt"=$5, "updatedBy"=$6
	WHERE id = $1 AND "orgId"=$7`
	_, err := conn.Exec(ctx, query,
		id, req.EN, req.TH, req.Active,
		now, username, orgId,
	)
	logger.Debug("Update Case SQL Args",
		zap.String("query", query),
		zap.Any("Input", []any{
			id, req.EN, req.TH, req.Active,
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

// @summary Delete Mmd Types Sources
// @id Delete Mmd Types Sources
// @security ApiKeyAuth
// @accept json
// @tags Mobile device management (Units)
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/mdm/types/{id} [delete]
func DeleteMmdUnitType(c *gin.Context) {

	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	orgId := GetVariableFromToken(c, "orgId")
	id := c.Param("id")
	query := `DELETE FROM public."mdm_unit_types" WHERE id = $1 AND "orgId"=$2`
	logger.Debug("Query", zap.String("query", query), zap.Any("id", id))
	_, err := conn.Exec(ctx, query, id, orgId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Update failed", zap.Error(err))
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Delete successfully",
	})
}

// @summary Get Mmd Companies
// @tags Mobile device management (Units)
// @security ApiKeyAuth
// @id Get Mmd Companies
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/mdm/companies [get]
func GetMmdCompanies(c *gin.Context) {
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
	// orgId := GetVariableFromToken(c, "orgId")
	query := `SELECT id, name, "legalName", domain, email, "phoneNumber", address, "logoUrl", "websiteUrl", description, "createdAt", "updatedAt", "createdBy", "updatedBy"
	FROM public.mdm_companies LIMIT $1 OFFSET $2`

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
	var org model.MmdCompanies
	var orgList []model.MmdCompanies
	found := false
	for rows.Next() {
		err := rows.Scan(
			&org.ID,
			&org.Name,
			&org.LegalName,
			&org.Domain,
			&org.Email,
			&org.PhoneNumber,
			&org.Address,
			&org.LogoURL,
			&org.WebsiteURL,
			&org.Description,
			&org.CreatedAt,
			&org.UpdatedAt,
			&org.CreatedBy,
			&org.UpdatedBy,
		)
		if err != nil {
			logger.Warn("Scan failed", zap.Error(err))
			response := model.Response{
				Status: "-1",
				Msg:    "Failed",
				Desc:   errorMsg,
			}
			c.JSON(http.StatusInternalServerError, response)
			return
		}
		orgList = append(orgList, org)
		found = true
	}
	if !found {
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
			Data:   orgList,
			Desc:   "",
		}
		c.JSON(http.StatusOK, response)
	}
}

// @summary Get Mmd Companies by Id
// @tags Mobile device management (Units)
// @security ApiKeyAuth
// @id Get Mmd Companies by Id
// @accept json
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/mdm/companies/{id} [get]
func GetMmdCompaniesById(c *gin.Context) {
	logger := config.GetLog()
	id := c.Param("id")
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	// orgId := GetVariableFromToken(c, "orgId")
	defer cancel()
	defer conn.Close(ctx)
	query := `SELECT id, name, "legalName", domain, email, "phoneNumber", address, "logoUrl", "websiteUrl", description, "createdAt", "updatedAt", "createdBy", "updatedBy"
	FROM public.mdm_companies WHERE id=$1`

	var org model.MmdCompanies
	logger.Debug(`Query`, zap.String("query", query))
	err := conn.QueryRow(ctx, query, id).Scan(
		&org.ID,
		&org.Name,
		&org.LegalName,
		&org.Domain,
		&org.Email,
		&org.PhoneNumber,
		&org.Address,
		&org.LogoURL,
		&org.WebsiteURL,
		&org.Description,
		&org.CreatedAt,
		&org.UpdatedAt,
		&org.CreatedBy,
		&org.UpdatedBy,
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
		Data:   org,
		Desc:   "",
	}
	c.JSON(http.StatusOK, response)

}

// @summary Create Mmd Companies Types
// @id Create Mmd Companies Types
// @security ApiKeyAuth
// @tags Mobile device management (Units)
// @accept json
// @produce json
// @param Body body model.MmdCompaniesInsert true "Create Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/mdm/companies/add [post]
func InsertMmdCompanies(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	var req model.MmdCompaniesInsert
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
	// orgId := GetVariableFromToken(c, "orgId")
	uuid := uuid.New()
	query := `
	INSERT INTO public."mdm_companies"(
		id, name, "legalName", domain, email, "phoneNumber", address, "logoUrl", "websiteUrl",
		 description, "createdAt", "updatedAt", "createdBy", "updatedBy")
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	RETURNING id ;
	`

	err := conn.QueryRow(ctx, query,
		uuid, req.Name, req.LegalName, req.Domain, req.Email, req.PhoneNumber, req.Address, req.LogoURL,
		req.WebsiteURL, req.Description, now, now, username, username).Scan(&id)

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

	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Create successfully",
	})

}

// @summary Update Mmd Companies
// @id Update Mmd Companies
// @security ApiKeyAuth
// @accept json
// @produce json
// @tags Mobile device management (Units)
// @Param id path int true "id"
// @param Body body model.MmdCompaniesUpdate true "Update data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/mdm/companies/{id} [patch]
func UpdateMmdCompanies(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	id := c.Param("id")

	var req model.MmdCompaniesUpdate
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
	// orgId := GetVariableFromToken(c, "orgId")
	query := `UPDATE public."mdm_companies"
	SET name= $2, "legalName"= $3, domain= $4, email= $5, "phoneNumber"= $6, address= $7, "logoUrl"= $8,
	 "websiteUrl"= $9, description= $10, "updatedAt"= $11, "updatedBy"= $12
	WHERE id = $1`
	_, err := conn.Exec(ctx, query,
		id, req.Name, req.LegalName, req.Domain, req.Email, req.PhoneNumber, req.Address, req.LogoURL,
		req.WebsiteURL, req.Description, now, username,
	)
	logger.Debug("Update Case SQL Args",
		zap.String("query", query),
		zap.Any("Input", []any{
			id, req.Name, req.LegalName, req.Domain, req.Email, req.PhoneNumber, req.Address, req.LogoURL,
			req.WebsiteURL, req.Description, now, username,
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

// @summary Delete Mmd Companies
// @id Delete Mmd Companies
// @security ApiKeyAuth
// @accept json
// @tags Mobile device management (Units)
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/mdm/companies/{id} [delete]
func DeleteMmdCompanies(c *gin.Context) {

	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	// orgId := GetVariableFromToken(c, "orgId")
	id := c.Param("id")
	query := `DELETE FROM public."mdm_companies" WHERE id = $1`
	logger.Debug("Query", zap.String("query", query), zap.Any("id", id))
	_, err := conn.Exec(ctx, query, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Update failed", zap.Error(err))
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Delete successfully",
	})
}

// @summary Get Mmd Unit Status
// @tags Mobile device management (Units)
// @security ApiKeyAuth
// @id Get Mmd Unit Status
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/mdm/status [get]
func GetMmdUnitStatus(c *gin.Context) {
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
	// orgId := GetVariableFromToken(c, "orgId")
	query := `SELECT id, "sttId", "sttName", "createdAt", "updatedAt", "createdBy", "updatedBy"
	FROM public.mdm_unit_statuses LIMIT $1 OFFSET $2`

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
	var org model.MmdUnitStatus
	var orgList []model.MmdUnitStatus
	found := false
	for rows.Next() {
		err := rows.Scan(
			&org.ID,
			&org.SttID,
			&org.SttName,
			&org.CreatedAt,
			&org.UpdatedAt,
			&org.CreatedBy,
			&org.UpdatedBy,
		)
		if err != nil {
			logger.Warn("Scan failed", zap.Error(err))
			response := model.Response{
				Status: "-1",
				Msg:    "Failed",
				Desc:   errorMsg,
			}
			c.JSON(http.StatusInternalServerError, response)
			return
		}
		orgList = append(orgList, org)
		found = true
	}
	if !found {
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
			Data:   orgList,
			Desc:   "",
		}
		c.JSON(http.StatusOK, response)
	}
}

// @summary Get Mmd Unit Status by Id
// @tags Mobile device management (Units)
// @security ApiKeyAuth
// @id Get Mmd Unit Status by Id
// @accept json
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/mdm/status/{id} [get]
func GetMmdUnitStatusById(c *gin.Context) {
	logger := config.GetLog()
	id := c.Param("id")
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	// orgId := GetVariableFromToken(c, "orgId")
	defer cancel()
	defer conn.Close(ctx)
	query := `SELECT id, "sttId", "sttName", "createdAt", "updatedAt", "createdBy", "updatedBy"
	FROM public.mdm_unit_statuses WHERE id=$1`

	var org model.MmdUnitStatus
	logger.Debug(`Query`, zap.String("query", query))
	err := conn.QueryRow(ctx, query, id).Scan(
		&org.ID,
		&org.SttID,
		&org.SttName,
		&org.CreatedAt,
		&org.UpdatedAt,
		&org.CreatedBy,
		&org.UpdatedBy,
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
		Data:   org,
		Desc:   "",
	}
	c.JSON(http.StatusOK, response)

}

// @summary Create Mmd Unit Status
// @id Create Mmd Unit Status
// @security ApiKeyAuth
// @tags Mobile device management (Units)
// @accept json
// @produce json
// @param Body body model.MmdUnitStatusInsert true "Create Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/mdm/status/add [post]
func InsertMmdUnitStatus(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	var req model.MmdUnitStatusInsert
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
	// orgId := GetVariableFromToken(c, "orgId")
	uuid := uuid.New()
	query := `
	INSERT INTO public."mdm_unit_statuses"(
		"sttId", "sttName", "createdAt", "updatedAt", "createdBy", "updatedBy")
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id ;
	`

	err := conn.QueryRow(ctx, query,
		uuid, req.SttID, req.SttName, now, now, username, username).Scan(&id)

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

	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Create successfully",
	})

}

// @summary Update Mmd Unit Status
// @id Update Mmd Unit Status
// @security ApiKeyAuth
// @accept json
// @produce json
// @tags Mobile device management (Units)
// @Param id path int true "id"
// @param Body body model.MmdUnitStatusUpdate true "Update data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/mdm/status/{id} [patch]
func UpdateMmdUnitStatus(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	id := c.Param("id")

	var req model.MmdUnitStatusUpdate
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
	// orgId := GetVariableFromToken(c, "orgId")
	query := `UPDATE public."mdm_unit_statuses"
	SET sttId= $2, "sttName"= $3, "updatedAt"= $4, "updatedBy"= $5
	WHERE id = $1`
	_, err := conn.Exec(ctx, query,
		id, req.SttID, req.SttName, now, username,
	)
	logger.Debug("Update Case SQL Args",
		zap.String("query", query),
		zap.Any("Input", []any{
			id, req.SttID, req.SttName, now, username,
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

// @summary Delete Mmd Unit Status
// @id Delete Mmd Unit Status
// @security ApiKeyAuth
// @accept json
// @tags Mobile device management (Units)
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/mdm/status/{id} [delete]
func DeleteMmdUnitStatus(c *gin.Context) {

	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	// orgId := GetVariableFromToken(c, "orgId")
	id := c.Param("id")
	query := `DELETE FROM public."mdm_unit_statuses" WHERE id = $1`
	logger.Debug("Query", zap.String("query", query), zap.Any("id", id))
	_, err := conn.Exec(ctx, query, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Update failed", zap.Error(err))
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Delete successfully",
	})
}

// @summary Get Mmd Unit
// @tags Mobile device management (Units)
// @security ApiKeyAuth
// @id Get Mmd Unit
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/mdm/units [get]
func GetMmdUnit(c *gin.Context) {
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
	query := `SELECT id, "orgId", "unitId", "unitName", "unitSourceId", "unitTypeId", priority, "compId", "deptId", "commId", "stnId", "plateNo", "provinceCode", active, username, "isLogin", "isFreeze", "isOutArea", "locLat", "locLon", "locAlt", "locBearing", "locSpeed", "locProvider", "locGpsTime", "locSatellites", "locAccuracy", "locLastUpdateTime", "breakDuration", "healthChk", "healthChkTime", "sttId", "createdAt", "updatedAt", "createdBy", "updatedBy"
	FROM public.mdm_units WHERE "orgId"= $3 LIMIT $1 OFFSET $2`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	rows, err = conn.Query(ctx, query, length, start, orgId)
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
	var unit model.MmdUnit
	var unitList []model.MmdUnit
	found := false
	for rows.Next() {
		err := rows.Scan(
			&unit.ID,
			&unit.OrgID,
			&unit.UnitID,
			&unit.UnitName,
			&unit.UnitSourceID,
			&unit.UnitTypeID,
			&unit.Priority,
			&unit.CompID,
			&unit.DeptID,
			&unit.CommID,
			&unit.StnID,
			&unit.PlateNo,
			&unit.ProvinceCode,
			&unit.Active,
			&unit.Username,
			&unit.IsLogin,
			&unit.IsFreeze,
			&unit.IsOutArea,
			&unit.LocLat,
			&unit.LocLon,
			&unit.LocAlt,
			&unit.LocBearing,
			&unit.LocSpeed,
			&unit.LocProvider,
			&unit.LocGpsTime,
			&unit.LocSatellites,
			&unit.LocAccuracy,
			&unit.LocLastUpdateTime,
			&unit.BreakDuration,
			&unit.HealthChk,
			&unit.HealthChkTime,
			&unit.SttID,
			&unit.CreatedAt,
			&unit.UpdatedAt,
			&unit.CreatedBy,
			&unit.UpdatedBy,
		)
		if err != nil {
			logger.Warn("Scan failed", zap.Error(err))
			response := model.Response{
				Status: "-1",
				Msg:    "Failed",
				Desc:   errorMsg,
			}
			c.JSON(http.StatusInternalServerError, response)
			return
		}
		unitList = append(unitList, unit)
		found = true
	}
	if !found {
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
			Data:   unitList,
			Desc:   "",
		}
		c.JSON(http.StatusOK, response)
	}
}

// @summary Get Mmd Unit by Id
// @tags Mobile device management (Units)
// @security ApiKeyAuth
// @id Get Mmd Unit by Id
// @accept json
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/mdm/units/{id} [get]
func GetMmdUnitById(c *gin.Context) {
	logger := config.GetLog()
	id := c.Param("id")
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	orgId := GetVariableFromToken(c, "orgId")
	defer cancel()
	defer conn.Close(ctx)
	query := `SELECT id, "orgId", "unitId", "unitName", "unitSourceId", "unitTypeId", priority, "compId", "deptId", "commId", "stnId", "plateNo", "provinceCode", active, username, "isLogin", "isFreeze", "isOutArea", "locLat", "locLon", "locAlt", "locBearing", "locSpeed", "locProvider", "locGpsTime", "locSatellites", "locAccuracy", "locLastUpdateTime", "breakDuration", "healthChk", "healthChkTime", "sttId", "createdAt", "updatedAt", "createdBy", "updatedBy"
	FROM public.mdm_units WHERE id=$1 AND "orgId"=$2`

	var unit model.MmdUnit
	logger.Debug(`Query`, zap.String("query", query))
	err := conn.QueryRow(ctx, query, id, orgId).Scan(
		&unit.ID,
		&unit.OrgID,
		&unit.UnitID,
		&unit.UnitName,
		&unit.UnitSourceID,
		&unit.UnitTypeID,
		&unit.Priority,
		&unit.CompID,
		&unit.DeptID,
		&unit.CommID,
		&unit.StnID,
		&unit.PlateNo,
		&unit.ProvinceCode,
		&unit.Active,
		&unit.Username,
		&unit.IsLogin,
		&unit.IsFreeze,
		&unit.IsOutArea,
		&unit.LocLat,
		&unit.LocLon,
		&unit.LocAlt,
		&unit.LocBearing,
		&unit.LocSpeed,
		&unit.LocProvider,
		&unit.LocGpsTime,
		&unit.LocSatellites,
		&unit.LocAccuracy,
		&unit.LocLastUpdateTime,
		&unit.BreakDuration,
		&unit.HealthChk,
		&unit.HealthChkTime,
		&unit.SttID,
		&unit.CreatedAt,
		&unit.UpdatedAt,
		&unit.CreatedBy,
		&unit.UpdatedBy,
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
		Data:   unit,
		Desc:   "",
	}
	c.JSON(http.StatusOK, response)

}

// @summary Create Mmd Unit
// @id Create Mmd Unit
// @security ApiKeyAuth
// @tags Mobile device management (Units)
// @accept json
// @produce json
// @param Body body model.MmdUnitInsert true "Create Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/mdm/units/add [post]
func InsertMmdUnit(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	var unit model.MmdUnitInsert
	if err := c.ShouldBindJSON(&unit); err != nil {
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
	query := `INSERT INTO public.mdm_units(
	"orgId", "unitId", "unitName", "unitSourceId", "unitTypeId", priority, "compId", "deptId", "commId",
	 "stnId", "plateNo", "provinceCode", active, username, "isLogin", "isFreeze", "isOutArea", "locLat",
	  "locLon", "locAlt", "locBearing", "locSpeed", "locProvider", "locGpsTime", "locSatellites", "locAccuracy",
	   "locLastUpdateTime", "breakDuration", "healthChk", "healthChkTime", "sttId", "createdAt", "updatedAt",
	    "createdBy", "updatedBy")
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
	 $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29,
	  $30, $31, $32, $33, $34)
	RETURNING id ;
	`

	err := conn.QueryRow(ctx, query,
		orgId,
		unit.UnitID,
		unit.UnitName,
		unit.UnitSourceID,
		unit.UnitTypeID,
		unit.Priority,
		unit.CompID,
		unit.DeptID,
		unit.CommID,
		unit.StnID,
		unit.PlateNo,
		unit.ProvinceCode,
		unit.Active,
		unit.Username,
		unit.IsLogin,
		unit.IsFreeze,
		unit.IsOutArea,
		unit.LocLat,
		unit.LocLon,
		unit.LocAlt,
		unit.LocBearing,
		unit.LocSpeed,
		unit.LocProvider,
		unit.LocGpsTime,
		unit.LocSatellites,
		unit.LocAccuracy,
		unit.LocLastUpdateTime,
		unit.BreakDuration,
		unit.HealthChk,
		unit.HealthChkTime,
		unit.SttID, now, now, username, username).Scan(&id)

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

	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Create successfully",
	})

}

// @summary Update Mmd Unit
// @id Update Mmd Unit
// @security ApiKeyAuth
// @accept json
// @produce json
// @tags Mobile device management (Units)
// @Param id path int true "id"
// @param Body body model.MmdUnitUpdate true "Update data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/mdm/units/{id} [patch]
func UpdateMmdUnit(c *gin.Context) {
	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	id := c.Param("id")

	var req model.MmdUnitUpdate
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
	query := `UPDATE public."mdm_units"
	SET "unitId"=$2, "unitName"=$3, "unitSourceId"=$4, "unitTypeId"=$5, priority=$6, "compId"=$7,
	 "deptId"=$8, "commId"=$9, "stnId"=$10, "plateNo"=$11, "provinceCode"=$12, active=$13, username=$14,
	  "isLogin"=$15, "isFreeze"=$16, "isOutArea"=$17, "locLat"=$18, "locLon"=$19, "locAlt"=$20, "locBearing"=$21,
	   "locSpeed"=$22, "locProvider"=$23, "locGpsTime"=$24, "locSatellites"=$25, "locAccuracy"=$26,
	    "locLastUpdateTime"=$27, "breakDuration"=$28, "healthChk"=$29, "healthChkTime"=$30, "sttId"=$31,
		 "updatedAt"=$32, "updatedBy"=$33
	WHERE id = $1 AND "orgId"=$34`
	_, err := conn.Exec(ctx, query,
		id, req.UnitID, req.UnitName, req.UnitSourceID, req.UnitTypeID, req.Priority, req.CompID,
		req.DeptID, req.CommID, req.StnID, req.PlateNo, req.ProvinceCode, req.Active, req.Username,
		req.IsLogin, req.IsFreeze, req.IsOutArea, req.LocLat, req.LocLon, req.LocAlt, req.LocBearing,
		req.LocSpeed, req.LocProvider, req.LocGpsTime, req.LocSatellites, req.LocAccuracy,
		req.LocLastUpdateTime, req.BreakDuration, req.HealthChk, req.HealthChkTime, req.SttID, now, username, orgId,
	)
	logger.Debug("Update Case SQL Args",
		zap.String("query", query),
		zap.Any("Input", []any{
			id, req.UnitID, req.UnitName, req.UnitSourceID, req.UnitTypeID, req.Priority, req.CompID,
			req.DeptID, req.CommID, req.StnID, req.PlateNo, req.ProvinceCode, req.Active, req.Username,
			req.IsLogin, req.IsFreeze, req.IsOutArea, req.LocLat, req.LocLon, req.LocAlt, req.LocBearing,
			req.LocSpeed, req.LocProvider, req.LocGpsTime, req.LocSatellites, req.LocAccuracy,
			req.LocLastUpdateTime, req.BreakDuration, req.HealthChk, req.HealthChkTime, req.SttID, now, username, orgId,
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

// @summary Delete Mmd Unit
// @id Delete Mmd Unit
// @security ApiKeyAuth
// @accept json
// @tags Mobile device management (Units)
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/mdm/units/{id} [delete]
func DeleteMmdUnit(c *gin.Context) {

	logger := config.GetLog()
	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	orgId := GetVariableFromToken(c, "orgId")
	id := c.Param("id")
	query := `DELETE FROM public."mdm_units" WHERE id = $1 AND "orgId"=$2`
	logger.Debug("Query", zap.String("query", query), zap.Any("id", id))
	_, err := conn.Exec(ctx, query, id, orgId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})
		logger.Warn("Update failed", zap.Error(err))
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Delete successfully",
	})
}

// @summary Get Mmd Unit With Property
// @tags Mobile device management (Units)
// @security ApiKeyAuth
// @id Get Mmd Unit With Property
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @Param unitId path string true "unitId"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/mdm/units/properties/{unitId} [get]
func GetMmdUnitWithProperty(c *gin.Context) {
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
	unitId := c.Param("unitId")
	orgId := GetVariableFromToken(c, "orgId")
	query := `SELECT id, "orgId", "unitId", "propId", active, "createdAt", "updatedAt", "createdBy", "updatedBy"
	FROM public.mdm_unit_with_properties WHERE "orgId"= $3 AND "unitId"=$4 LIMIT $1 OFFSET $2`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
	rows, err = conn.Query(ctx, query, length, start, orgId, unitId)
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
	var unit model.MdmUnitProperty
	var unitList []model.MdmUnitProperty
	found := false
	for rows.Next() {
		err := rows.Scan(
			&unit.ID,
			&unit.OrgID,
			&unit.UnitID,
			&unit.PropID,
			&unit.Active,
			&unit.CreatedAt,
			&unit.UpdatedAt,
			&unit.CreatedBy,
			&unit.UpdatedBy,
		)
		if err != nil {
			logger.Warn("Scan failed", zap.Error(err))
			response := model.Response{
				Status: "-1",
				Msg:    "Failed",
				Desc:   errorMsg,
			}
			c.JSON(http.StatusInternalServerError, response)
			return
		}
		unitList = append(unitList, unit)
		found = true
	}
	if !found {
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
			Data:   unitList,
			Desc:   "",
		}
		c.JSON(http.StatusOK, response)
	}
}
