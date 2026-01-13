package model

import "time"

type CaseType struct {
	Id        string    `json:"id"`
	TypeId    string    `json:"typeId"`
	OrgId     string    `json:"orgId"`
	En        *string   `json:"en"`
	Th        *string   `json:"th"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	CreatedBy *string   `json:"createdBy"`
	UpdatedBy *string   `json:"updatedBy"`
}

type CaseTypeInsert struct {
	En     *string `json:"en"`
	Th     *string `json:"th"`
	Active bool    `json:"active"`
}

type CaseTypeUpdate struct {
	En     *string `json:"en"`
	Th     *string `json:"th"`
	Active bool    `json:"active"`
}

type CaseSubType struct {
	Id              string    `json:"id"`
	TypeID          string    `json:"typeId"`
	STypeID         string    `json:"sTypeId"`
	STypeCode       string    `json:"sTypeCode"`
	OrgID           string    `json:"orgId"`
	EN              string    `json:"en"`
	TH              string    `json:"th"`
	WFID            string    `json:"wfId"`
	CaseSLA         string    `json:"caseSla"`
	Priority        string    `json:"priority"`
	UserSkillList   []string  `json:"userSkillList"`
	UnitPropLists   []string  `json:"unitPropLists"`
	Active          bool      `json:"active"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	CreatedBy       string    `json:"createdBy"`
	UpdatedBy       string    `json:"updatedBy"`
	MDeviceType     *string   `json:"mDeviceType"`
	MWorkOrderType  *string   `json:"mWorkOrderType"`
	MDeviceTypeName *string   `json:"mDeviceTypeName"`
}

type CaseSubTypeInsert struct {
	TypeID          string   `json:"typeId"`
	STypeCode       string   `json:"sTypeCode"`
	EN              string   `json:"en"`
	TH              string   `json:"th"`
	WFID            string   `json:"wfId"`
	CaseSLA         string   `json:"caseSla"`
	Priority        string   `json:"priority"`
	UserSkillList   []string `json:"userSkillList"`
	UnitPropLists   []string `json:"unitPropLists"`
	Active          bool     `json:"active"`
	MDeviceType     *string  `json:"mDeviceType"`
	MWorkOrderType  *string  `json:"mWorkOrderType"`
	MDeviceTypeName *string  `json:"mDeviceTypeName"`
}

type CaseSubTypeUpdate struct {
	TypeID          string   `json:"typeId"`
	STypeCode       string   `json:"sTypeCode"`
	EN              string   `json:"en"`
	TH              string   `json:"th"`
	WFID            string   `json:"wfId"`
	CaseSLA         string   `json:"caseSla"`
	Priority        string   `json:"priority"`
	UserSkillList   []string `json:"userSkillList"`
	UnitPropLists   []string `json:"unitPropLists"`
	Active          bool     `json:"active"`
	MDeviceType     *string  `json:"mDeviceType"`
	MWorkOrderType  *string  `json:"mWorkOrderType"`
	MDeviceTypeName *string  `json:"mDeviceTypeName"`
}

type CaseTypeWithSubType struct {
	TypeID        string      `json:"typeId" db:"typeId"`
	OrgID         string      `json:"orgId" db:"orgId"`
	TypeEN        string      `json:"en" db:"en"`
	TypeTH        string      `json:"th" db:"th"`
	TypeActive    bool        `json:"active" db:"active"`
	SubTypeID     *string     `json:"sTypeId" db:"sTypeId"`
	SubTypeCode   *string     `json:"sTypeCode" db:"sTypeCode"`
	SubTypeEN     *string     `json:"subTypeEn" db:"en"`
	SubTypeTH     *string     `json:"subTypeTh" db:"th"`
	WfID          *string     `json:"wfId" db:"wfId"`
	CaseSla       *string     `json:"caseSla" db:"caseSla"`
	Priority      *int        `json:"priority" db:"priority"`
	UserSkillList interface{} `json:"userSkillList" db:"userSkillList"`
	UnitPropLists interface{} `json:"unitPropLists" db:"unitPropLists"`
	SubTypeActive *bool       `json:"subTypeActive" db:"active"`
}
