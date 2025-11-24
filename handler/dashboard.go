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

func CalDashboardCaseSummary(
	ctx context.Context,
	conn *pgx.Conn,
	orgId string,
	recipients []model.Recipient,
	username, caseTypeId, countryId, provId, distId string,
) error {

	// ใช้เวลาไทย
	now := getTimeNow()

	currentDate := now.Format("2006-01-02") // DATE format "YYYY-MM-DD"
	currentHour := now.Hour()
	currentTime := fmt.Sprintf("%02d:00:00", currentHour) // e.g., "16:00:00"

	year, month, day := now.Year(), int(now.Month()), now.Day()

	// โหลดข้อมูล groupType
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

	// UPSERT พร้อม year, month, day, date
	query := `
        INSERT INTO d_case_summary 
    ("orgId", date, time, "groupTypeId", "countryId", "provId", "distId", year, month, day, total, "new")
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 1, 1)
ON CONFLICT ("orgId", date, time, "groupTypeId", "countryId", "provId", "distId")
DO UPDATE SET total = d_case_summary.total + 1,
              "new" = d_case_summary.new + 1;
    `

	_, err = conn.Exec(ctx, query,
		orgId,
		currentDate, // DATE
		currentTime, // TIME
		groupTypeId,
		countryId,
		provId,
		distId,
		year,  // INTEGER
		month, // INTEGER
		day,   // INTEGER
	)

	if err != nil {
		return fmt.Errorf("failed to upsert summary: %v", err)
	}

	// ส่งแดชบอร์ด
	err = CoreDashboard(ctx, conn, orgId, username, recipients, true, false, true)
	if err != nil {
		log.Printf("Dashboard notification error: %v", err)
	}

	log.Printf("✅ Upsert summary success for groupTypeId=%s, countryId=%s, provId=%s, distId=%s",
		groupTypeId, countryId, provId, distId)
	return nil
}

