package model

import "time"

type Command struct {
	DeptID    string    `json:"deptId"`    // Department ID
	OrgID     string    `json:"orgId"`     // Organization ID
	CommID    string    `json:"commId"`    // Community ID
	En        string    `json:"en"`        // English name
	Th        string    `json:"th"`        // Thai name
	Active    bool      `json:"active"`    // Active status
	CreatedAt time.Time `json:"createdAt"` // Created timestamp
	UpdatedAt time.Time `json:"updatedAt"` // Updated timestamp
	CreatedBy string    `json:"createdBy"` // User who created the record
	UpdatedBy string    `json:"updatedBy"` // User who last updated the record
}
