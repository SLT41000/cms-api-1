package model

import (
	"time"
)

type CustomCaseCurrentStage struct {
	CaseID   string  `json:"caseId" db:"caseId"`
	WfID     *string `json:"wfId" db:"wfId"`
	NodeID   string  `json:"nodeId" db:"nodeId"`
	StatusID string  `json:"statusId" db:"statusId"`
}

type CaseCurrentStage struct {
	Id        int       `json:"id"`
	OrgID     string    `json:"orgId" db:"orgId"`
	CaseID    string    `json:"caseId" db:"caseId"`
	WfID      string    `json:"wfId" db:"wfId"`
	NodeID    string    `json:"nodeId" db:"nodeId"`
	Versions  string    `json:"versions" db:"versions"`
	Type      string    `json:"type" db:"type"`
	Section   string    `json:"section" db:"section"`
	Data      string    `json:"data" db:"data"` // consider json.RawMessage if it's JSON
	Pic       string    `json:"pic" db:"pic"`   // or []string if it's a list
	Group     string    `json:"group" db:"group"`
	FormID    string    `json:"formId" db:"formId"`
	CreatedAt time.Time `json:"createdAt" db:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" db:"updatedAt"`
	CreatedBy string    `json:"createdBy" db:"createdBy"`
	UpdatedBy string    `json:"updatedBy" db:"updatedBy"`
}

type CaseCurrentStageInsert struct {
	CaseID   string `json:"caseId" db:"caseId"`
	WfID     string `json:"wfId" db:"wfId"`
	NodeID   string `json:"nodeId" db:"nodeId"`
	Versions string `json:"versions" db:"versions"`
	Type     string `json:"type" db:"type"`
	Section  string `json:"section" db:"section"`
	Data     string `json:"data" db:"data"` // consider json.RawMessage if it's JSON
	Pic      string `json:"pic" db:"pic"`   // or []string if it's a list
	Group    string `json:"group" db:"group"`
	FormID   string `json:"formId" db:"formId"`
}
