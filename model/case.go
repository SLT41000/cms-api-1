package model

import (
	"encoding/json"
	"time"
)

type Case struct {
	ID              string          `json:"id"`
	OrgID           string          `json:"orgId"`
	CaseID          string          `json:"caseId"`
	CaseVersion     string          `json:"caseVersion"`
	ReferCaseID     *string         `json:"referCaseId"`
	CaseTypeID      string          `json:"caseTypeId"`
	CaseSTypeID     string          `json:"caseSTypeId"`
	Priority        int             `json:"priority"`
	WfID            *string         `json:"wfId"`
	WfVersions      *string         `json:"versions"`
	Source          string          `json:"source"`
	DeviceID        *string         `json:"deviceId"`
	PhoneNo         *string         `json:"phoneNo"`
	PhoneNoHide     bool            `json:"phoneNoHide"`
	CaseDetail      *string         `json:"caseDetail"`
	ExtReceive      *string         `json:"extReceive"`
	StatusID        string          `json:"statusId"`
	CaseLat         *string         `json:"caseLat"`
	CaseLon         *string         `json:"caseLon"`
	CaseLocAddr     *string         `json:"caselocAddr"`
	CaseLocAddrDecs *string         `json:"caselocAddrDecs"`
	CountryID       string          `json:"countryId"`
	ProvID          string          `json:"provId"`
	DistID          string          `json:"distId"`
	CaseDuration    int             `json:"caseDuration"`
	CreatedDate     *time.Time      `json:"createdDate"`
	StartedDate     *time.Time      `json:"startedDate"`
	CommandedDate   *time.Time      `json:"commandedDate"`
	ReceivedDate    *time.Time      `json:"receivedDate"`
	ArrivedDate     *time.Time      `json:"arrivedDate"`
	ClosedDate      *time.Time      `json:"closedDate"`
	UserCreate      *string         `json:"usercreate"`
	UserCommand     *string         `json:"usercommand"`
	UserReceive     *string         `json:"userreceive"`
	UserArrive      *string         `json:"userarrive"`
	UserClose       *string         `json:"userclose"`
	ResID           *string         `json:"resId"`
	ResDetail       *string         `json:"resDetail"`
	ScheduleFlag    *bool           `json:"scheduleFlag"`
	ScheduleDate    *time.Time      `json:"scheduleDate"`
	CreatedAt       *time.Time      `json:"createdAt"`
	UpdatedAt       *time.Time      `json:"updatedAt"`
	CreatedBy       *string         `json:"createdBy"`
	UpdatedBy       *string         `json:"updatedBy"`
	SOP             interface{}     `json:"sop"`
	CurrentStage    interface{}     `json:"currentStage"`
	NextStage       interface{}     `json:"nextStage"`
	DispatchStage   interface{}     `json:"dispatchStage"`
	ReferCaseLists  []string        `json:"referCaseLists"`
	UnitLists       interface{}     `json:"unitLists"`
	FormAnswer      interface{}     `json:"formAnswer"`
	SlaTimelines    interface{}     `json:"slaTimelines"`
	CaseSLA         *int            `json:"caseSla" `
	DeviceMetaData  json.RawMessage `db:"deviceMetaData" json:"deviceMetaData"`
}

