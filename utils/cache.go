package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"mainPackage/model"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func GetSubTypeByID(ctx context.Context, conn *pgx.Conn, orgId string, DeviceType string, WorlkOrderType string) (*model.WorkflowBySubType, error) {
	query := `
		SELECT 
    cs."id", cs."typeId", cs."sTypeId", cs."sTypeCode", cs."orgId",
    cs."en", cs."th", cs."wfId", cs."caseSla", cs."priority",
    cs."userSkillList", cs."unitPropLists", cs."active",
    cs."createdAt", cs."updatedAt", cs."createdBy", cs."updatedBy",
    wf."title"       AS wfTitle,
    wf."desc"        AS wfDesc,
    wf."versions"    AS wfVersions,
    wn."section"     AS wfSection,
    wn."data"        AS wfData,
    wn."nodeId"      AS wfNodeId
FROM public.case_sub_types cs
LEFT JOIN public.wf_definitions wf
       ON cs."wfId" = wf."wfId" 
      AND cs."orgId" = wf."orgId"
LEFT JOIN public.wf_nodes wn
       ON wf."wfId" = wn."wfId"
      AND wf."versions" = wn."versions"
      AND LOWER(wn."type") = 'process'
      AND wn."data"->'data'->'config'->>'action' = 'S001'
WHERE cs."orgId" = $1 
  AND cs."mDeviceType" = $2
  AND cs."mWorkOrderType" = $3
  AND cs."active" = TRUE
LIMIT 1;
	`

	var subType model.WorkflowBySubType
	err := conn.QueryRow(ctx, query, orgId, DeviceType, WorlkOrderType).Scan(
		&subType.Id,
		&subType.TypeID,
		&subType.STypeID,
		&subType.STypeCode,
		&subType.OrgID,
		&subType.EN,
		&subType.TH,
		&subType.WFID,
		&subType.CaseSLA,
		&subType.Priority,
		&subType.UserSkillList,
		&subType.UnitPropLists,
		&subType.Active,
		&subType.CreatedAt,
		&subType.UpdatedAt,
		&subType.CreatedBy,
		&subType.UpdatedBy,
		&subType.WfTitle,
		&subType.WfDesc,
		&subType.WfVersions,
		&subType.WfSection,
		&subType.WfData,
		&subType.WfNodeId,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // not found
		}
		return nil, err
	}

	return &subType, nil
}

func GetUserByUsername(ctx context.Context, conn *pgx.Conn, orgId, username string) (*model.User, error) {

	//Get Cache
	val, err := UsernameGet(username)
	if err != nil {
		log.Println("Redis GET error:", err)
	} else if val != "" {
		var cachedUser model.User
		if err := json.Unmarshal([]byte(val), &cachedUser); err == nil {
			return &cachedUser, nil
		}
	}

	query := `
	SELECT  "username", "email", "displayName", 
	       "roleId", "active", "photo", "empId", "firstName", "lastName", "photo", "mobileNo", "password"
	FROM public.um_users
	WHERE "orgId" = $1 AND "username" = $2
	LIMIT 1;
	`

	var u model.User
	err = conn.QueryRow(ctx, query, orgId, username).Scan(
		&u.Username,
		&u.Email,
		&u.DisplayName,
		&u.RoleID,
		&u.Active,
		&u.Photo,
		&u.EmpID,
		&u.FirstName,
		&u.LastName,
		&u.Photo,
		&u.MobileNo,
		&u.Password,
	)

	if err == pgx.ErrNoRows {
		return nil, nil // not found
	}
	if err != nil {
		return nil, fmt.Errorf("query user failed: %w", err)
	}

	//Set Cache
	userJSON, err := json.Marshal(u)
	if err != nil {
		log.Println("Error marshaling user:", err)
	} else {
		err = UsernameSet(username, string(userJSON))
		if err != nil {
			log.Println("Redis SET error:", err)
		}
	}

	return &u, nil
}

func GetAreaByNamespace(ctx context.Context, conn *pgx.Conn, orgId, namespace string) (*model.AreaDistrict, error) {
	// Example: "bma.n3-laksi-district" → "n3-laksi-district"
	// parts := strings.Split(namespace, ".")
	// ns := parts[len(parts)-1]

	query := `
		SELECT "countryId", "provId", "distId"
		FROM public.area_districts
		WHERE "orgId" = $1 AND "nameSpace" = $2
		LIMIT 1;
	`

	var a model.AreaDistrict
	err := conn.QueryRow(ctx, query, orgId, namespace).Scan(
		&a.CountryID,
		&a.ProvID,
		&a.DistID,
	)

	if err == pgx.ErrNoRows {
		return nil, nil // not found
	}
	if err != nil {
		return nil, fmt.Errorf("query area failed: %w", err)
	}

	return &a, nil
}

