package handler

import (
	"context"
	"fmt"
	"mainPackage/model"

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
	       "roleId", "active", "photo"
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
	)

	if err == pgx.ErrNoRows {
		return nil, nil // not found
	}
	if err != nil {
		return nil, fmt.Errorf("query user failed: %w", err)
	}

	return &u, nil
}
