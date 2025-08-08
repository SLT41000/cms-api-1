package model

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