func GetAreaById(ctx context.Context, conn *pgx.Conn, orgId, Id string) (*model.AreaDistrict, error) {
	// Example: "bma.n3-laksi-district" → "n3-laksi-district"

	query := `
		SELECT "countryId", "provId", "distId", "nameSpace"
		FROM public.area_districts
		WHERE "orgId" = $1 AND "distId" = $2
		LIMIT 1;
	`

	var a model.AreaDistrict
	err := conn.QueryRow(ctx, query, orgId, Id).Scan(
		&a.CountryID,
		&a.ProvID,
		&a.DistID,
		&a.NameSpace,
	)

	if err == pgx.ErrNoRows {
		return nil, nil // not found
	}
	if err != nil {
		return nil, fmt.Errorf("query area failed: %w", err)
	}

	return &a, nil
}

func GetCaseSubTypeByCode(ctx context.Context, conn *pgx.Conn, orgId string, sTypeCode string) (*model.CaseSubType, error) {
	query := `
	SELECT 
	    "id", "typeId", "sTypeId", "sTypeCode", "orgId",
	    "en", "th", "wfId", "caseSla", "priority",
	    "userSkillList", "unitPropLists", "active",
	    "createdAt", "updatedAt", "createdBy", "updatedBy",
		"mDeviceType", "mWorkOrderType", "mDeviceTypeName"
	FROM public.case_sub_types
	WHERE "orgId" = $1
	  AND "sTypeId" = $2
	  AND "active" = TRUE
	LIMIT 1;
	`

	var subType model.CaseSubType
	err := conn.QueryRow(ctx, query, orgId, sTypeCode).Scan(
		&subType.Id,
		&subType.TypeID,
		&subType.STypeID,
		&subType.STypeCode,
		&subType.OrgID,
		&subType.EN,
		&subType.TH,
		&subType.WFID,
		&subType.CaseSLA,
		&subType.Priority,
		&subType.UserSkillList,
		&subType.UnitPropLists,
		&subType.Active,
		&subType.CreatedAt,
		&subType.UpdatedAt,
		&subType.CreatedBy,
		&subType.UpdatedBy,
		&subType.MDeviceType,
		&subType.MWorkOrderType,
		&subType.MDeviceTypeName,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query CaseSubType failed: %w", err)
	}

	return &subType, nil
}

func GetCaseStatusList(ctx *gin.Context, conn *pgx.Conn, orgID string) ([]model.CaseStatus, error) {
	query := `
		SELECT  "statusId", th, en, color, active 
		FROM public.case_status 
	`
	log.Print("===GetCaseStatusList=")
	log.Print(query)
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []model.CaseStatus
	for rows.Next() {
		var s model.CaseStatus
		if err := rows.Scan(
			&s.StatusID, &s.Th, &s.En, &s.Color, &s.Active,
		); err != nil {
			return nil, err
		}
		statuses = append(statuses, s)
	}

	return statuses, nil
}

// ====== GroupType Cache & Loader ======
func GroupTypeGetOrLoad(conn *pgx.Conn) ([]model.CaseGroupType, error) {
	cacheData, err := GroupTypeGet()
	if err == nil && cacheData != "" {
		// ✅ ถ้ามี cache
		var cached []model.CaseGroupType
		if jsonErr := json.Unmarshal([]byte(cacheData), &cached); jsonErr == nil {
			log.Println("✅ Loaded GroupType from Redis cache")
			return cached, nil
		}
	}

	// ❌ ไม่มี cache → query จาก DB
	rows, err := conn.Query(context.Background(), `
		SELECT id, "orgId", "groupTypeId", en, th, "groupTypeLists", "prefix"
		FROM case_type_groups`)
	if err != nil {
		return nil, fmt.Errorf("query group types failed: %v", err)
	}
	defer rows.Close()

	var result []model.CaseGroupType
	for rows.Next() {
		var g model.CaseGroupType
		var groupLists string
		if err := rows.Scan(&g.ID, &g.OrgId, &g.GroupTypeId, &g.En, &g.Th, &groupLists, &g.Prefix); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(groupLists), &g.GroupTypeLists)
		result = append(result, g)
	}

	// ✅ save to cache
	jsonData, _ := json.Marshal(result)
	_ = GroupTypeSet(string(jsonData))
	log.Println("💾 GroupType cached in Redis")

	return result, nil
}

