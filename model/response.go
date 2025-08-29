package model

type Response struct {
	Status string      `json:"status"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data,omitempty"`
	Desc   string      `json:"desc"`
}

type ResponseDataFormList struct {
	Status string         `json:"status"`
	Msg    string         `json:"msg"`
	Data   []FormsManager `json:"data"`
	Desc   string         `json:"desc"`
}

type ResponseCreateCase struct {
	Status string      `json:"status"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data,omitempty"`
	Desc   string      `json:"desc"`
	CaseID string      `json:"caseId"`
}
