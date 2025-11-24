package model

import "time"

type Form struct {
	FormId        *string               `json:"formId"`
	FormName      *string               `json:"formName"`
	FormColSpan   int                   `json:"formColSpan"`
	FormFieldJson []IndividualFormField `json:"formFieldJson"`
}

type FormRule struct {
	MaxLength        *int     `json:"maxLength,omitempty"`
	MinLength        *int     `json:"minLength,omitempty"`
	Contain          *string  `json:"contain,omitempty"`
	MaxNumber        *float64 `json:"maxnumber,omitempty"`
	MinNumber        *float64 `json:"minnumber,omitempty"`
	ValidEmailFormat *bool    `json:"validEmailFormat,omitempty"`
	MaxSelections    *int     `json:"maxSelections,omitempty"`
	MinSelections    *int     `json:"minSelections,omitempty"`
	MaxFileSize      *int     `json:"maxFileSize,omitempty"`
	AllowedFileTypes []string `json:"allowedFileTypes,omitempty"`
	AllowedCountries []string `json:"allowedCountries,omitempty"`
	HasUppercase     *bool    `json:"hasUppercase,omitempty"`
	HasLowercase     *bool    `json:"hasLowercase,omitempty"`
	HasNumber        *bool    `json:"hasNumber,omitempty"`
	HasSpecialChar   *bool    `json:"hasSpecialChar,omitempty"`
	NoWhitespace     *bool    `json:"noWhitespace,omitempty"`
	MinDate          *string  `json:"minDate,omitempty"`
	MaxDate          *string  `json:"maxDate,omitempty"`
	MinLocalDate     *string  `json:"minLocalDate,omitempty"`
	MaxLocalDate     *string  `json:"maxLocalDate,omitempty"`
	FutureDateOnly   *bool    `json:"futureDateOnly,omitempty"`
	PastDateOnly     *bool    `json:"pastDateOnly,omitempty"`
	MinFiles         *int     `json:"minFiles,omitempty"`
	MaxFiles         *int     `json:"maxFiles,omitempty"`
}

type IndividualFormField struct {
	ID                  string        `json:"id"`
	Label               string        `json:"label"`
	ShowLabel           *bool         `json:"showLabel,omitempty"`
	Type                string        `json:"type"`
	Value               interface{}   `json:"value"`
	EnableSearch        *bool         `json:"enableSearch,omitempty"`
	Options             []interface{} `json:"options,omitempty"`
	Placeholder         *string       `json:"placeholder,omitempty"`
	Required            bool          `json:"required"`
	ColSpan             *int          `json:"colSpan,omitempty"`
	IsChild             *bool         `json:"isChild,omitempty"`
	GroupColSpan        *int          `json:"GroupColSpan,omitempty"`
	DynamicFieldColSpan *int          `json:"DynamicFieldColSpan,omitempty"`
	FormRule            *FormRule     `json:"formRule,omitempty"`
	UID                 *string       `json:"uid,omitempty"`
}

type FormFieldOption struct {
	Value string                            `json:"value"`
	Form  []IndividualFormFieldWithChildren `json:"form,omitempty"`
}

type IndividualFormFieldWithChildren struct {
	ID                  string        `json:"id"`
	Label               string        `json:"label"`
	ShowLabel           *bool         `json:"showLabel,omitempty"`
	Type                string        `json:"type"`
	Value               interface{}   `json:"value"`
	EnableSearch        *bool         `json:"enableSearch,omitempty"`
	Options             []interface{} `json:"options,omitempty"`
	Placeholder         *string       `json:"placeholder,omitempty"`
	Required            bool          `json:"required"`
	ColSpan             *int          `json:"colSpan,omitempty"`
	IsChild             *bool         `json:"isChild,omitempty"`
	GroupColSpan        *int          `json:"GroupColSpan,omitempty"`
	DynamicFieldColSpan *int          `json:"DynamicFieldColSpan,omitempty"`
	FormRule            *FormRule     `json:"formRule,omitempty"`
}

type VersionInfo struct {
	Version string `json:"version"`
	Publish bool   `json:"publish"`
}

type FormsManager struct {
	Form
	Active           bool          `json:"active"`
	Publish          bool          `json:"publish"`
	Versions         string        `json:"versions"`
	Locks            bool          `json:"locks"`
	CreatedAt        time.Time     `json:"createdAt"`
	UpdatedAt        time.Time     `json:"updatedAt"`
	CreatedBy        string        `json:"createdBy"`
	UpdatedBy        string        `json:"updatedBy"`
	VersionsInfoList []VersionInfo `json:"versionsInfoList"`
}

type FormsManagerShotModel struct {
	FormId           *string       `json:"formId"`
	FormName         *string       `json:"formName"`
	Publish          bool          `json:"publish"`
	Versions         string        `json:"versions"`
	CreatedAt        time.Time     `json:"createdAt"`
	CreatedBy        string        `json:"createdBy"`
	VersionsInfoList []VersionInfo `json:"versionsInfoList"`
}

