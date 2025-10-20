package model

import "time"

type CaseStageInfo struct {
	CaseId       string        `json:"caseId"`
	StatusId     string        `json:"statusId"`
	Data         string        `json:"data"`
	CreatedDate  *time.Time    `json:"createdDate"`
	Versions     string        `json:"versions"`
	OverSlaCount string        `json:"overSlaCount"`
	WfId         string        `json:"wfId"`
	NodeId       string        `json:"nodeId"`
	NextNode     *WorkflowNode `json:"nextNode,omitempty"`
}

type WorkflowNodeData struct {
	Data struct {
		Config struct {
			Action string   `json:"action"`
			FormID string   `json:"formId"`
			Group  []string `json:"group"`
			Pic    []string `json:"pic"`
			SLA    string   `json:"sla"` // <- this is what you want
		} `json:"config"`
		Label string `json:"label"`
	} `json:"data"`
	ID       string `json:"id"`
	Position struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"position"`
	Type string `json:"type"`
}
