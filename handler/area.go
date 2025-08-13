package handler

import (
	"mainPackage/config"
	"mainPackage/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
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
	logger := config.GetLog()

	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	orgId := GetVariableFromToken(c, "orgId")
	query := `SELECT t1.id, t1."orgId", t1."countryId", t1."provId", t1."distId",
	 	t1.en, t1.th, t1.active,
	  	t2.en, t2.th, t2.active,
	  	t3.en, t3.th, t3.active
		FROM public.area_districts t1
		FULL JOIN public.area_provinces t2 ON t1."provId" = t2."provId"
		FULL JOIN public.area_countries t3 ON t2."countryId" = t3."countryId"
	WHERE (t1."orgId" = $1 OR t1."orgId" IS NULL) ORDER by t1."countryId"  NULLS LAST;`

	var rows pgx.Rows
	logger.Debug(`Query`, zap.String("query", query), zap.Any("orgId", orgId))
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
	var Area model.AreaDistrictWithDetails
	var AreaList []model.AreaDistrictWithDetails
	found := false
	for rows.Next() {
		err := rows.Scan(&Area.ID, &Area.OrgID, &Area.CountryID, &Area.ProvID, &Area.DistID,
			&Area.DistrictEn, &Area.DistrictTh, &Area.DistrictActive,
			&Area.ProvinceEn, &Area.ProvinceTh, &Area.ProvinceActive,
			&Area.CountryEn, &Area.CountryTh, &Area.CountryActive)
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
		AreaList = append(AreaList, Area)
		found = true
	}
	if !found {
		response := model.Response{
			Status: "-1",
			Msg:    "Failed",
			Desc:   "Not found",
		}
		c.JSON(http.StatusInternalServerError, response)
	} else {
		response := model.Response{
			Status: "0",
			Msg:    "Success",
			Data:   AreaList,
			Desc:   "",
		}
		c.JSON(http.StatusOK, response)
	}
}
