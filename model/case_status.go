package model

import (
	"time"
)

type CaseStatus struct {
	ID        *string   `json:"id" db:"id"`
	StatusID  *string   `json:"statusId" db:"statusId"`
	Th        *string   `json:"th" db:"th"`
	En        *string   `json:"en" db:"en"`
	Color     *string   `json:"color" db:"color"`
	Active    bool      `json:"active" db:"active"`
	CreatedAt time.Time `json:"createdAt" db:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" db:"updatedAt"`
	CreatedBy *string   `json:"createdBy" db:"createdBy"`
	UpdatedBy *string   `json:"updatedBy" db:"updatedBy"`
}

type CaseStatusInsert struct {
	Th     *string `json:"th" db:"th"`
	En     *string `json:"en" db:"en"`
	Color  *string `json:"color" db:"color"`
	Active bool    `json:"active" db:"active"`
}

type CaseStatusUpdate struct {
	Th     *string `json:"th" db:"th"`
	En     *string `json:"en" db:"en"`
	Color  *string `json:"color" db:"color"`
	Active bool    `json:"active" db:"active"`
}
