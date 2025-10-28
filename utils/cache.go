package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"mainPackage/model"
	"strings"

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
	query := `
	SELECT  "username", "email", "displayName", 
	       "roleId", "active", "photo", "empId", "firstName", "lastName", "photo", "mobileNo"
	FROM public.um_users
	WHERE "orgId" = $1 AND "username" = $2
	LIMIT 1;
	`

	var u model.User
	err := conn.QueryRow(ctx, query, orgId, username).Scan(
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
	)

	if err == pgx.ErrNoRows {
		return nil, nil // not found
	}
	if err != nil {
		return nil, fmt.Errorf("query user failed: %w", err)
	}

	return &u, nil
}

func GetAreaByNamespace(ctx context.Context, conn *pgx.Conn, orgId, namespace string) (*model.AreaDistrict, error) {
	// Example: "bma.n3-laksi-district" → "n3-laksi-district"
	parts := strings.Split(namespace, ".")
	ns := parts[len(parts)-1]

	query := `
		SELECT "countryId", "provId", "distId"
		FROM public.area_districts
		WHERE "orgId" = $1 AND "nameSpace" = $2
		LIMIT 1;
	`

	var a model.AreaDistrict
	err := conn.QueryRow(ctx, query, orgId, ns).Scan(
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
		"mDeviceType", "mWorkOrderType"
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
