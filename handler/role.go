package handler

import (
	"github.com/gin-gonic/gin"
)

// ListUserSkill godoc
// @summary Get UserSkill
// @tags Dispatch
// @security ApiKeyAuth
// @id Get UserSkill
// @accept json
// @produce json
// @Param start query int false "start" default(0)
// @Param length query int false "length" default(10)
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/departments [get]
func GetUserSkill(c *gin.Context) {
	// logger := config.GetLog()
	// orgId := GetVariableFromToken(c, "orgId")
	// conn, ctx, cancel := config.ConnectDB()
	// if conn == nil {
	// 	return
	// }
	// defer cancel()
	// defer conn.Close(ctx)
	// startStr := c.DefaultQuery("start", "0")
	// start, err := strconv.Atoi(startStr)
	// if err != nil {
	// 	start = 0
	// }
	// lengthStr := c.DefaultQuery("length", "1000")
	// length, err := strconv.Atoi(lengthStr)
	// if err != nil {
	// 	length = 1000
	// }
	// query := `SELECT id,"orgId", "skillId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy"
	// FROM public.sec_departments WHERE "orgId"=$1 LIMIT $2 OFFSET $3`

	// var rows pgx.Rows
	// logger.Debug(`Query`, zap.String("query", query))
	// rows, err = conn.Query(ctx, query, orgId, length, start)
	// if err != nil {
	// 	logger.Warn("Query failed", zap.Error(err))
	// 	c.JSON(http.StatusInternalServerError, model.Response{
	// 		Status: "-1",
	// 		Msg:    "Failure",
	// 		Desc:   err.Error(),
	// 	})
	// 	return
	// }
	// defer rows.Close()
	// var errorMsg string
	// var UserSkill model.UserSkill
	// var UserSkillList []model.UserSkill
	// rowIndex := 0
	// for rows.Next() {
	// 	rowIndex++
	// 	err := rows.Scan(&UserSkill.ID, &UserSkill.DeptID, &UserSkill.OrgID, &UserSkill.En, &UserSkill.Th,
	// 		&UserSkill.Active, &UserSkill.CreatedAt, &UserSkill.UpdatedAt, &UserSkill.CreatedBy, &UserSkill.UpdatedBy)
	// 	if err != nil {
	// 		logger.Warn("Scan failed", zap.Error(err))
	// 		response := model.Response{
	// 			Status: "-1",
	// 			Msg:    "Failed",
	// 			Desc:   errorMsg,
	// 		}
	// 		c.JSON(http.StatusInternalServerError, response)
	// 	}
	// 	UserSkillList = append(UserSkillList, UserSkill)
	// }
	// if errorMsg != "" {
	// 	response := model.Response{
	// 		Status: "-1",
	// 		Msg:    "Failed",
	// 		Desc:   errorMsg,
	// 	}
	// 	c.JSON(http.StatusInternalServerError, response)
	// } else {
	// 	response := model.Response{
	// 		Status: "0",
	// 		Msg:    "Success",
	// 		Data:   UserSkillList,
	// 		Desc:   "",
	// 	}
	// 	c.JSON(http.StatusOK, response)
	// }
}

// ListUserSkill godoc
// @summary Get UserSkill by ID
// @tags Dispatch
// @security ApiKeyAuth
// @id Get UserSkill by ID
// @accept json
// @produce json
// @Param id path string true "id" "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/departments/{id} [get]
func GetUserSkillbyId(c *gin.Context) {
	// logger := config.GetLog()
	// id := c.Param("id")
	// conn, ctx, cancel := config.ConnectDB()
	// if conn == nil {
	// 	return
	// }
	// defer cancel()
	// defer conn.Close(ctx)
	// orgId := GetVariableFromToken(c, "orgId")
	// query := `SELECT id,"deptId", "orgId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy"
	// FROM public.sec_departments WHERE "deptId" = $1 AND "orgId"=$2`

	// var rows pgx.Rows
	// logger.Debug(`Query`, zap.String("query", query))
	// rows, err := conn.Query(ctx, query, id, orgId)
	// if err != nil {
	// 	logger.Warn("Query failed", zap.Error(err))
	// 	c.JSON(http.StatusInternalServerError, model.Response{
	// 		Status: "-1",
	// 		Msg:    "Failure",
	// 		Desc:   err.Error(),
	// 	})
	// 	return
	// }
	// defer rows.Close()
	// var errorMsg string
	// var UserSkill model.UserSkill
	// err = rows.Scan(&UserSkill.ID, &UserSkill.DeptID, &UserSkill.OrgID, &UserSkill.En, &UserSkill.Th,
	// 	&UserSkill.Active, &UserSkill.CreatedAt, &UserSkill.UpdatedAt, &UserSkill.CreatedBy, &UserSkill.UpdatedBy)
	// if err != nil {
	// 	logger.Warn("Scan failed", zap.Error(err))
	// 	errorMsg = err.Error()
	// 	response := model.Response{
	// 		Status: "-1",
	// 		Msg:    "Failed",
	// 		Desc:   errorMsg,
	// 	}
	// 	c.JSON(http.StatusInternalServerError, response)
	// 	return
	// }

	// if errorMsg != "" {
	// 	response := model.Response{
	// 		Status: "-1",
	// 		Msg:    "Failed",
	// 		Desc:   errorMsg,
	// 	}
	// 	c.JSON(http.StatusInternalServerError, response)
	// } else {
	// 	response := model.Response{
	// 		Status: "0",
	// 		Msg:    "Success",
	// 		Data:   UserSkill,
	// 		Desc:   "",
	// 	}
	// 	c.JSON(http.StatusOK, response)
	// }
}

