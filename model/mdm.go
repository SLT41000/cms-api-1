package model

import "time"

type MmdProperty struct {
	ID        string    `json:"id" db:"id"`
	PropID    string    `json:"propId" db:"propId"`
	OrgID     string    `json:"orgId" db:"orgId"`
	EN        string    `json:"en" db:"en"`
	TH        string    `json:"th" db:"th"`
	Active    bool      `json:"active" db:"active"`
	CreatedAt time.Time `json:"createdAt" db:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" db:"updatedAt"`
	CreatedBy string    `json:"createdBy" db:"createdBy"`
	UpdatedBy string    `json:"updatedBy" db:"updatedBy"`
}

type MmdPropertyInsert struct {
	EN     string `json:"en" db:"en"`
	TH     string `json:"th" db:"th"`
	Active bool   `json:"active" db:"active"`
}

type MmdPropertyUpdate struct {
	EN     string `json:"en" db:"en"`
	TH     string `json:"th" db:"th"`
	Active bool   `json:"active" db:"active"`
}

type MmdUnitSource struct {
	ID           string    `json:"id" db:"id"`
	UnitSourceID string    `json:"unitSourceId" db:"unitSourceId"`
	OrgID        string    `json:"orgId" db:"orgId"`
	EN           string    `json:"en" db:"en"`
	TH           string    `json:"th" db:"th"`
	Active       bool      `json:"active" db:"active"`
	CreatedAt    time.Time `json:"createdAt" db:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updatedAt"`
	CreatedBy    string    `json:"createdBy" db:"createdBy"`
	UpdatedBy    string    `json:"updatedBy" db:"updatedBy"`
}

type MmdUnitSourceInsert struct {
	EN     string `json:"en" db:"en"`
	TH     string `json:"th" db:"th"`
	Active bool   `json:"active" db:"active"`
}

type MmdUnitSourceUpdate struct {
	EN     string `json:"en" db:"en"`
	TH     string `json:"th" db:"th"`
	Active bool   `json:"active" db:"active"`
}

type MmdUnitType struct {
	ID         string    `json:"id" db:"id"`
	UnitTypeId string    `json:"unitTypeId" db:"unitTypeId"`
	OrgID      string    `json:"orgId" db:"orgId"`
	EN         string    `json:"en" db:"en"`
	TH         string    `json:"th" db:"th"`
	Active     bool      `json:"active" db:"active"`
	CreatedAt  time.Time `json:"createdAt" db:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt" db:"updatedAt"`
	CreatedBy  string    `json:"createdBy" db:"createdBy"`
	UpdatedBy  string    `json:"updatedBy" db:"updatedBy"`
}

type MmdUnitTypeInsert struct {
	EN     string `json:"en" db:"en"`
	TH     string `json:"th" db:"th"`
	Active bool   `json:"active" db:"active"`
}

type MmdUnitTypeUpdate struct {
	EN     string `json:"en" db:"en"`
	TH     string `json:"th" db:"th"`
	Active bool   `json:"active" db:"active"`
}

type MmdCompanies struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	LegalName   string                 `json:"legalName" db:"legalName"`
	Domain      string                 `json:"domain" db:"domain"`
	Email       string                 `json:"email" db:"email"`
	PhoneNumber string                 `json:"phoneNumber" db:"phoneNumber"`
	Address     map[string]interface{} `json:"address" db:"address"`
	LogoURL     string                 `json:"logoUrl" db:"logoUrl"`
	WebsiteURL  string                 `json:"websiteUrl" db:"websiteUrl"`
	Description string                 `json:"description" db:"description"`
	CreatedAt   time.Time              `json:"createdAt" db:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt" db:"updatedAt"`
	CreatedBy   string                 `json:"createdBy" db:"createdBy"`
	UpdatedBy   string                 `json:"updatedBy" db:"updatedBy"`
}

type MmdCompaniesInsert struct {
	Name        string                 `json:"name" db:"name"`
	LegalName   string                 `json:"legalName" db:"legalName"`
	Domain      string                 `json:"domain" db:"domain"`
	Email       string                 `json:"email" db:"email"`
	PhoneNumber string                 `json:"phoneNumber" db:"phoneNumber"`
	Address     map[string]interface{} `json:"address" db:"address"`
	LogoURL     string                 `json:"logoUrl" db:"logoUrl"`
	WebsiteURL  string                 `json:"websiteUrl" db:"websiteUrl"`
	Description string                 `json:"description" db:"description"`
}