func GetUserSkills(ctx context.Context, conn *pgx.Conn, orgID string) ([]model.GetSkills, error) {
	query := `
		SELECT "skillId", "en", "th"
		FROM public.um_skills
		WHERE "orgId" = $1 AND "active" = true
		ORDER BY id ASC;
	`

	rows, err := conn.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var skills []model.GetSkills
	for rows.Next() {
		var s model.GetSkills
		if err := rows.Scan(&s.SkillID, &s.En, &s.Th); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		skills = append(skills, s)
	}

	// If no records, return empty slice instead of nil
	if skills == nil {
		return []model.GetSkills{}, nil
	}

	return skills, nil
}

func GetUnitProp(ctx context.Context, conn *pgx.Conn, orgID string) ([]model.GetUnisProp, error) {
	query := `
		SELECT "propId", "en", "th"
		FROM public.mdm_properties
		WHERE "orgId" = $1 AND "active" = true
		ORDER BY id ASC;
	`

	rows, err := conn.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var props []model.GetUnisProp
	for rows.Next() {
		var s model.GetUnisProp
		if err := rows.Scan(&s.PropId, &s.En, &s.Th); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		props = append(props, s)
	}

	// If no records, return empty slice instead of nil
	if props == nil {
		return []model.GetUnisProp{}, nil
	}

	return props, nil
}

// ---------- Query User Profile ----------

func GetUserProfileFromDB(ctx context.Context, dbConn *pgx.Conn, orgId, username string) (*model.UserConnectionInfo, error) {
	log.Printf("Database: Querying for user '%s' in organization '%s'", username, orgId)

	var userProfile model.UserConnectionInfo
	var roleID string
	var distIdListsJSON []byte
	var GrpID []string

	// 1) ลองอ่านจาก user_connections ก่อน (เก็บ grpId เป็น array อยู่แล้ว)
	connectionQuery := `
        SELECT "empId", "username", "orgId", "deptId", "commId", "stnId", "roleId", "grpId", "distIdLists", COALESCE("ip", '') as ip
        FROM user_connections
        WHERE "orgId" = $1 AND "username" = $2
        LIMIT 1;
    `
	err := dbConn.QueryRow(ctx, connectionQuery, orgId, username).Scan(
		&userProfile.ID, &userProfile.Username, &userProfile.OrgID,
		&userProfile.DeptID, &userProfile.CommID, &userProfile.StnID,
		&roleID, &GrpID, &userProfile.DistIdLists, &userProfile.Ip, // scan array -> []string
	)
	if err == nil {
		userProfile.RoleID = roleID
		userProfile.GrpID = GrpID
		log.Printf("Database: Found existing connection for '%s'", username)
		return &userProfile, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		log.Printf("ERROR: Failed to query user connections for '%s': %v", username, err)
		return nil, err
	}

	// 2) ไม่เจอใน user_connections -> ไปอ่านจาก um_users + um_user_with_groups
	//    รวมหลายแถวของ grpId ให้เป็น array ด้วย array_agg(DISTINCT ...)
	query := `
        SELECT 
          COALESCE(u."empId"::text, '')  AS "empId",
          u."username",
          COALESCE(u."orgId"::text, '')  AS "orgId",
          COALESCE(u."deptId"::text, '') AS "deptId",
          COALESCE(u."commId"::text, '') AS "commId",
          COALESCE(u."stnId"::text, '')  AS "stnId",
          COALESCE(u."roleId"::text, '') AS "roleId",
          COALESCE(array_agg(DISTINCT ug."grpId"::text) FILTER (WHERE ug."grpId" IS NOT NULL), '{}') AS "grpIds",
          COALESCE(uar."distIdLists", '[]'::jsonb) AS "distIdLists"
        FROM um_users u
        LEFT JOIN um_user_with_groups ug 
               ON u."username" = ug."username"
        LEFT JOIN um_user_with_area_response uar 
               ON u."username" = uar."username"
        WHERE u."orgId"::text = $1 
          AND u."username" = $2 
          AND u."active" = true
        GROUP BY u."empId", u."username", u."orgId", u."deptId", u."commId", u."stnId", u."roleId", uar."distIdLists"
        LIMIT 1;
    `
	err = dbConn.QueryRow(ctx, query, orgId, username).Scan(
		&userProfile.ID, &userProfile.Username, &userProfile.OrgID,
		&userProfile.DeptID, &userProfile.CommID, &userProfile.StnID,
		&roleID, &GrpID, &distIdListsJSON,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found or is not active")
		}
		log.Printf("ERROR: Failed to query user profile for '%s': %v", username, err)
		return nil, err
	}

	userProfile.RoleID = roleID
	userProfile.GrpID = GrpID

	// distIdLists จากตาราง uar เป็น jsonb -> unmarshal เป็น []string
	if len(distIdListsJSON) > 0 {
		if err := json.Unmarshal(distIdListsJSON, &userProfile.DistIdLists); err != nil {
			log.Printf("WARNING: Failed to parse distIdLists for user '%s': %v", username, err)
			userProfile.DistIdLists = []string{}
		}
	}

	return &userProfile, nil
}