type FormsModel struct {
	FormId   *string `json:"formId"`
	FormName *string `json:"formName"`
	Publish  bool    `json:"publish"`
	Versions string  `json:"versions"`
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
	FormName      *string               `json:"formName"`
	FormColSpan   int                   `json:"formColSpan"`
	Active        bool                  `json:"active"`
	Publish       bool                  `json:"publish"`
	Locks         bool                  `json:"locks"`
	FormFieldJson []IndividualFormField `json:"formFieldJson"`
}

type FormActive struct {
	Active bool   `json:"active"`
	FormID string `json:"formId"`
}

type FormPublish struct {
	FormID  string `json:"formId"`
	Publish bool   `json:"publish"`
}

type FormChangeVersion struct {
	FormID  string `json:"formId"`
	Version string `json:"version"`
}

type FormLock struct {
	FormID string `json:"formId"`
	Locks  bool   `json:"locks"`
}
type FormUpdate struct {
	FormName      *string               `json:"formName"`
	FormColSpan   int                   `json:"formColSpan"`
	Active        bool                  `json:"active"`
	Publish       bool                  `json:"publish"`
	Locks         bool                  `json:"locks"`
	FormFieldJson []IndividualFormField `json:"formFieldJson"`
}

type WorkFlow struct {
	Nodes       []map[string]interface{} `json:"nodes"`
	Connections []map[string]interface{} `json:"connections"`
	MetaData    interface{}              `json:"metadata"`
}

type WorkFlowMetadata struct {
	Id         int       `json:"id"`
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
	Label  string `json:"label"`
}

type FormByCasesubtype struct {
	CaseSubType *string `json:"caseSubType"`
}

type WorkflowModel struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"orgId"`
	WfID      string    `json:"wfId"`
	Title     string    `json:"title"`
	Desc      string    `json:"desc"`
	Active    bool      `json:"active"`
	Publish   bool      `json:"publish"`
	Locks     bool      `json:"locks"`
	Versions  string    `json:"versions"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	CreatedBy string    `json:"createdBy"`
	UpdatedBy string    `json:"updatedBy"`
}

type WfNode struct {
	ID        string                 `json:"id" db:"id"`
	OrgID     string                 `json:"orgId" db:"orgId"`
	WfID      string                 `json:"wfId" db:"wfId"`
	NodeID    string                 `json:"nodeId" db:"nodeId"`
	Versions  string                 `json:"versions" db:"versions"`
	Type      string                 `json:"type" db:"type"`
	Section   string                 `json:"section" db:"section"`
	Data      map[string]interface{} `json:"data" db:"data"` // Or json.RawMessage if it's JSON
	Pic       *string                `json:"pic" db:"pic"`   // Change to []string if stored as array
	Group     *string                `json:"group" db:"group"`
	FormID    *string                `json:"formId" db:"formId"`
	CreatedAt time.Time              `json:"createdAt" db:"createdAt"`
	UpdatedAt time.Time              `json:"updatedAt" db:"updatedAt"`
	CreatedBy string                 `json:"createdBy" db:"createdBy"`
	UpdatedBy string                 `json:"updatedBy" db:"updatedBy"`
}

type FormAnswer struct {
	ID        int64                  `json:"id"`
	OrgId     string                 `json:"orgId"`
	CaseId    string                 `json:"caseId"`
	FormId    string                 `json:"formId"`
	Versions  string                 `json:"versions"`
	EleNumber int                    `json:"eleNumber"`
	EleData   map[string]interface{} `json:"eleData"`
	CreatedAt time.Time              `json:"createdAt"`
	UpdatedAt time.Time              `json:"updatedAt"`
	CreatedBy string                 `json:"createdBy"`
	UpdatedBy string                 `json:"updatedBy"`
}

type FormAnswerRequest struct {
	NextNodeId    string                `json:"nextNodeId"`
	Versions      string                `json:"versions"`
	WfId          string                `json:"wfId"`
	FormId        string                `json:"formId"`
	FormName      string                `json:"formName"`
	FormColSpan   int                   `json:"formColSpan"`
	FormFieldJson []IndividualFormField `json:"formFieldJson"`
	UID           *string               `json:"uid,omitempty"`
}

type NodeData struct {
	ID   string `json:"id"`
	Data struct {
		Label  string `json:"label"`
		Config struct {
			Pic    []string `json:"pic"`
			SLA    string   `json:"sla"` // SLA in minutes as string
			Group  []string `json:"group"`
			Action string   `json:"action"`
			FormID string   `json:"formId"`
		} `json:"config"`
	} `json:"data"`
	Type     string `json:"type"`
	Position struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"position"`
}