type CaseInsert struct {
	CaseId          *string            `json:"caseId" `
	CaseVersion     string             `json:"caseVersion" binding:"required"`
	ReferCaseID     *string            `json:"referCaseId"`
	CaseTypeID      string             `json:"caseTypeId"`
	CaseSTypeID     string             `json:"caseSTypeId"`
	Priority        int                `json:"priority"`
	WfID            *string            `json:"wfId"`
	WfVersions      *string            `json:"versions"`
	NodeID          string             `json:"nodeId" db:"nodeId"`
	Source          string             `json:"source"`
	DeviceID        *string            `json:"deviceId"`
	PhoneNo         *string            `json:"phoneNo"`
	PhoneNoHide     bool               `json:"phoneNoHide"`
	CaseDetail      *string            `json:"caseDetail"`
	ExtReceive      *string            `json:"extReceive"`
	StatusID        string             `json:"statusId"`
	CaseLat         *string            `json:"caseLat"`
	CaseLon         *string            `json:"caseLon"`
	CaseLocAddr     *string            `json:"caselocAddr"`
	CaseLocAddrDecs *string            `json:"caselocAddrDecs"`
	CountryID       string             `json:"countryId"`
	ProvID          string             `json:"provId"`
	DistID          string             `json:"distId"`
	CaseDuration    int                `json:"caseDuration"`
	CreatedDate     *time.Time         `json:"createdDate"`
	StartedDate     *time.Time         `json:"startedDate"`
	CommandedDate   *time.Time         `json:"commandedDate"`
	ReceivedDate    *time.Time         `json:"receivedDate"`
	ArrivedDate     *time.Time         `json:"arrivedDate"`
	ClosedDate      *time.Time         `json:"closedDate"`
	UserCreate      *string            `json:"usercreate"`
	UserCommand     *string            `json:"usercommand"`
	UserReceive     *string            `json:"userreceive"`
	UserArrive      *string            `json:"userarrive"`
	UserClose       *string            `json:"userclose"`
	ResID           *string            `json:"resId"`
	ResDetail       *string            `json:"resDetail"`
	ScheduleFlag    *bool              `json:"scheduleFlag"`
	ScheduleDate    *time.Time         `json:"scheduleDate"`
	FormData        *FormAnswerRequest `json:"formData"`
}

type CaseUpdate struct {
	CaseId          *string    `json:"caseId" `
	CaseVersion     string     `json:"caseVersion"`
	ReferCaseID     *string    `json:"referCaseId"`
	CaseTypeID      string     `json:"caseTypeId"`
	CaseSTypeID     string     `json:"caseSTypeId"`
	Priority        int        `json:"priority"`
	WfID            string     `json:"wfId" db:"wfId"`
	Source          string     `json:"source"`
	DeviceID        string     `json:"deviceId"`
	PhoneNo         string     `json:"phoneNo"`
	PhoneNoHide     bool       `json:"phoneNoHide"`
	CaseDetail      *string    `json:"caseDetail"`
	ExtReceive      string     `json:"extReceive"`
	StatusID        string     `json:"statusId"`
	CaseLat         string     `json:"caseLat"`
	CaseLon         string     `json:"caseLon"`
	CaseLocAddr     string     `json:"caselocAddr"`
	CaseLocAddrDecs string     `json:"caselocAddrDecs"`
	CountryID       string     `json:"countryId"`
	ProvID          string     `json:"provId"`
	DistID          string     `json:"distId"`
	CaseDuration    int        `json:"caseDuration"`
	CreatedDate     time.Time  `json:"createdDate"`
	StartedDate     time.Time  `json:"startedDate"`
	CommandedDate   time.Time  `json:"commandedDate"`
	ReceivedDate    time.Time  `json:"receivedDate"`
	ArrivedDate     time.Time  `json:"arrivedDate"`
	ClosedDate      time.Time  `json:"closedDate"`
	UserCreate      string     `json:"usercreate"`
	UserCommand     string     `json:"usercommand"`
	UserReceive     string     `json:"userreceive"`
	UserArrive      string     `json:"userarrive"`
	UserClose       string     `json:"userclose"`
	ResID           *string    `json:"resId"`
	ResDetail       *string    `json:"resDetail"`
	ScheduleFlag    *bool      `json:"scheduleFlag"`
	ScheduleDate    *time.Time `json:"scheduleDate"`
}

