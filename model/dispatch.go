package model

import (
	"time"
)

type SOPRequest struct {
	CaseID string `json:"caseId" binding:"required"` // จาก JSON หรือ Query param
	WfID   string `json:"wfId" binding:"required"`
}

type WorkflowNode struct {
	NodeId  string      `json:"nodeId"`
	Type    string      `json:"type"`
	Section string      `json:"section"`
	Data    interface{} `json:"data"` // jsonb
}

type CurrentStage struct {
	CaseId    string      `json:"caseId" db:"caseId"`
	WfID      string      `json:"wfId" db:"wfId"`
	NodeId    string      `json:"nodeId"`
	Versions  string      `json:"versions"`
	Type      string      `json:"type"`
	Section   string      `json:"section"`
	Data      interface{} `json:"data"` // jsonb
	Pic       *string     `json:"pic"`
	Group     *string     `json:"group"`
	FormId    *string     `json:"formId"`
	StageType string      `json:"stageType"`
	UnitID    string      `json:"unitId"`
	UserOwner *string     `json:"username"`
}

type UnitUser struct {
	OrgID             string      `json:"orgId"`
	UnitID            string      `json:"unitId"`
	UnitName          string      `json:"unitName"`
	UnitSourceID      string      `json:"unitSourceId"`
	UnitTypeID        string      `json:"unitTypeId"`
	Priority          int         `json:"priority"`
	CompID            string      `json:"compId"`
	DeptID            string      `json:"deptId"`
	CommID            string      `json:"commId"`
	StnID             string      `json:"stnId"`
	PlateNo           string      `json:"plateNo"`
	ProvinceCode      *string     `json:"provinceCode"`
	Active            bool        `json:"active"`
	Username          string      `json:"username"`
	IsLogin           bool        `json:"isLogin"`
	IsFreeze          bool        `json:"isFreeze"`
	IsOutArea         bool        `json:"isOutArea"`
	LocLat            *float64    `json:"locLat"`
	LocLon            *float64    `json:"locLon"`
	LocAlt            *float64    `json:"locAlt"`
	LocBearing        *float64    `json:"locBearing"`
	LocSpeed          *float64    `json:"locSpeed"`
	LocProvider       *string     `json:"locProvider"`
	LocGpsTime        *time.Time  `json:"locGpsTime"`
	LocSatellites     *int        `json:"locSatellites"`
	LocAccuracy       *float64    `json:"locAccuracy"`
	LocLastUpdateTime *time.Time  `json:"locLastUpdateTime"`
	BreakDuration     *int        `json:"breakDuration"`
	HealthChk         *string     `json:"healthChk"`
	HealthChkTime     *time.Time  `json:"healthChkTime"`
	SttID             *string     `json:"sttId"`
	CreatedBy         string      `json:"createdBy"`
	UpdatedBy         string      `json:"updatedBy"`
	UnitPropLists     *[]string   `json:"unitPropLists"`
	UserSkillList     *[]string   `json:"userSkillList"`
	SkillLists        interface{} `json:"skillLists"`
	ProplLists        interface{} `json:"proplLists"`
}

type UpdateStageRequest struct {
	CaseId    string `json:"caseId"`
	Status    string `json:"status"`
	UnitId    string `json:"unitId"`
	UnitUser  string `json:"unitUser"`
	NodeId    string `json:"nodeId"`
	ResID     string `json:"resId"`
	ResDetail string `json:"resDetail"`
}

type StageResult struct {
	NextNodeId   string `json:"nextNodeId"`
	NextNodeType string `json:"nextNodeType"`
	StageType    string `json:"stageType"`
	CaseId       string `json:"caseId"`
}

type Connection struct {
	Id     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
	Label  string `json:"label"`
}

type GetSkills struct {
	SkillID string `json:"skillId"`
	En      string `json:"en"`
	Th      string `json:"th"`
}

// Struct for result
type UnitDispatch struct {
	UnitID    string `json:"unitId"`
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type GetUnisProp struct {
	PropId string `json:"propId"`
	En     string `json:"en"`
	Th     string `json:"th"`
}

type CaseResponderCustom struct {
	OrgID     string    `json:"orgId" db:"orgId"`
	CaseID    string    `json:"caseId" db:"caseId"`
	UnitID    string    `json:"unitId" db:"unitId"`
	UserOwner string    `json:"userOwner" db:"userOwner"`
	StatusID  string    `json:"statusId" db:"statusId"`
	StatusTh  *string   `json:"statusTh" db:"statusTh"`
	StatusEn  *string   `json:"statusEn" db:"statusEn"`
	CreatedAt time.Time `json:"createdAt" db:"createdAt"`
	Duration  int64     `json:"duration"` // duration in seconds
}

type CaseHistoryEvent struct {
	OrgID     string      `json:"orgId"`
	CaseID    string      `json:"caseId"`
	Username  string      `json:"username"`
	Type      string      `json:"type"`     // 'comment' หรือ 'event'
	FullMsg   string      `json:"fullMsg"`  // optional
	JsonData  interface{} `json:"jsonData"` // optional, will be marshaled to TEXT
	CreatedBy string      `json:"createdBy"`
}

type CloseCaseRequest struct {
	CaseID    string `json:"caseId"`
	StatusID  string `json:"statusId"`
	ResID     string `json:"resId"`
	ResDetail string `json:"resDetail"`
	UpdatedBy string `json:"updatedBy"`
}

type CancelUnitRequest struct {
	CaseId    string `json:"caseId"`
	ResDetail string `json:"resDetail"`
	ResId     string `json:"resId"`
	UnitId    string `json:"unitId"`
	UnitUser  string `json:"unitUser"`
}

type CancelCaseRequest struct {
	CaseId    string `json:"caseId"`
	ResDetail string `json:"resDetail"`
	ResId     string `json:"resId"`
}
