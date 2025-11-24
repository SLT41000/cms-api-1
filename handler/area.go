package handler

import (
	"mainPackage/model"
	"mainPackage/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// @summary Get Country Province Districts
// @tags Area
// @security ApiKeyAuth
// @id Get Country Province Districts
// @accept json
// @produce json
// @response 200 {object} model.Response "OK - Request successful"
// @Router /api/v1/area/country_province_districts [get]
func GetCountryProvinceDistricts(c *gin.Context) {
	//logger := utils.GetLog()
	startTime := time.Now()

	conn, ctx, cancel := utils.ConnectDB()
	if conn == nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failure",
			Desc:   "Database connection failed",
		})
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	txtId := uuid.New().String()

	result, err := utils.GetCountryProvinceDistrictsOrLoad(ctx, conn, orgId.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   err.Error(),
		})
		return
	}

	// Success
	response := model.Response{
		Status: "0",
		Msg:    "Success",
		Data:   result,
	}

	_ = utils.InsertAuditLogs(
		c, conn, orgId.(string), username.(string),
		txtId, "", "Area", "GetCountryProvinceDistricts", "",
		"search", 0, startTime, GetQueryParams(c), response, "Get successfully",
	)

	c.JSON(http.StatusOK, response)
}
