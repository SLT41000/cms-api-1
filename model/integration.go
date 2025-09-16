package model

import "time"

type WorkOrder struct {
	WorkOrderNumber       string        `json:"work_order_number"`
	ParentWorkOrderNumber string        `json:"parent_work_order_number"`
	WorkOrderRefNumber    string        `json:"work_order_ref_number"`
	IncidentNumber        string        `json:"incident_number"`
	WorkOrderType         string        `json:"work_order_type"`
	WorkOrderMetadata     WorkOrderMeta `json:"work_order_metadata"`
	UserMetadata          UserMeta      `json:"user_metadata"`
	DeviceMetadata        DeviceMeta    `json:"device_metadata"`
	SopMetadata           SopMeta       `json:"sop_metadata"`
	Status                string        `json:"status"`
	WorkDate              string        `json:"work_date"`
	Workspace             string        `json:"workspace"`
	Namespace             string        `json:"namespace"`
}

type WorkOrderMeta struct {
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Severity    string      `json:"severity"`
	Location    GeoLocation `json:"location"`
	Images      []string    `json:"images"`
}

type UserMeta struct {
	AssignedEmployeeCode  string   `json:"assigned_employee_code"`
	AssociateEmployeeCode []string `json:"associate_employee_code"`
}

type DeviceMeta struct {
	DeviceID           string      `json:"device_id"`
	DeviceName         string      `json:"device_name"`
	DeviceType         string      `json:"device_type"`
	DeviceSerialNumber string      `json:"device_serial_number"`
	DeviceModel        string      `json:"device_model"`
	DeviceBrand        string      `json:"device_brand"`
	DeviceLocation     GeoLocation `json:"device_location"`
}

type SopMeta struct {
	// ใส่ field เพิ่มเติมถ้ามี
}

type GeoLocation struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}

type WorkflowBySubType struct {
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

	// extra from wf_definitions
	WfTitle    string `json:"wfTitle" db:"wfTitle"`
	WfDesc     string `json:"wfDesc" db:"wfDesc"`
	WfVersions string `json:"wfVersions" db:"wfVersions"`
	WfSection  string `json:"wfSection" db:"wfSection"`
	WfData     string `json:"wfData" db:"wfData"`
	WfNodeId   string `json:"wfNodeId" db:"wfNodeId"`
}
