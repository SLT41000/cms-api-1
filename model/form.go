package model

import "time"

type Form struct {
	FormId        *string                  `json:"formId"`
	FormName      *string                  `json:"formName"`
	FormColSpan   int                      `json:"formColSpan"`
	FormFieldJson []map[string]interface{} `json:"formFieldJson"`
}
type FormByCasesubtypeOpt struct {
	NodeId        *string                  `json:"nodeId"`
	FormId        *string                  `json:"formId"`
	FormName      *string                  `json:"formName"`
	FormColSpan   int                      `json:"formColSpan"`
	FormFieldJson []map[string]interface{} `json:"formFieldJson"`
}

type FormsManager struct {
	Form
	Active    bool      `json:"active"`
	Publish   bool      `json:"publish"`
	Versions  string    `json:"versions"`
	Locks     bool      `json:"locks"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	CreatedBy string    `json:"createdBy"`
	UpdatedBy string    `json:"updatedBy"`
}

type FormBuilder struct {
	OrgID       string    `json:"orgId"`
	FormID      string    `json:"formId"`
	FormName    string    `json:"formName"`
	FormColSpan int       `json:"formColSpan"`
	Active      bool      `json:"active"`
	Publish     bool      `json:"publish"`
	Versions    string    `json:"versions"`
	Locks       bool      `json:"locks"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	CreatedBy   string    `json:"createdBy"`
	UpdatedBy   string    `json:"updatedBy"`
}

type FormInsert struct {
	FormName      *string                  `json:"formName"`
	FormColSpan   int                      `json:"formColSpan"`
	Active        bool                     `json:"active"`
	Publish       bool                     `json:"publish"`
	Locks         bool                     `json:"locks"`
	FormFieldJson []map[string]interface{} `json:"formFieldJson"`
}

type FormActive struct {
	Active bool   `json:"active"`
	FormID string `json:"formId"`
}

type FormPublish struct {
	FormID  string `json:"formId"`
	Publish bool   `json:"publish"`
}

type FormLock struct {
	FormID string `json:"formId"`
	Locks  bool   `json:"locks"`
}
type FormUpdate struct {
	FormName      *string                  `json:"formName"`
	FormColSpan   int                      `json:"formColSpan"`
	Active        bool                     `json:"active"`
	Publish       bool                     `json:"publish"`
	Locks         bool                     `json:"locks"`
	FormFieldJson []map[string]interface{} `json:"formFieldJson"`
}

type WorkFlow struct {
	Nodes       []map[string]interface{} `json:"nodes"`
	Connections []map[string]interface{} `json:"connections"`
	MetaData    interface{}              `json:"metadata"`
}

type WorkFlowMetadata struct {
	Title      *string   `json:"title"`
	CaseTypeId *string   `json:"caseTypeId,omitempty"`
	Desc       *string   `json:"description"`
	Status     *string   `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type WorkFlowInsert struct {
	Nodes       []WorkFlowNode       `json:"nodes"`
	Connections []WorkFlowConnection `json:"connections"`
	MetaData    WorkFlowMetadata     `json:"metadata"`
}

type WorkFlowNode struct {
	Id       string                 `json:"id"`
	Type     string                 `json:"type"`
	Position map[string]interface{} `json:"position"`
	Data     *NodeConfig            `json:"data"`
}

type NodeConfig struct {
	Label  string                  `json:"label"`
	Config *map[string]interface{} `json:"config"`
}

type WorkFlowConnection struct {
	Id     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
}

type FormByCasesubtype struct {
	CaseSubType *string `json:"caseSubType"`
}
