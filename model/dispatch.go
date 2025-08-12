package model

import "time"

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
	CaseId   string      `json:"caseId"`
	NodeId   string      `json:"nodeId"`
	Versions string      `json:"versions"`
	Type     string      `json:"type"`
	Section  string      `json:"section"`
	Data     interface{} `json:"data"` // jsonb
	Pic      *string     `json:"pic"`
	Group    *string     `json:"group"`
	FormId   *string     `json:"formId"`
}

type UnitUser struct {
	OrgID             string    `json:"orgId"`
	UnitID            string    `json:"unitId"`
	UnitName          string    `json:"unitName"`
	UnitSourceID      string    `json:"unitSourceId"`
	UnitTypeID        string    `json:"unitTypeId"`
	Priority          int       `json:"priority"`
	CompID            string    `json:"compId"`
	DeptID            string    `json:"deptId"`
	CommID            string    `json:"commId"`
	StnID             string    `json:"stnId"`
	PlateNo           string    `json:"plateNo"`
	ProvinceCode      string    `json:"provinceCode"`
	Active            bool      `json:"active"`
	Username          string    `json:"username"`
	IsLogin           bool      `json:"isLogin"`
	IsFreeze          bool      `json:"isFreeze"`
	IsOutArea         bool      `json:"isOutArea"`
	LocLat            float64   `json:"locLat"`
	LocLon            float64   `json:"locLon"`
	LocAlt            float64   `json:"locAlt"`
	LocBearing        float64   `json:"locBearing"`
	LocSpeed          float64   `json:"locSpeed"`
	LocProvider       float64   `json:"locProvider"`
	LocGpsTime        time.Time `json:"locGpsTime"`
	LocSatellites     int       `json:"locSatellites"`
	LocAccuracy       float64   `json:"locAccuracy"`
	LocLastUpdateTime time.Time `json:"locLastUpdateTime"`
	BreakDuration     int       `json:"breakDuration"`
	HealthChk         string    `json:"healthChk"`
	HealthChkTime     time.Time `json:"healthChkTime"`
	SttID             string    `json:"sttId"`
	CreatedBy         string    `json:"createdBy"`
	UpdatedBy         string    `json:"updatedBy"`
}
