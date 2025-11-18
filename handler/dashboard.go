package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mainPackage/model"
	"mainPackage/utils"
	"time"

	"github.com/jackc/pgx/v5"
)

func CalDashboardCaseSummary(ctx context.Context, conn *pgx.Conn, orgId string, recipients []model.Recipient, username, caseTypeId, countryId, provId, distId string) error {

	// ใช้เวลาไทย
	//loc, _ := time.LoadLocation("Asia/Bangkok")
	//now := time.Now().In(loc)
	now := getTimeNow()

	currentDate := now.Format("2006/01/02") // 2025/11/12
	currentHour := now.Hour()
	//currentTime := now.Format("%02d:00:00") // 16:00:00 จากเวลาไทยจริง
	currentTime := fmt.Sprintf("%02d:00:00", currentHour) // e.g., "01:00:00"

	// ✅ โหลดข้อมูล groupType จาก Redis หรือ DB
	groupTypes, err := utils.GroupTypeGetOrLoad(conn)
	if err != nil {
		return fmt.Errorf("cannot load group types: %v", err)
	}

	// หาว่า caseTypeId อยู่ group ไหน
	groupTypeId := ""
	found := false
	for _, g := range groupTypes {
		for _, id := range g.GroupTypeLists {
			if id == caseTypeId {
				groupTypeId = g.GroupTypeId
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return fmt.Errorf("groupTypeId not found: %s", caseTypeId)
	}

	// ✅ UPSERT (เพิ่ม +1 ถ้ามีอยู่แล้ว)
	query := `
		INSERT INTO d_case_summary 
			("orgId", date, time, "groupTypeId", "countryId", "provId", "distId", total)
		VALUES ($1, $2, $3, $4, $5, $6, $7, '1')
		ON CONFLICT ("orgId", date, time, "groupTypeId", "countryId", "provId", "distId")
		DO UPDATE SET total = (d_case_summary.total::int + 1)::varchar;
	`

	_, err = conn.Exec(ctx, query,
		orgId,
		currentDate,
		currentTime,
		groupTypeId,
		countryId,
		provId,
		distId,
	)

	if err != nil {
		return fmt.Errorf("failed to upsert summary: %v", err)
	}

	// ส่งแดชบอร์ด
	err = SendDashboardSummaryFromCaseSummary(ctx, conn, orgId, username, recipients)
	if err != nil {
		log.Printf("Dashboard notification error: %v", err)
	}

	log.Printf("✅ Upsert summary success for groupTypeId=%s, countryId=%s, provId=%s, distId=%s", groupTypeId, countryId, provId, distId)
	return nil
}

func SendDashboardSummaryFromCaseSummary(
	c context.Context,
	conn *pgx.Conn,
	orgId string,
	username string,
	recipients []model.Recipient,
) error {
	// ✅ โหลด group type ทั้งหมด
	groupTypes, err := utils.GroupTypeGetOrLoad(conn)
	if err != nil {
		return fmt.Errorf("cannot load group types: %v", err)
	}

	groupMap := make(map[string]struct {
		Prefix string
		En     string
		Th     string
	})
	for _, g := range groupTypes {
		groupMap[g.GroupTypeId] = struct {
			Prefix string
			En     string
			Th     string
		}{
			Prefix: g.Prefix,
			En:     g.En,
			Th:     g.Th,
		}
	}

	// ✅ ดึงข้อมูลจาก d_case_summary
	query := `
		SELECT s."groupTypeId", COALESCE(SUM(CAST(s.total AS INT)), 0) AS total
		FROM d_case_summary s
		WHERE s."orgId" = $1
		  AND s.date = TO_CHAR(CURRENT_DATE, 'YYYY/MM/DD')
		GROUP BY s."groupTypeId"
	`
	rows, err := conn.Query(c, query, orgId)
	if err != nil {
		return fmt.Errorf("query dashboard summary failed: %w", err)
	}
	defer rows.Close()

	type SummaryData struct {
		GroupTypeId string
		Val         int
	}
	summaryMap := make(map[string]int)
	totalSum := 0

	for rows.Next() {
		var item SummaryData
		if err := rows.Scan(&item.GroupTypeId, &item.Val); err != nil {
			return fmt.Errorf("scan dashboard summary failed: %w", err)
		}
		summaryMap[item.GroupTypeId] = item.Val
		totalSum += item.Val
	}

	// ✅ เตรียมผลลัพธ์ JSON
	data := []interface{}{
		map[string]interface{}{
			"total_en": "Total",
			"total_th": "ทั้งหมด",
			"val":      totalSum,
		},
	}

	// ✅ รวม group ทั้งหมด (แม้ไม่มีข้อมูล ก็ให้ val = 0)
	for _, g := range groupTypes {
		val := 0
		if v, ok := summaryMap[g.GroupTypeId]; ok {
			val = v
		}

		data = append(data, map[string]interface{}{
			fmt.Sprintf("%s_en", g.Prefix): g.En,
			fmt.Sprintf("%s_th", g.Prefix): g.Th,
			"val":                          val,
		})
	}

	// ✅ สร้าง payload
	summary := model.DashboardSummary{
		Type:    "CASE-SUMMARY",
		TitleEn: "Work Order Summary",
		TitleTh: "สรุปใบสั่งงาน",
		Data:    data,
	}

	jsonBytes, err := json.Marshal(summary)
	if err != nil {
		return fmt.Errorf("marshal dashboard summary failed: %w", err)
	}
	raw := json.RawMessage(jsonBytes)

	// ✅ ส่ง Notification
	err = genNotiCustom(
		c, conn, orgId, username, username, "",
		"hidden", nil, "", recipients, "", "User", "DASHBOARD", &raw,
	)
	if err != nil {
		return fmt.Errorf("send dashboard notification failed: %w", err)
	}

	log.Printf("✅ Dashboard summary sent successfully: total=%d groups=%d", totalSum, len(groupTypes))
	return nil
}

func CalDashboardSLA(
	ctx context.Context,
	conn *pgx.Conn,
	orgId string,
	username, caseId string,
) error {
	log.Print("--CalDashboardSLA--1")
	now := time.Now()
	currentDate := now.Format("2006/01/02")
	currentTime := fmt.Sprintf("%02d:00:00", now.Hour())

	// load groupType from cache/db
	groupTypes, err := utils.GroupTypeGetOrLoad(conn)
	if err != nil {
		return fmt.Errorf("cannot load group types: %v", err)
	}
	log.Print("--CalDashboardSLA--2")
	// load case data
	item, err := GetCaseClose(ctx, conn, orgId, caseId)
	if err != nil {
		return err
	}
	if item == nil {
		return fmt.Errorf("case not found")
	}
	log.Print("--CalDashboardSLA--3")
	// match groupTypeId
	var groupTypeId string
	found := false

	for _, g := range groupTypes {
		log.Print(groupTypes, g)
		for _, id := range g.GroupTypeLists {
			log.Print(id, "====", item.CaseTypeID)
			// ⭐ IF system uses caseSTypeId → change this line
			if id == item.CaseTypeID {
				groupTypeId = g.GroupTypeId
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	log.Print("--CalDashboardSLA--4")
	if !found {
		return fmt.Errorf("groupTypeId not found for caseTypeId: %s", item.CaseTypeID)
	}

	// --------------------
	// ⭐ SLA calculation
	// --------------------
	if item.CreatedDate == nil || item.CaseSLA == nil {
		return fmt.Errorf("case missing CreatedDate or CaseSLA")
	}

	// fmt.Println("---- CASE FIELDS ----")
	// fmt.Println("CaseID:", item.CaseID)
	// fmt.Println("StatusID:", item.StatusID)
	// fmt.Println("CreatedDate:", item.CreatedDate)
	// fmt.Println("CaseSLA:", item.CaseSLA)
	// fmt.Println("ResponderCreatedAt:", item.CreatedAt)
	// fmt.Println("CountryID:", item.CountryID)
	// fmt.Println("ProvID:", item.ProvID)
	// fmt.Println("DistID:", item.DistID)
	// fmt.Println("CaseTypeID:", item.CaseTypeID)
	// fmt.Println("----------------------")

	created := *item.CreatedDate
	slaMinutes := *item.CaseSLA

	closeDuration := item.CreatedAt.Sub(created)
	slaDuration := time.Duration(slaMinutes) * time.Minute

	inSla := closeDuration <= slaDuration

	// convert duration → seconds (INT)
	caseDuration := int(closeDuration.Seconds())

	if !inSla {
		caseDuration = 0
	}
	log.Print("--CalDashboardSLA--5")
	log.Print(inSla)
	// --------------------
	// UPSERT
	// --------------------
	var query string
	if inSla {
		query = `
        INSERT INTO d_case_summary 
            ("orgId", date, time, "groupTypeId", "countryId", "provId", "distId", "inSla", "caseDuration")
        VALUES ($1, $2, $3, $4, $5, $6, $7, '1')
        ON CONFLICT ("orgId", date, time, "groupTypeId", "countryId", "provId", "distId")
        DO UPDATE SET "inSla" = (d_case_summary."inSla"::int + 1)::varchar,
			"caseDuration" = (d_case_summary."caseDuration"::int + $8)::varchar;
        `
	} else {
		caseDuration = 0
		query = `
        INSERT INTO d_case_summary 
            ("orgId", date, time, "groupTypeId", "countryId", "provId", "distId", "overSla", "caseDuration")
        VALUES ($1, $2, $3, $4, $5, $6, $7, '1')
        ON CONFLICT ("orgId", date, time, "groupTypeId", "countryId", "provId", "distId")
        DO UPDATE SET "overSla" = (d_case_summary."overSla"::int + 1)::varchar,
			"caseDuration" = (d_case_summary."caseDuration"::int + $8)::varchar;
        `
	}
	log.Print(query)
	fmt.Println("---- CASE FIELDS ----")
	fmt.Println("orgId:", orgId)
	fmt.Println("currentDate:", currentDate)
	fmt.Println("currentTime:", currentTime)
	fmt.Println("groupTypeId:", groupTypeId)
	fmt.Println("CountryID:", item.CountryID)
	fmt.Println("ProvID:", item.ProvID)
	fmt.Println("DistID:", item.DistID)
	fmt.Println("caseDuration:", caseDuration)
	fmt.Println("----------------------")
	_, err = conn.Exec(
		ctx,
		query,
		orgId,
		currentDate,
		currentTime,
		groupTypeId,
		item.CountryID, item.ProvID, item.DistID,
		caseDuration,
	)

	if err != nil {
		return fmt.Errorf("failed to upsert summary: %v", err)
	}

	// recipients := []model.Recipient{
	// 	{Type: "provId", Value: item.ProvID},
	// }

	return nil
}

func GetCaseClose(
	c context.Context,
	conn *pgx.Conn,
	orgId string,
	caseId string,
) (*model.Case, error) {

	const queryGetCaseClose = `
SELECT 
    c."caseId", 
    c."statusId",
    c."createdDate",
    c."caseSla", 
    r."createdAt",
    c."countryId",
    c."provId",
    c."distId",
    c."caseTypeId"
FROM tix_cases c
LEFT JOIN tix_case_responders r 
       ON r."caseId" = c."caseId"
WHERE c."statusId" = 'S007'
  AND r."statusId" = 'S007'
  AND c."orgId" = $1
  AND c."caseId" = $2
LIMIT 1
`

	row := conn.QueryRow(c, queryGetCaseClose, orgId, caseId)

	var item model.Case

	err := row.Scan(
		&item.CaseID,
		&item.StatusID,
		&item.CreatedDate, // *time.Time
		&item.CaseSLA,     // *int
		&item.CreatedAt,   // time.Time
		&item.CountryID,
		&item.ProvID,
		&item.DistID,
		&item.CaseTypeID,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("scan GetCaseClose failed: %w", err)
	}

	return &item, nil
}