type MmdCompaniesUpdate struct {
	Name        string                 `json:"name" db:"name"`
	LegalName   string                 `json:"legalName" db:"legalName"`
	Domain      string                 `json:"domain" db:"domain"`
	Email       string                 `json:"email" db:"email"`
	PhoneNumber string                 `json:"phoneNumber" db:"phoneNumber"`
	Address     map[string]interface{} `json:"address" db:"address"`
	LogoURL     string                 `json:"logoUrl" db:"logoUrl"`
	WebsiteURL  string                 `json:"websiteUrl" db:"websiteUrl"`
	Description string                 `json:"description" db:"description"`
}

type MmdUnitStatus struct {
	ID        string    `json:"id" db:"id"`
	SttID     string    `json:"sttId" db:"sttId"`
	SttName   string    `json:"sttName" db:"sttName"`
	CreatedAt time.Time `json:"createdAt" db:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" db:"updatedAt"`
	CreatedBy string    `json:"createdBy" db:"createdBy"`
	UpdatedBy string    `json:"updatedBy" db:"updatedBy"`
}

type MmdUnitStatusInsert struct {
	SttID   string `json:"sttId" db:"sttId"`
	SttName string `json:"sttName" db:"sttName"`
}

type MmdUnitStatusUpdate struct {
	SttID   string `json:"sttId" db:"sttId"`
	SttName string `json:"sttName" db:"sttName"`
}

type MmdUnit struct {
	ID                string     `json:"id" db:"id"`
	OrgID             string     `json:"orgId" db:"orgid"`
	UnitID            string     `json:"unitId" db:"unitid"`
	UnitName          string     `json:"unitName" db:"unitname"`
	UnitSourceID      string     `json:"unitSourceId" db:"unitsourceid"`
	UnitTypeID        string     `json:"unitTypeId" db:"unittypeid"`
	Priority          int        `json:"priority" db:"priority"`
	CompID            string     `json:"compId" db:"compid"`
	DeptID            string     `json:"deptId" db:"deptid"`
	CommID            string     `json:"commId" db:"commid"`
	StnID             string     `json:"stnId" db:"stnid"`
	PlateNo           *string    `json:"plateNo" db:"plateno"`
	ProvinceCode      *string    `json:"provinceCode" db:"provincecode"`
	Active            bool       `json:"active" db:"active"`
	Username          string     `json:"username" db:"username"`
	IsLogin           *bool      `json:"isLogin" db:"islogin"`
	IsFreeze          *bool      `json:"isFreeze" db:"isfreeze"`
	IsOutArea         *bool      `json:"isOutArea" db:"isoutarea"`
	LocLat            *float64   `json:"locLat" db:"loclat"`
	LocLon            *float64   `json:"locLon" db:"loclon"`
	LocAlt            *float64   `json:"locAlt" db:"localt"`
	LocBearing        *float64   `json:"locBearing" db:"locbearing"`
	LocSpeed          *float64   `json:"locSpeed" db:"locspeed"`
	LocProvider       *string    `json:"locProvider" db:"locprovider"`
	LocGpsTime        *time.Time `json:"locGpsTime" db:"locgpstime"`
	LocSatellites     *int       `json:"locSatellites" db:"locsatellites"`
	LocAccuracy       *float64   `json:"locAccuracy" db:"locaccuracy"`
	LocLastUpdateTime *time.Time `json:"locLastUpdateTime" db:"loclastupdatetime"`
	BreakDuration     *int       `json:"breakDuration" db:"breakduration"`
	HealthChk         *string    `json:"healthChk" db:"healthchk"`
	HealthChkTime     *time.Time `json:"healthChkTime" db:"healthchktime"`
	SttID             *string    `json:"sttId" db:"sttid"`
	CreatedAt         time.Time  `json:"createdAt" db:"createdat"`
	UpdatedAt         time.Time  `json:"updatedAt" db:"updatedat"`
	CreatedBy         string     `json:"createdBy" db:"createdby"`
	UpdatedBy         string     `json:"updatedBy" db:"updatedby"`
}