type OutputCase struct {
	Id                    string     `json:"id"`
	CaseId                *string    `json:"caseId"`
	Casetype_code         *string    `json:"CasetypeCode"`
	Priority              *int       `json:"priority"`
	Phone_number          *string    `json:"phoneNumber"`
	Case_status_code      *string    `json:"caseStatusCode"`
	Case_status_name      *string    `json:"caseStatusName"`
	Case_detail           *string    `json:"caseDetail"`
	Police_station_name   *string    `json:"stationName"`
	Case_location_address *string    `json:"caseLocationAddress"`
	Case_location_detail  *string    `json:"caseLocationDetail"`
	Special_emergency     *int       `json:"specialEmergency"`
	Urgent_amount         *string    `json:"urgentAmount"`
	Created_date          *time.Time `json:"createdDate"`
	Casetype_name         *string    `json:"casetypeName"`
	Media_type            *int       `json:"mediaType"`
	VOwner                *int       `json:"vOwner"`
	VVin                  *string    `json:"vVin"`
}

type CaseListData struct {
	Draw            int          `json:"draw"`
	RecordsTotal    int          `json:"recordsTotal"`
	RecordsFiltered int          `json:"recordsFiltered"`
	Data            []OutputCase `json:"data"`
	Error           string       `json:"error"`
}

type CaseDetailData struct {
	ID                  int        `json:"id"`
	CaseID              string     `json:"caseId"`
	ReferCaseID         *string    `json:"referCaseId"`
	CasetypeCode        string     `json:"casetypeCode"`
	Priority            int        `json:"priority"`
	Ways                int        `json:"ways"`
	PhoneNumber         string     `json:"phoneNumber"`
	CaseStatusCode      string     `json:"caseStatusCode"`
	CaseStatusName      string     `json:"caseStatusName"`
	CaseDetail          string     `json:"caseDetail"`
	DepartmentName      *string    `json:"departmentName"`
	CommandCode         *string    `json:"commandCode"`
	CommandName         *string    `json:"commandName"`
	PoliceStationCode   *string    `json:"policeStationCode"`
	PoliceStationName   *string    `json:"policeStationName"`
	CaseLocationAddress *string    `json:"caseLocationAddress"`
	CaseLocationDetail  *string    `json:"caseLocationDetail"`
	CaseLat             *string    `json:"caseLat"`
	CaseLon             *string    `json:"caseLon"`
	TransImg            string     `json:"transImg"`
	CitizenCode         string     `json:"citizenCode"`
	ExtensionReceive    string     `json:"extensionReceive"`
	SpecialEmergency    int        `json:"specialEmergency"`
	UrgentAmount        *string    `json:"urgentAmount"`
	OpenedDate          time.Time  `json:"openedDate"`
	SavedDate           *time.Time `json:"savedDate"`
	CreatedDate         time.Time  `json:"createdDate"`
	StartedDate         time.Time  `json:"startedDate"`
	ModifiedDate        time.Time  `json:"modifiedDate"`
	UserCreate          string     `json:"userCreate"`
	UserCreateID        *string    `json:"userCreateId"`
	UserModify          string     `json:"userModify"`
	Responsible         *string    `json:"responsible"`
	ApprovedStatus      *int       `json:"approvedStatus"`
	CasetypeName        string     `json:"casetypeName"`
	MediaType           int        `json:"mediaType"`
	VOwner              int        `json:"vOwner"`
	VVin                string     `json:"vVin"`
	DestLocationAddress *string    `json:"destLocationAddress"`
	DestLocationDetail  *string    `json:"destLocationDetail"`
	DestLat             *string    `json:"destLat"`
	DestLon             *string    `json:"destLon"`
	CitizenFullname     string     `json:"citizenFullname"`
}

type CaseResponse struct {
	Status string          `json:"status"`
	Msg    string          `json:"msg"`
	Data   *CaseDetailData `json:"data,omitempty"`
	Desc   string          `json:"desc"`
}

type CaseListResponse struct {
	Status string       `json:"status"`
	Msg    string       `json:"msg"`
	Data   CaseListData `json:"data,omitempty"`
	Desc   string       `json:"desc"`
}

