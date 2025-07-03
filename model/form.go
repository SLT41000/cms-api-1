package model

type FormGetOptModel struct {
	FormId        *string                  `json:"formId"`
	FormName      *string                  `json:"formName"`
	FormColSpan   int                      `json:"formColSpan"`
	FormFieldJson []map[string]interface{} `json:"formFieldJson"`
}
