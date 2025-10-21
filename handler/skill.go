package handler

import (
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

// ListSkill godoc
// @summary Get Skill
// @tags User
// @security ApiKeyAuth
// @id Get Skill
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/skill [get]
func GetSkill(c *gin.Context) {
	logger := utils.GetLog()
	id := c.Param("id")
	now := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()

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
	query := `SELECT id,"orgId", "skillId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy"
	FROM public.um_skills WHERE "orgId"=$1 LIMIT $2 OFFSET $3`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query))
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
			txtId, id, "Skill", "GetSkill", "",
			"search", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	defer rows.Close()
	var errorMsg string
	var Skill model.Skill
	var SkillList []model.Skill
	rowIndex := 0
	for rows.Next() {
		rowIndex++
		err := rows.Scan(&Skill.ID, &Skill.OrgID, &Skill.SkillID, &Skill.En, &Skill.Th,
			&Skill.Active, &Skill.CreatedAt, &Skill.UpdatedAt, &Skill.CreatedBy, &Skill.UpdatedBy)
		if err != nil {
			logger.Warn("Scan failed", zap.Error(err))
			response := model.Response{
				Status: "-1",
				Msg:    "Failed",
				Desc:   errorMsg,
			}
			//=======AUDIT_START=====//
			_ = utils.InsertAuditLogs(
				c, conn, orgId.(string), username.(string),
				txtId, id, "Skill", "GetSkill", "",
				"search", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
			)
			//=======AUDIT_END=====//
			c.JSON(http.StatusInternalServerError, response)
		}
		SkillList = append(SkillList, Skill)
	}
	if errorMsg != "" {
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   errorMsg,
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Skill", "GetSkill", "",
			"search", -1, now, GetQueryParams(c), response, "Failed : "+errorMsg,
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
	} else {
		response := model.Response{
			Status: "0",
			Msg:    "Success",
			Data:   SkillList,
			Desc:   "",
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Skill", "GetSkill", "",
			"search", 0, now, GetQueryParams(c), response, "GetSkill Success",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusOK, response)
	}
}

// ListSkill godoc
// @summary Get Skill by ID
// @tags User
// @security ApiKeyAuth
// @id Get Skill by ID
// @accept json
// @produce json
// @Param id path string true "id" "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/skill/{id} [get]
func GetSkillbyId(c *gin.Context) {
	logger := utils.GetLog()
	id := c.Param("id")
	now := time.Now()
	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	query := `SELECT id,"orgId", "skillId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy"
	FROM public.um_skills WHERE "skillId" = $1 AND "orgId"=$2`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query),
		zap.Any("Input", []any{
			id, orgId,
		}))
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
			txtId, id, "Skill", "GetSkillById", "",
			"search", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		})

		return
	}
	defer rows.Close()
	var errorMsg string
	var Skill model.Skill
	err = rows.Scan(&Skill.ID, &Skill.OrgID, &Skill.SkillID, &Skill.En, &Skill.Th,
		&Skill.Active, &Skill.CreatedAt, &Skill.UpdatedAt, &Skill.CreatedBy, &Skill.UpdatedBy)
	if err != nil {
		logger.Warn("Scan failed", zap.Error(err))
		errorMsg = err.Error()
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   errorMsg,
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Skill", "GetSkillById", "",
			"search", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	if errorMsg != "" {
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   errorMsg,
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Skill", "GetSkillById", "",
			"search", -1, now, GetQueryParams(c), response, "Failed : "+errorMsg,
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
	} else {
		response := model.Response{
			Status: "0",
			Msg:    "Success",
			Data:   Skill,
			Desc:   "",
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Skill", "GetSkillById", "",
			"search", -1, now, GetQueryParams(c), response, "GetSkillbyId Success",
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusOK, response)
	}
}

// @summary Create Skill
// @id Create Skill
// @security ApiKeyAuth
// @tags User
// @accept json
// @produce json
// @param Body body model.SkillInsert true "Create Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/skill/add [post]
func InsertSkill(c *gin.Context) {
	logger := utils.GetLog()
	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)
	defer cancel()

	var req model.SkillInsert
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
	INSERT INTO public."um_skills"(
	"orgId", "skillId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy")
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id ;
	`

	err := conn.QueryRow(ctx, query,
		orgId, uuid, req.En, req.Th, req.Active, now,
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
			uuid.String(), strconv.Itoa(id), "Skill", "InsertSkill", "",
			"create", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusInternalServerError, response)
		logger.Warn("Insert failed", zap.Error(err))
		return
	}

	// Continue logic...
	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Desc:   "Create successfully",
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		uuid.String(), strconv.Itoa(id), "Skill", "InsertSkill", "",
		"create", 0, now, GetQueryParams(c), response, "InsertSkill Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)

}

// @summary Update Skill
// @id Update Skill
// @security ApiKeyAuth
// @accept json
// @produce json
// @tags User
// @Param id path int true "id"
// @param Body body model.SkillUpdate true "Update data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/skill/{id} [patch]
func UpdateSkill(c *gin.Context) {
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

	var req model.SkillUpdate
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
			txtId, id, "Skill", "UpdateSkill", "",
			"update", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
		)
		//=======AUDIT_END=====//
		c.JSON(http.StatusBadRequest, response)
		return
	}

	query := `UPDATE public."um_skills"
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
		response := model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   err.Error(),
		}
		//=======AUDIT_START=====//
		_ = utils.InsertAuditLogs(
			c, conn, orgId.(string), username.(string),
			txtId, id, "Skill", "UpdateSkill", "",
			"update", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
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
		Desc:   "Update successfully",
	}
	//=======AUDIT_START=====//
	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, id, "Skill", "UpdateSkill", "",
		"update", 0, now, GetQueryParams(c), response, "UpdateSkill Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}

// @summary Delete Skill
// @id Delete Skill
// @security ApiKeyAuth
// @accept json
// @tags User
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/skill/{id} [delete]
func DeleteSkill(c *gin.Context) {

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
	query := `DELETE FROM public."um_skills" WHERE id = $1 AND "orgId"=$2`
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
			txtId, id, "Skill", "DeleteSkill", "",
			"delete", -1, now, GetQueryParams(c), response, "Failed : "+err.Error(),
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
		txtId, id, "Skill", "DeleteSkill", "",
		"delete", 0, now, GetQueryParams(c), response, "DeleteSkill Success.",
	)
	//=======AUDIT_END=====//
	c.JSON(http.StatusOK, response)
}