type MmdUnitInsert struct {
	OrgID             string     `json:"orgId" db:"orgid"`
	UnitID            string     `json:"unitId" db:"unitid"`
	UnitName          string     `json:"unitName" db:"unitname"`
	UnitSourceID      string     `json:"unitSourceId" db:"unitsourceid"`
	UnitTypeID        string     `json:"unitTypeId" db:"unittypeid"`
	Priority          int        `json:"priority" db:"priority"`
	CompID            string     `json:"compId" db:"compid"`
	DeptID            string     `json:"deptId" db:"deptid"`
	CommID            string     `json:"commId" db:"commid"`
	StnID             string     `json:"stnId" db:"stnid"`
	PlateNo           *string    `json:"plateNo" db:"plateno"`
	ProvinceCode      *string    `json:"provinceCode" db:"provincecode"`
	Active            bool       `json:"active" db:"active"`
	Username          *string    `json:"username" db:"username"`
	IsLogin           *bool      `json:"isLogin" db:"islogin"`
	IsFreeze          *bool      `json:"isFreeze" db:"isfreeze"`
	IsOutArea         *bool      `json:"isOutArea" db:"isoutarea"`
	LocLat            *float64   `json:"locLat" db:"loclat"`
	LocLon            *float64   `json:"locLon" db:"loclon"`
	LocAlt            *float64   `json:"locAlt" db:"localt"`
	LocBearing        *float64   `json:"locBearing" db:"locbearing"`
	LocSpeed          *float64   `json:"locSpeed" db:"locspeed"`
	LocProvider       *string    `json:"locProvider" db:"locprovider"`
	LocGpsTime        *time.Time `json:"locGpsTime" db:"locgpstime"`
	LocSatellites     *int       `json:"locSatellites" db:"locsatellites"`
	LocAccuracy       *float64   `json:"locAccuracy" db:"locaccuracy"`
	LocLastUpdateTime *time.Time `json:"locLastUpdateTime" db:"loclastupdatetime"`
	BreakDuration     *int       `json:"breakDuration" db:"breakduration"`
	HealthChk         *string    `json:"healthChk" db:"healthchk"`
	HealthChkTime     *time.Time `json:"healthChkTime" db:"healthchktime"`
	SttID             string     `json:"sttId" db:"sttid"`
}

type MmdUnitUpdate struct {
	UnitID            string     `json:"unitId" db:"unitid"`
	UnitName          string     `json:"unitName" db:"unitname"`
	UnitSourceID      string     `json:"unitSourceId" db:"unitsourceid"`
	UnitTypeID        string     `json:"unitTypeId" db:"unittypeid"`
	Priority          int        `json:"priority" db:"priority"`
	CompID            string     `json:"compId" db:"compid"`
	DeptID            string     `json:"deptId" db:"deptid"`
	CommID            string     `json:"commId" db:"commid"`
	StnID             string     `json:"stnId" db:"stnid"`
	PlateNo           *string    `json:"plateNo" db:"plateno"`
	ProvinceCode      *string    `json:"provinceCode" db:"provincecode"`
	Active            bool       `json:"active" db:"active"`
	Username          *string    `json:"username" db:"username"`
	IsLogin           *bool      `json:"isLogin" db:"islogin"`
	IsFreeze          *bool      `json:"isFreeze" db:"isfreeze"`
	IsOutArea         *bool      `json:"isOutArea" db:"isoutarea"`
	LocLat            *float64   `json:"locLat" db:"loclat"`
	LocLon            *float64   `json:"locLon" db:"loclon"`
	LocAlt            *float64   `json:"locAlt" db:"localt"`
	LocBearing        *float64   `json:"locBearing" db:"locbearing"`
	LocSpeed          *float64   `json:"locSpeed" db:"locspeed"`
	LocProvider       *string    `json:"locProvider" db:"locprovider"`
	LocGpsTime        *time.Time `json:"locGpsTime" db:"locgpstime"`
	LocSatellites     *int       `json:"locSatellites" db:"locsatellites"`
	LocAccuracy       *float64   `json:"locAccuracy" db:"locaccuracy"`
	LocLastUpdateTime *time.Time `json:"locLastUpdateTime" db:"loclastupdatetime"`
	BreakDuration     *int       `json:"breakDuration" db:"breakduration"`
	HealthChk         *string    `json:"healthChk" db:"healthchk"`
	HealthChkTime     *time.Time `json:"healthChkTime" db:"healthchktime"`
	SttID             string     `json:"sttId" db:"sttid"`
}

type MdmUnitProperty struct {
	ID        string    `json:"id" db:"id"`
	OrgID     string    `json:"orgId" db:"orgid"`
	UnitID    string    `json:"unitId" db:"unitid"`
	PropID    string    `json:"propId" db:"propid"`
	Active    bool      `json:"active" db:"active"`
	CreatedAt time.Time `json:"createdAt" db:"createdat"`
	UpdatedAt time.Time `json:"updatedAt" db:"updatedat"`
	CreatedBy string    `json:"createdBy" db:"createdby"`
	UpdatedBy string    `json:"updatedBy" db:"updatedby"`
}