func GetAreaByUsernameOrLoad(ctx context.Context, dbConn *pgx.Conn, orgId string, username string) (*model.Um_User_Login, error) {

	// ==== LOAD FROM CACHE ====
	cacheData, err := UserPermissionGet(username)
	if err == nil && cacheData != "" {
		var cached model.Um_User_Login
		if jsonErr := json.Unmarshal([]byte(cacheData), &cached); jsonErr == nil {
			log.Println("✅ Loaded UserPermission from Redis cache")
			return &cached, nil
		}
	}

	log.Print("=====GetAreaByUsernameOrLoad=====")

	sql := `
        SELECT
            u.id,
            u."orgId",
            u.username,
            u."displayName",
            u."firstName",
            u."lastName",
            u.email,
            u."mobileNo",
            u.photo, 
            ar."distIdLists"
        FROM um_users u
        LEFT JOIN um_user_with_area_response ar
            ON ar.username = u.username
            AND ar."orgId" = u."orgId"
        WHERE u.username = $1
          AND u."orgId" = $2
        LIMIT 1;
    `

	row := dbConn.QueryRow(ctx, sql, username, orgId)

	var ua model.Um_User_Login
	var distIDLists []byte

	err = row.Scan(
		&ua.ID,
		&ua.OrgID,
		&ua.Username,
		&ua.DisplayName,
		&ua.FirstName,
		&ua.LastName,
		&ua.Email,
		&ua.MobileNo,
		&ua.Photo,
		&distIDLists,
	)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	if distIDLists != nil {
		json.Unmarshal(distIDLists, &ua.DistIdLists)
	}

	// ==== SAVE TO CACHE ====
	jsonData, _ := json.Marshal(ua)
	UserPermissionSet(username, string(jsonData))

	log.Println("Saved to cache:", string(jsonData))

	return &ua, nil
}

