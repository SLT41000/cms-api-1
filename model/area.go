package model

type AreaDistrictWithDetails struct {
	ID             *string `json:"id"`
	OrgID          *string `json:"orgId"`
	CountryID      *string `json:"countryId"`
	ProvID         *string `json:"provId"`
	DistrictEn     *string `json:"districtEn"`
	DistrictTh     *string `json:"districtTh"`
	DistrictActive *bool   `json:"districtActive"`

	ProvinceEn     *string `json:"provinceEn"`
	ProvinceTh     *string `json:"provinceTh"`
	ProvinceActive *bool   `json:"provinceActive"`

	CountryEn     *string `json:"countryEn"`
	CountryTh     *string `json:"countryTh"`
	CountryActive *bool   `json:"countryActive"`
}
