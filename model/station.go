package model

import "time"

type Station struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"orgId"`     // Organization ID
	DeptID    string    `json:"deptId"`    // Department ID
	CommID    string    `json:"commId"`    // Community ID
	StnID     string    `json:"stnId"`     // Station ID
	En        string    `json:"en"`        // English name
	Th        string    `json:"th"`        // Thai name
	Active    bool      `json:"active"`    // Active status
	CreatedAt time.Time `json:"createdAt"` // Created timestamp
	UpdatedAt time.Time `json:"updatedAt"` // Updated timestamp
	CreatedBy string    `json:"createdBy"` // User who created the record
	UpdatedBy string    `json:"updatedBy"` // User who last updated the record
}

type StationInsert struct {
	DeptID string `json:"deptId"` // Department ID
	CommID string `json:"commId"` // Community ID
	En     string `json:"en"`     // English name
	Th     string `json:"th"`     // Thai name
	Active bool   `json:"active"` // Active status\
}

type StationUpdate struct {
	DeptID string `json:"deptId"` // Department ID
	CommID string `json:"commId"` // Community ID
	En     string `json:"en"`     // English name
	Th     string `json:"th"`     // Thai name
	Active bool   `json:"active"` // Active status\
}

type StationWithCommandDept struct {
	ID     string `json:"id"`
	OrgId  string `json:"orgId"`
	DeptId string `json:"deptId"`
	CommId string `json:"commId"`
	StnId  string `json:"stnId"`

	// Station fields
	StationEn     string `json:"stationEn"`
	StationTh     string `json:"stationTh"`
	StationActive bool   `json:"stationActive"`

	// Command fields
	CommandEn     string `json:"commandEn"`
	CommandTh     string `json:"commandTh"`
	CommandActive bool   `json:"commandActive"`

	// Department fields
	DeptEn     string `json:"deptEn"`
	DeptTh     string `json:"deptTh"`
	DeptActive bool   `json:"deptActive"`
}
