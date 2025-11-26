package model

import "time"

type AreaDistrictWithDetails struct {
	ID             *string `json:"id"`
	OrgID          *string `json:"orgId"`
	CountryID      *string `json:"countryId"`
	ProvID         *string `json:"provId"`
	DistrictEn     *string `json:"districtEn"`
	DistrictTh     *string `json:"districtTh"`
	DistrictActive *bool   `json:"districtActive"`
	DistID         *string `json:"distId"`

	ProvinceEn     *string `json:"provinceEn"`
	ProvinceTh     *string `json:"provinceTh"`
	ProvinceActive *bool   `json:"provinceActive"`

	CountryEn     *string `json:"countryEn"`
	CountryTh     *string `json:"countryTh"`
	CountryActive *bool   `json:"countryActive"`
	NameSpace     *string `json:"nameSpace"`
}

type AreaDistrict struct {
	ID        *int64     `json:"id"`        // SERIAL -> int64
	OrgID     *string    `json:"orgId"`     // UUID -> string
	CountryID *string    `json:"countryId"` // VARCHAR(10)
	ProvID    *string    `json:"provId"`    // VARCHAR(2)
	DistID    *string    `json:"distId"`    // VARCHAR(3)
	En        *string    `json:"en"`        // VARCHAR(100)
	Th        *string    `json:"th"`        // VARCHAR(100)
	Postcode  *string    `json:"postcode"`  // VARCHAR(10)
	Active    *bool      `json:"active"`    // BOOLEAN
	NameSpace *string    `json:"nameSpace"` // VARCHAR(50)
	CreatedAt *time.Time `json:"createdAt"` // TIMESTAMP
	UpdatedAt *time.Time `json:"updatedAt"` // TIMESTAMP
	CreatedBy *string    `json:"createdBy"` // VARCHAR(50)
	UpdatedBy *string    `json:"updatedBy"` // VARCHAR(50)
}