// @summary Create UserSkill
// @id Create UserSkill
// @security ApiKeyAuth
// @tags Dispatch
// @accept json
// @produce json
// @param Body body model.UserSkillInsert true "Create Data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/departments/add [post]
func InsertUserSkill(c *gin.Context) {
	// logger := config.GetLog()
	// conn, ctx, cancel := config.ConnectDB()
	// if conn == nil {
	// 	return
	// }
	// defer cancel()
	// defer conn.Close(ctx)
	// defer cancel()

	// var req model.UserSkillInsert
	// if err := c.ShouldBindJSON(&req); err != nil {
	// 	c.JSON(http.StatusBadRequest, model.Response{
	// 		Status: "-1",
	// 		Msg:    "Failure",
	// 		Desc:   err.Error(),
	// 	})
	// 	logger.Warn("Insert failed", zap.Error(err))
	// 	return
	// }
	// username := GetVariableFromToken(c, "username")
	// orgId := GetVariableFromToken(c, "orgId")
	// now := time.Now()
	// uuid := uuid.New()
	// var id int
	// query := `
	// INSERT INTO public."sec_departments"(
	// "deptId", "orgId", en, th, active, "createdAt", "updatedAt", "createdBy", "updatedBy")
	// VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	// RETURNING id ;
	// `

	// err := conn.QueryRow(ctx, query,
	// 	uuid, orgId, req.En, req.Th, req.Active, now,
	// 	now, username, username).Scan(&id)

	// if err != nil {
	// 	// log.Printf("Insert failed: %v", err)
	// 	c.JSON(http.StatusInternalServerError, model.Response{
	// 		Status: "-1",
	// 		Msg:    "Failure",
	// 		Desc:   err.Error(),
	// 	})
	// 	logger.Warn("Insert failed", zap.Error(err))
	// 	return
	// }

	// // Continue logic...
	// c.JSON(http.StatusOK, model.Response{
	// 	Status: "0",
	// 	Msg:    "Success",
	// 	Desc:   "Create successfully",
	// })

}

// @summary Update UserSkill
// @id Update UserSkill
// @security ApiKeyAuth
// @accept json
// @produce json
// @tags Dispatch
// @Param id path int true "id"
// @param Body body model.UserSkillUpdate true "Update data"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/departments/{id} [patch]
func UpdateUserSkill(c *gin.Context) {
	// logger := config.GetLog()
	// conn, ctx, cancel := config.ConnectDB()
	// if conn == nil {
	// 	return
	// }
	// defer cancel()
	// defer conn.Close(ctx)
	// defer cancel()

	// id := c.Param("id")

	// var req model.UserSkillUpdate
	// if err := c.ShouldBindJSON(&req); err != nil {
	// 	logger.Warn("Update failed", zap.Error(err))
	// 	c.JSON(http.StatusBadRequest, model.Response{
	// 		Status: "-1",
	// 		Msg:    "Failure",
	// 		Desc:   err.Error(),
	// 	})
	// 	return
	// }
	// now := time.Now()
	// username := GetVariableFromToken(c, "username")
	// orgId := GetVariableFromToken(c, "orgId")
	// query := `UPDATE public."sec_departments"
	// SET en=$2, th=$3, active=$4,
	//  "updatedAt"=$5, "updatedBy"=$6
	// WHERE id = $1 AND "orgId"=$7`
	// _, err := conn.Exec(ctx, query,
	// 	id, req.En, req.Th, req.Active,
	// 	now, username, orgId,
	// )
	// logger.Debug("Update Case SQL Args",
	// 	zap.String("query", query),
	// 	zap.Any("Input", []any{
	// 		id, req.En, req.Th, req.Active,
	// 		now, username, orgId,
	// 	}))
	// if err != nil {
	// 	// log.Printf("Insert failed: %v", err)
	// 	c.JSON(http.StatusInternalServerError, model.Response{
	// 		Status: "-1",
	// 		Msg:    "Failure",
	// 		Desc:   err.Error(),
	// 	})
	// 	logger.Warn("Update failed", zap.Error(err))
	// 	return
	// }

	// // Continue logic...
	// c.JSON(http.StatusOK, model.Response{
	// 	Status: "0",
	// 	Msg:    "Success",
	// 	Desc:   "Update successfully",
	// })
}

// @summary Delete UserSkill
// @id Delete UserSkill
// @security ApiKeyAuth
// @accept json
// @tags Dispatch
// @produce json
// @Param id path int true "id"
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/departments/{id} [delete]
func DeleteUserSkill(c *gin.Context) {

	// logger := config.GetLog()
	// conn, ctx, cancel := config.ConnectDB()
	// if conn == nil {
	// 	return
	// }
	// defer cancel()
	// defer conn.Close(ctx)
	// defer cancel()
	// orgId := GetVariableFromToken(c, "orgId")
	// id := c.Param("id")
	// query := `DELETE FROM public."sec_departments" WHERE id = $1 AND "orgId"=$2`
	// logger.Debug("Query", zap.String("query", query), zap.Any("id", id))
	// _, err := conn.Exec(ctx, query, id, orgId)
	// if err != nil {
	// 	// log.Printf("Insert failed: %v", err)
	// 	c.JSON(http.StatusInternalServerError, model.Response{
	// 		Status: "-1",
	// 		Msg:    "Failure",
	// 		Desc:   err.Error(),
	// 	})
	// 	logger.Warn("Update failed", zap.Error(err))
	// 	return
	// }

	// // Continue logic...
	// c.JSON(http.StatusOK, model.Response{
	// 	Status: "0",
	// 	Msg:    "Success",
	// 	Desc:   "Delete successfully",
	// })
}
