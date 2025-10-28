package model

type CaseGroupType struct {
	ID             int      `json:"id"`
	OrgId          string   `json:"orgId"`
	GroupTypeId    string   `json:"groupTypeId"`
	En             string   `json:"en"`
	Th             string   `json:"th"`
	GroupTypeLists []string `json:"groupTypeLists"`
	Prefix         string   `json:"prefix"`
}

type DCaseSummary struct {
	ID          int    `json:"id"`
	OrgId       string `json:"orgId"`
	Date        string `json:"date"`
	Time        string `json:"time"`
	GroupTypeId string `json:"groupTypeId"`
	Total       int    `json:"total"`
}

// Prepare final JSON structure
type DashboardSummary struct {
	Type    string        `json:"type"`
	TitleEn string        `json:"title_en"`
	TitleTh string        `json:"title_th"`
	Data    []interface{} `json:"data"`
}