func SendDashboardSummary_Socket(
	c context.Context,
	conn *pgx.Conn,
	orgId string,
	username string,
	recipients []model.Recipient,
) error {

	// 1) Load user info + distId list
	userInfo, err := utils.GetAreaByUsernameOrLoad(c, conn, orgId, username)
	if err != nil {
		return fmt.Errorf("GetAreaByUsernameOrLoad failed: %w", err)
	}

	distList := userInfo.DistIdLists // []string

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
		  AND s.date = CURRENT_DATE
		  AND "distId" = ANY($2)
		GROUP BY s."groupTypeId"
	`
	rows, err := conn.Query(c, query, orgId, distList)
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

	// โหลด groupType
	groupTypes, err := utils.GroupTypeGetOrLoad(conn)
	if err != nil {
		return fmt.Errorf("cannot load group types: %v", err)
	}
	log.Print("--CalDashboardSLA--2")

	// โหลด case data
	item, err := GetCaseClose(ctx, conn, orgId, caseId)
	if err != nil {
		return err
	}
	if item == nil {
		return fmt.Errorf("case not found")
	}
	log.Print("--CalDashboardSLA--3")

	// match groupTypeId
	groupTypeId := ""
	found := false
	for _, g := range groupTypes {
		for _, id := range g.GroupTypeLists {
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
	if !found {
		return fmt.Errorf("groupTypeId not found for caseTypeId: %s", item.CaseTypeID)
	}

	// --------------------
	// SLA calculation
	// --------------------
	if item.CreatedDate == nil || item.CaseSLA == nil {
		return fmt.Errorf("case missing CreatedDate or CaseSLA")
	}
	created := *item.CreatedDate
	slaMinutes := *item.CaseSLA
	closeDuration := item.CreatedAt.Sub(created)
	slaDuration := time.Duration(slaMinutes) * time.Minute
	inSla := closeDuration <= slaDuration
	caseDuration := int(closeDuration.Seconds())

	// --------------------
	// Date parts
	// --------------------
	// currentDate เป็น DATE type (เวลาเป็น 00:00:00)
	currentDate := time.Date(created.Year(), created.Month(), created.Day(), 0, 0, 0, 0, created.Location())
	// currentTime เป็น string HH:00:00
	currentTime := fmt.Sprintf("%02d:00:00", created.Hour())
	year, month, day := created.Year(), int(created.Month()), created.Day()

	// --------------------
	// UPSERT
	// --------------------
	var query string
	if inSla {
		query = `
		INSERT INTO d_case_summary
			("orgId", date, time, "groupTypeId", "countryId", "provId", "distId",
			 year, month, day, "inSla", "caseDuration", inprogress, complete)
		VALUES ($1, $2, $3, $4, $5, $6, $7,
				$8, $9, $10, 1, $11, -1, 1)
		ON CONFLICT ("orgId", date, time, "groupTypeId", "countryId", "provId", "distId")
		DO UPDATE SET
			"inSla" = d_case_summary."inSla" + 1,
			"caseDuration" = d_case_summary."caseDuration" + EXCLUDED."caseDuration",
			inprogress = d_case_summary.inprogress - 1,
			complete = d_case_summary.complete + 1;
		`
	} else {
		query = `
		INSERT INTO d_case_summary
			("orgId", date, time, "groupTypeId", "countryId", "provId", "distId",
			 year, month, day, "overSla", "caseDuration", inprogress, complete)
		VALUES ($1, $2, $3, $4, $5, $6, $7,
				$8, $9, $10, 1, $11, -1, 1)
		ON CONFLICT ("orgId", date, time, "groupTypeId", "countryId", "provId", "distId")
		DO UPDATE SET
			"overSla" = d_case_summary."overSla" + 1,
			"caseDuration" = d_case_summary."caseDuration" + EXCLUDED."caseDuration",
			inprogress = d_case_summary.inprogress - 1,
			complete = d_case_summary.complete + 1;
		`
	}

	// --------------------
	// Debug values
	// --------------------
	fmt.Println("=== DEBUG VALUES ===")
	fmt.Println("orgId      :", orgId)
	fmt.Println("currentDate:", currentDate)
	fmt.Println("currentTime:", currentTime)
	fmt.Println("groupTypeId:", groupTypeId)
	fmt.Println("CountryID  :", item.CountryID)
	fmt.Println("ProvID     :", item.ProvID)
	fmt.Println("DistID     :", item.DistID)
	fmt.Println("year       :", year)
	fmt.Println("month      :", month)
	fmt.Println("day        :", day)
	fmt.Println("caseDuration:", caseDuration)
	fmt.Println("====================")
	fmt.Println("UPSERT SQL:", query)

	_, err = conn.Exec(
		ctx,
		query,
		orgId,
		currentDate,
		currentTime,
		groupTypeId,
		item.CountryID, item.ProvID, item.DistID,
		year, month, day,
		caseDuration,
	)
	log.Print(err)
	if err != nil {
		return fmt.Errorf("failed to upsert summary: %v", err)
	}

	// ส่งแดชบอร์ด
	recipients := []model.Recipient{
		{Type: "provId", Value: item.ProvID},
	}
	err = CoreDashboard(ctx, conn, orgId, username, recipients, false, true, true)
	if err != nil {
		log.Printf("Dashboard notification error: %v", err)
	}

	log.Printf("✅ Upsert SLA success for groupTypeId=%s, countryId=%s, provId=%s, distId=%s",
		groupTypeId, item.CountryID, item.ProvID, item.DistID)
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

func SendDashboardSLA_Socket(
	c context.Context,
	conn *pgx.Conn,
	orgId string,
	username string,
	recipients []model.Recipient,
) error {

	// 1) Load user info + distId list
	userInfo, err := utils.GetAreaByUsernameOrLoad(c, conn, orgId, username)
	if err != nil {
		return fmt.Errorf("GetAreaByUsernameOrLoad failed: %w", err)
	}

	distList := userInfo.DistIdLists // []string

	// 2) Query SLA summary filtered by distId
	query := `
		SELECT 
			COALESCE(SUM(CAST("inSla" AS INT)), 0) AS "inSLA",
			COALESCE(SUM(CAST("overSla" AS INT)), 0) AS "overSLA",
			COALESCE(SUM(CAST("caseDuration" AS INT)), 0) AS "totalDuration"
		FROM d_case_summary
		WHERE "orgId" = $1
		  AND date = CURRENT_DATE 
		  AND "distId" = ANY($2)
	`

	var inSLA, overSLA, totalDuration int

	err = conn.QueryRow(c, query, orgId, distList).Scan(&inSLA, &overSLA, &totalDuration)
	if err != nil {
		return fmt.Errorf("query SLA summary failed: %w", err)
	}

	total := inSLA + overSLA

	// 3) Calculate %
	percentage := "0%"
	if total > 0 {
		percentage = fmt.Sprintf("%d%%", (inSLA*100)/total)
	}

	avgResponse := 0
	if total > 0 {
		avgResponse = (totalDuration / 60) / total
	}

	// 4) Prepare JSON
	data := []interface{}{
		map[string]interface{}{"total_en": "Total", "total_th": "ทั้งหมด", "val": total},
		map[string]interface{}{"inSLA_en": "InSLA", "inSLA_th": "ปฏิบัติตาม SLA", "val": inSLA},
		map[string]interface{}{"overSLA_en": "OverSLA", "overSLA_th": "เกินกำหนด SLA", "val": overSLA},
		map[string]interface{}{
			"percentage_inSLA_en": "On Time",
			"percentage_inSLA_th": "เสร็จทันเวลา",
			"val":                 percentage,
		},
		map[string]interface{}{
			"avg_respose_time_en": "Average Response Time",
			"avg_respose_time_th": "เวลาเฉลี่ยในการแก้ปัญหา",
			"val":                 avgResponse,
			"unit":                "min",
		},
	}

	payload := map[string]interface{}{
		"type":     "SLA-PERFORMANCE",
		"title_en": "SLA Performance",
		"title_th": "ประสิทธิภาพการทำงาน",
		"data":     data,
	}

	jsonBytes, _ := json.Marshal(payload)
	raw := json.RawMessage(jsonBytes)

	// 5) Send notification
	err = genNotiCustom(
		c, conn,
		orgId,
		username, username, "",
		"hidden",
		nil,
		"",
		recipients,
		"",
		"User",
		"DASHBOARD",
		&raw,
	)

	if err != nil {
		return fmt.Errorf("send dashboard SLA notification failed: %w", err)
	}

	log.Printf("✅ SLA dashboard sent successfully: total=%d in=%d over=%d", total, inSLA, overSLA)
	return nil
}

func CalDashboardStatus(ctx context.Context,
	conn *pgx.Conn,
	orgId string,
	username, caseId string,
	status string,
) error {

	log.Print("===CalDashboardStatus===")

	item, err := GetCaseByID(ctx, conn, orgId, caseId)
	if err != nil {
		log.Print(model.Response{Status: "-1", Msg: "Failure.CalDashboardStatus.0" + caseId, Desc: err.Error()})
		return err
	}
	recipients := []model.Recipient{
		{Type: "provId", Value: item.ProvID},
	}

	// load groupType
	groupTypes, err := utils.GroupTypeGetOrLoad(conn)
	if err != nil {
		return fmt.Errorf("cannot load group types: %v", err)
	}

	// match groupTypeId
	groupTypeId := ""
	found := false
	for _, g := range groupTypes {
		for _, id := range g.GroupTypeLists {
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
	if !found {
		return fmt.Errorf("groupTypeId not found for caseTypeId: %s", item.CaseTypeID)
	}

	if item.CreatedDate == nil {
		return fmt.Errorf("case missing CreatedDate")
	}

	created := *item.CreatedDate
	currentDate := created.Format("2006-01-02") // DATE
	currentTime := fmt.Sprintf("%02d:00:00", created.Hour())
	year, month, day := created.Year(), int(created.Month()), created.Day()

	// --------------------
	// UPSERT
	// --------------------
	var query string
	if status == "cancel" {
		query = `
		INSERT INTO d_case_summary
			("orgId", date, time, "groupTypeId", "countryId", "provId", "distId",
			 year, month, day, new, inprogress)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,1,1)
		ON CONFLICT ("orgId", date, time, "groupTypeId", "countryId", "provId", "distId")
		DO UPDATE SET new = d_case_summary.new + 1,
		              inprogress = d_case_summary.inprogress - 1;
		`
	} else {
		query = `
		INSERT INTO d_case_summary
			("orgId", date, time, "groupTypeId", "countryId", "provId", "distId",
			 year, month, day, new, inprogress)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,1,1)
		ON CONFLICT ("orgId", date, time, "groupTypeId", "countryId", "provId", "distId")
		DO UPDATE SET new = d_case_summary.new - 1,
		              inprogress = d_case_summary.inprogress + 1;
		`
	}

	_, err = conn.Exec(ctx, query,
		orgId,
		currentDate,
		currentTime,
		groupTypeId,
		item.CountryID,
		item.ProvID,
		item.DistID,
		year,
		month,
		day,
	)

	if err != nil {
		return fmt.Errorf("failed to upsert summary: %v", err)
	}

	// ส่งแดชบอร์ด
	err = CoreDashboard(ctx, conn, orgId, username, recipients, false, false, true)
	if err != nil {
		log.Printf("Dashboard notification error: %v", err)
	}

	log.Printf("✅ Upsert Dashboard for groupTypeId=%s, countryId=%s, provId=%s, distId=%s",
		groupTypeId, item.CountryID, item.ProvID, item.DistID)
	return nil
}

func SendDashboardStatus_Socket(
	c context.Context,
	conn *pgx.Conn,
	orgId string,
	username string,
	recipients []model.Recipient,
) error {
	// 1) Load districts of user
	userInfo, err := utils.GetAreaByUsernameOrLoad(c, conn, orgId, username)
	if err != nil {
		return fmt.Errorf("GetAreaByUsernameOrLoad failed: %w", err)
	}

	distList := userInfo.DistIdLists // []string

	// 2) Calculate last 12 months
	now := time.Now()
	months := make([]struct{ Year, Month int }, 12)
	for i := 0; i < 12; i++ {
		t := now.AddDate(0, -i, 0)
		months[i] = struct{ Year, Month int }{t.Year(), int(t.Month())}
	}

	// 3) Query monthly summary for last 12 months
	query := `
	SELECT year, month,
	       SUM("new") AS new,
	       SUM("inprogress") AS inprogress,
	       SUM("complete") AS complete
	FROM d_case_summary
	WHERE "orgId" = $1 
	  AND "distId" = ANY($2)
	  AND (year, month) IN (` +
		func() string {
			s := ""
			for i, m := range months {
				if i > 0 {
					s += ","
				}
				s += fmt.Sprintf("(%d,%d)", m.Year, m.Month)
			}
			return s
		}() + `)
	GROUP BY year, month
	ORDER BY year, month
	`

	rows, err := conn.Query(c, query, orgId, distList)
	if err != nil {
		return fmt.Errorf("query last 12 months summary failed: %w", err)
	}
	defer rows.Close()

	// map[year-month] → counts
	dataMap := make(map[string]map[string]int)
	for rows.Next() {
		var y, m, n, ip, c int
		if err := rows.Scan(&y, &m, &n, &ip, &c); err != nil {
			return fmt.Errorf("scan row failed: %w", err)
		}
		key := fmt.Sprintf("%d-%02d", y, m)
		dataMap[key] = map[string]int{"new": n, "inprogress": ip, "complete": c}
	}

	// Prepare JSON
	payloadData := []interface{}{}
	thMonths := []string{"ม.ค.", "ก.พ.", "มี.ค.", "เม.ย.", "พ.ค.", "มิ.ย.", "ก.ค.", "ส.ค.", "ก.ย.", "ต.ค.", "พ.ย.", "ธ.ค."}
	enMonths := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

	// Total summary
	totalNew, totalIP, totalC := 0, 0, 0
	for _, m := range months {
		key := fmt.Sprintf("%d-%02d", m.Year, m.Month)
		d := dataMap[key]
		totalNew += d["new"]
		totalIP += d["inprogress"]
		totalC += d["complete"]
	}
	payloadData = append(payloadData, map[string]interface{}{
		"total_en":   "Total",
		"total_th":   "ทั้งหมด",
		"new":        totalNew,
		"inprogress": totalIP,
		"complete":   totalC,
	})

	// Each month from oldest → newest
	for i := len(months) - 1; i >= 0; i-- {
		m := months[i]
		key := fmt.Sprintf("%d-%02d", m.Year, m.Month)
		d := dataMap[key]
		payloadData = append(payloadData, map[string]interface{}{
			fmt.Sprintf("m%d_en", m.Month): fmt.Sprintf("%s %d", enMonths[m.Month-1], m.Year),
			fmt.Sprintf("m%d_th", m.Month): fmt.Sprintf("%s %d", thMonths[m.Month-1], m.Year+543),
			"new":                          d["new"],
			"inprogress":                   d["inprogress"],
			"complete":                     d["complete"],
		})
	}

	payload := map[string]interface{}{
		"type":     "CASE-MONTHLY-SUMMARY",
		"title_en": "Work Order in Monthly Summary",
		"title_th": "สรุปคำสั่งงานประจำเดือน",
		"data":     payloadData,
	}

	jsonBytes, _ := json.Marshal(payload)
	raw := json.RawMessage(jsonBytes)

	err = genNotiCustom(
		c, conn,
		orgId,
		username, username, "",
		"hidden",
		nil,
		"",
		recipients,
		"",
		"User",
		"DASHBOARD",
		&raw,
	)
	if err != nil {
		return fmt.Errorf("send last 12 months dashboard notification failed: %w", err)
	}

	log.Printf("✅ Last 12 months dashboard sent successfully")
	return nil
}

func CoreDashboard(
	c context.Context,
	conn *pgx.Conn,
	orgId string,
	username string,
	recipients []model.Recipient,
	summary bool,
	sla bool,
	status bool,
) error {
	if summary {
		log.Print("==SendDashboardSummary_Socket=")
		err := SendDashboardSummary_Socket(c, conn, orgId, username, recipients)
		if err != nil {
			log.Printf("Dashboard notification error: %v", err)
		}
	}
	if sla {
		log.Print("==SendDashboardSLA_Socket=")
		err := SendDashboardSLA_Socket(c, conn, orgId, username, recipients)
		if err != nil {
			log.Printf("Dashboard notification error: %v", err)
		}
	}
	if status {
		log.Print("==SendDashboardStatus_Socket=")
		err := SendDashboardStatus_Socket(c, conn, orgId, username, recipients)
		if err != nil {
			log.Printf("Dashboard notification error: %v", err)
		}
	}

	return nil
}