type CaseForCreate struct {
	ReferCaseID         string    `json:"referCaseId" `
	CasetypeCode        string    `json:"casetypeCode" `
	Priority            int       `json:"priority" `
	Ways                int       `json:"ways" `
	PhoneNumber         string    `json:"phoneNumber" `
	PhoneNumberHide     int       `json:"phoneNumberHide" `
	Duration            int       `json:"duration" `
	CaseStatusCode      string    `json:"caseStatusCode" `
	CaseCondition       int       `json:"caseCondition" `
	CaseDetail          string    `json:"caseDetail" `
	CaseLocationType    string    `json:"caseLocationType" `
	CommandCode         string    `json:"commandCode" `
	PoliceStationCode   string    `json:"policeStationCode" `
	CaseLocationAddress string    `json:"caseLocationAddress" `
	CaseLocationDetail  string    `json:"caseLocationDetail" `
	CaseRoute           string    `json:"caseRoute" `
	CaseLat             string    `json:"caseLat" `
	CaseLon             string    `json:"caseLon" `
	CaseDirection       string    `json:"caseDirection" `
	CasePhoto           []Media   `json:"casePhoto" `
	TransImg            []Media   `json:"transImg" `
	Home                int       `json:"home" `
	CitizenCode         string    `json:"citizenCode" `
	ExtensionReceive    string    `json:"extensionReceive" `
	CaseSLA             int       `json:"caseSla" `
	ActionProCode       string    `json:"actionProCode" `
	SpecialEmergency    int       `json:"specialEmergency" `
	MediaCode           string    `json:"mediaCode" `
	MediaType           int       `json:"mediaType" `
	OpenedDate          time.Time `json:"openedDate"`
	CreatedDate         time.Time `json:"createdDate"`
	StartedDate         time.Time `json:"startedDate"`
	ClosedDate          time.Time `json:"closedDate"`
	ModifiedDate        time.Time `json:"modifiedDate" `
	UserCreate          string    `json:"userCreate" `
	UserClose           string    `json:"userClose" `
	UserModify          string    `json:"userModify" `
	NeedAmbulance       int       `json:"needAmbulance" `
	Backdated           int       `json:"backdated" `
	EscapeRoute         string    `json:"escapeRoute" `
	VOwner              int       `json:"vOwner" `
	VVin                string    `json:"vVin" `
	DestLocationAddress string    `json:"destLocationAddress" `
	DestLocationDetail  string    `json:"destLocationDetail" `
	DestLat             string    `json:"destLat" `
	DestLon             string    `json:"destLon" `
	Token               string    `json:"token" `
}

func (c *CaseForCreate) SetDefaults() {
	now := time.Now()
	if c.OpenedDate.IsZero() {
		c.OpenedDate = now
	}
	if c.CreatedDate.IsZero() {
		c.CreatedDate = now
	}
	if c.StartedDate.IsZero() {
		c.StartedDate = now
	}
	if c.ClosedDate.IsZero() {
		c.ClosedDate = now
	}
	if c.ModifiedDate.IsZero() {
		c.ModifiedDate = now
	}
}