func GetCountryProvinceDistrictsOrLoad(ctx context.Context, conn *pgx.Conn, orgId string) ([]model.AreaDistrictWithDetails, error) {

	// ==== LOAD FROM CACHE ====
	cacheData, err := OwnerDistGet(orgId)
	if err == nil && cacheData != "" {
		var cached []model.AreaDistrictWithDetails

		if jsonErr := json.Unmarshal([]byte(cacheData), &cached); jsonErr == nil {
			log.Println("✅ Loaded UserPermission from Redis cache")
			return cached, nil
		}
	}

	query := `
	SELECT 
	    t1.id, t1."orgId", t1."countryId", t1."provId", t1."distId",
	    t1.en AS dist_en, t1.th AS dist_th, t1.active AS dist_active,
	    t2.en AS prov_en, t2.th AS prov_th, t2.active AS prov_active,
	    t3.en AS country_en, t3.th AS country_th, t3.active AS country_active,
		t1."nameSpace"
	FROM area_districts t1
	LEFT JOIN area_provinces t2 
	    ON t1."provId" = t2."provId" 
	    AND t1."countryId" = t2."countryId"
	    AND t2."orgId" = t1."orgId"
	LEFT JOIN area_countries t3 
	    ON t1."countryId" = t3."countryId"
	    AND t3."orgId" = t1."orgId"
	WHERE t1."orgId" = $1
	ORDER BY t1."countryId";
	`

	rows, err := conn.Query(ctx, query, orgId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// log.Println("err --> ", err)
	// log.Println("data --> ", rows)
	var list []model.AreaDistrictWithDetails

	for rows.Next() {
		var a model.AreaDistrictWithDetails
		err := rows.Scan(
			&a.ID, &a.OrgID, &a.CountryID, &a.ProvID, &a.DistID,
			&a.DistrictEn, &a.DistrictTh, &a.DistrictActive,
			&a.ProvinceEn, &a.ProvinceTh, &a.ProvinceActive,
			&a.CountryEn, &a.CountryTh, &a.CountryActive,
			&a.NameSpace,
		)
		if err != nil {
			log.Println("Scan error:", err)
			return nil, err
		}
		//log.Printf("Row: %+v\n", a) // <-- debug
		list = append(list, a)
	}
	//log.Println("Total rows:", len(list))
	// ==== SAVE TO CACHE ====
	jsonData, _ := json.Marshal(list)
	OwnerDistSet(orgId, string(jsonData))

	//log.Println("Saved to cache:", string(jsonData))

	return list, nil
}

func GetDepartmentCommandStationOrLoad(ctx context.Context, conn *pgx.Conn, orgId string) ([]model.StationWithCommandDept, error) {
	// ==== LOAD FROM CACHE ====
	cacheData, err := OwnerStationGet(orgId)
	if err == nil && cacheData != "" {
		var cached []model.StationWithCommandDept

		if jsonErr := json.Unmarshal([]byte(cacheData), &cached); jsonErr == nil {
			log.Println("✅ Loaded UserPermission from Redis cache")
			return cached, nil
		}
	}

	query := `SELECT
	    s."id",
	    s."orgId",
	    s."deptId",
	    s."commId",
	    s."stnId",
	    s.en AS stn_en,
	    s.th AS stn_th,
	    s.active AS stn_active,
	    
	    c.en AS comm_en,
	    c.th AS comm_th,
	    c.active AS comm_active,
	    
	    d.en AS dept_en,
	    d.th AS dept_th,
	    d.active AS dept_active
	FROM public.sec_departments d
	JOIN public.sec_commands c
	    ON c."deptId" = d."deptId"
	JOIN public.sec_stations s
	    ON s."commId" = c."commId"
	WHERE d."orgId"=$1
	ORDER BY d."deptId", c."commId", s."stnId";`

	rows, err := conn.Query(ctx, query, orgId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stationList []model.StationWithCommandDept
	for rows.Next() {
		var station model.StationWithCommandDept
		err := rows.Scan(
			&station.ID, &station.OrgId, &station.DeptId, &station.CommId, &station.StnId,
			&station.StationEn, &station.StationTh, &station.StationActive,
			&station.CommandEn, &station.CommandTh, &station.CommandActive,
			&station.DeptEn, &station.DeptTh, &station.DeptActive,
		)
		if err != nil {
			return nil, err
		}
		stationList = append(stationList, station)
	}

	// ==== SAVE TO CACHE ====
	jsonData, _ := json.Marshal(stationList)
	OwnerStationSet(orgId, string(jsonData))

	return stationList, nil
}

func GetUserSkillsOrLoad(ctx context.Context, conn *pgx.Conn, orgId string) ([]model.Skill, error) {
	// ==== LOAD FROM CACHE ====
	cacheData, err := OwnerUserSkillsGet(orgId)
	if err == nil && cacheData != "" {
		var cached []model.Skill
		if jsonErr := json.Unmarshal([]byte(cacheData), &cached); jsonErr == nil {
			log.Println("✅ Loaded UserSkills from Redis cache")
			return cached, nil
		}
	}

	// ==== QUERY FROM DATABASE ====
	query := `
		SELECT 
		    id,
		    "orgId",
		    "skillId",
		    en,
		    th,
		    active,
		    "createdAt",
		    "updatedAt",
		    "createdBy",
		    "updatedBy"
		FROM public.um_skills
		WHERE "orgId" = $1
		ORDER BY en ASC;
	`

	rows, err := conn.Query(ctx, query, orgId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []model.Skill
	for rows.Next() {
		var skill model.Skill
		err := rows.Scan(
			&skill.ID,
			&skill.OrgID,
			&skill.SkillID,
			&skill.En,
			&skill.Th,
			&skill.Active,
			&skill.CreatedAt,
			&skill.UpdatedAt,
			&skill.CreatedBy,
			&skill.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		skills = append(skills, skill)
	}

	// ==== SAVE TO CACHE ====
	jsonData, _ := json.Marshal(skills)
	OwnerUserSkillsSet(orgId, string(jsonData))

	return skills, nil
}
