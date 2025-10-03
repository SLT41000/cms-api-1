package model

import "time"

type CaseStageInfo struct {
	CaseId       string     `json:"caseId"`
	StatusId     string     `json:"statusId"`
	Data         string     `json:"data"`
	CreatedDate  *time.Time `json:"createdDate"`
	Versions     string     `json:"versions"`
	OverSlaCount string     `json:"overSlaCount"`
}