type CaseForUpdate struct {
	ReferCaseID         string    `json:"referCaseId"`
	CasetypeCode        string    `json:"casetypeCode"`
	Priority            int       `json:"priority"`
	Ways                int       `json:"ways"`
	PhoneNumber         string    `json:"phoneNumber"`
	PhoneNumberHide     int       `json:"phoneNumberHide"`
	Duration            int       `json:"duration"`
	CaseStatusCode      string    `json:"caseStatusCode"`
	CaseCondition       int       `json:"caseCondition"`
	CaseDetail          string    `json:"caseDetail"`
	CaseLocationType    string    `json:"caseLocationType"`
	CommandCode         string    `json:"commandCode"`
	PoliceStationCode   string    `json:"policeStationCode"`
	CaseLocationAddress string    `json:"caseLocationAddress"`
	CaseLocationDetail  string    `json:"caseLocationDetail"`
	CaseRoute           string    `json:"caseRoute"`
	CaseLat             string    `json:"caseLat"`
	CaseLon             string    `json:"caseLon"`
	CaseDirection       string    `json:"caseDirection"`
	CasePhoto           []Media   `json:"casePhoto"`
	TransImg            []Media   `json:"transImg"`
	Home                int       `json:"home"`
	CitizenCode         string    `json:"citizenCode"`
	ExtensionReceive    string    `json:"extensionReceive"`
	CaseSLA             int       `json:"caseSla"`
	ActionProCode       string    `json:"actionProCode"`
	SpecialEmergency    int       `json:"specialEmergency"`
	MediaCode           string    `json:"mediaCode"`
	MediaType           int       `json:"mediaType"`
	OpenedDate          time.Time `json:"openedDate"`
	CreatedDate         time.Time `json:"createdDate"`
	StartedDate         time.Time `json:"startedDate"`
	ClosedDate          time.Time `json:"closedDate"`
	ModifiedDate        time.Time `json:"modifiedDate"`
	UserCreate          string    `json:"userCreate"`
	UserClose           string    `json:"userClose"`
	UserModify          string    `json:"userModify"`
	NeedAmbulance       int       `json:"needAmbulance"`
	Backdated           int       `json:"backdated"`
	EscapeRoute         string    `json:"escapeRoute"`
	VOwner              int       `json:"vOwner"`
	VVin                string    `json:"vVin"`
	DestLocationAddress string    `json:"destLocationAddress"`
	DestLocationDetail  string    `json:"destLocationDetail"`
	DestLat             string    `json:"destLat"`
	DestLon             string    `json:"destLon"`
}
type Media struct {
	URL string `json:"url"`
}

type CaseCloseInput struct {
	CaseStatusCode string    `json:"caseStatusCode"`
	ResultCode     string    `json:"resultCode"`
	ResultDetail   string    `json:"resultDetail"`
	TransImg       []Media   `json:"transImg"`
	ClosedDate     time.Time `json:"closedDate"`
	ModifiedDate   time.Time `json:"modifiedDate"`
	UserClose      string    `json:"userClose"`
	UserModify     string    `json:"userModify"`
}

type CreateCaseResponse struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
	Desc   string `json:"desc"`
	ID     string `json:"id"`
	CaseID string `json:"caseId"`
}

type UpdateCaseResponse struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
	Desc   string `json:"desc"`
	ID     int    `json:"id"`
}

type DeleteCaseResponse struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
	Desc   string `json:"desc"`
}

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
	Id            string    `json:"id"`
	TypeID        string    `json:"typeId"`
	STypeID       string    `json:"sTypeId"`
	STypeCode     string    `json:"sTypeCode"`
	OrgID         string    `json:"orgId"`
	EN            string    `json:"en"`
	TH            string    `json:"th"`
	WFID          string    `json:"wfId"`
	CaseSLA       string    `json:"caseSla"`
	Priority      string    `json:"priority"`
	UserSkillList []string  `json:"userSkillList"`
	UnitPropLists []string  `json:"unitPropLists"`
	Active        bool      `json:"active"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	CreatedBy     string    `json:"createdBy"`
	UpdatedBy     string    `json:"updatedBy"`
}

type CaseSubTypeInsert struct {
	TypeID        string   `json:"typeId"`
	STypeCode     string   `json:"sTypeCode"`
	EN            string   `json:"en"`
	TH            string   `json:"th"`
	WFID          string   `json:"wfId"`
	CaseSLA       string   `json:"caseSla"`
	Priority      string   `json:"priority"`
	UserSkillList []string `json:"userSkillList"`
	UnitPropLists []string `json:"unitPropLists"`
	Active        bool     `json:"active"`
}

type CaseSubTypeUpdate struct {
	STypeCode     string   `json:"sTypeCode"`
	EN            string   `json:"en"`
	TH            string   `json:"th"`
	WFID          string   `json:"wfId"`
	CaseSLA       string   `json:"caseSla"`
	Priority      string   `json:"priority"`
	UserSkillList []string `json:"userSkillList"`
	UnitPropLists []string `json:"unitPropLists"`
	Active        bool     `json:"active"`
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
