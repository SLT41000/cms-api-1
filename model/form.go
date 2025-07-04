package model

import "time"

type FormGetOptModel struct {
	FormId        *string                  `json:"formId"`
	FormName      *string                  `json:"formName"`
	FormColSpan   int                      `json:"formColSpan"`
	FormFieldJson []map[string]interface{} `json:"formFieldJson"`
}

type WorkFlowGetOptModel struct {
	Nodes       []map[string]interface{} `json:"nodes"`
	Connections []map[string]interface{} `json:"connections"`
	MetaData    interface{}              `json:"metadata"`
}

type WorkFlowMetadata struct {
	Title     *string   `json:"title"`
	Desc      *string   `json:"description"`
	Status    *string   `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
