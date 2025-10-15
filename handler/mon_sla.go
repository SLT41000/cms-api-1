package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mainPackage/model"
	"mainPackage/utils"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func SlaMonitor(c *gin.Context) error {
	counter := 0
	MONITOR_NAME := os.Getenv("MONITOR_NAME")
	orgId := os.Getenv("INTEGRATION_ORG_ID")

	hostname, err := os.Hostname()
	if err != nil {
		log.Println("Hostname error:", err)
		hostname = "unknown"
	}

	log.Printf("Starting SLA monitor on host: %s\n", hostname)

	for {
		val, err := utils.OwnerSLAGet()
		if err != nil {
			log.Println("Redis GET error:", err)
			return err
		}

		err = utils.OwnerSLASet(hostname)
		if err != nil {
			log.Println("Redis SET error:", err)
			return err
		}

		log.Printf("[Tick %d] Redis sla key = '%s'\n", counter, val)
		val = ""
		// Only this host should write if no one else owns the lock or it owns it
		if val == "" || val == hostname {

			conn, ctx, cancel := utils.ConnectDB()
			if conn == nil {
				log.Printf("DB connection is nil")
				return nil
			}
			defer cancel()
			defer conn.Close(ctx)

			stages, err := GetCaseStageData(ctx, conn, orgId)
			if err != nil {
				fmt.Println("Error:", err)
				return nil
			}

			//Set Alert By caseId
			c.Set("username", MONITOR_NAME)
			c.Set("orgId", orgId)

			for _, stage := range stages {
				fmt.Println("Stage Data JSON:", stage)
				caseId := stage.CaseId
				req := model.UpdateStageRequest{
					CaseId:   caseId,
					Status:   RecheckSLA(stage.StatusId),
					UnitUser: MONITOR_NAME, // หรือ set ค่า default
				}
				log.Print(req)
				delay, err := strconv.Atoi(stage.OverSlaCount)
				if err != nil {
					log.Printf("Invalid OverSlaCount=%s, defaulting to 0", stage.OverSlaCount)
					delay = 1
				}
				delay++
				if delay > 2 {
					delay = 2
				}
				GenerateNotiAndComment(c, conn, req, orgId, strconv.Itoa(delay))
				err = UpdateCaseSLAPlus(ctx, conn, orgId, caseId, true, time.Now())
				if err != nil {
					log.Printf("Failed to update SLA for case %s: %v", caseId, err)
				}
			}

			err = utils.OwnerSLASet("")
			if err != nil {
				log.Println("Redis SET error:", err)
				return err
			}
			log.Printf("SLA lock acquired/renewed by host: %s\n", hostname)
		} else {
			log.Printf("SLA lock held by another host: %s (this host: %s)\n", val, hostname)
		}

		counter++

		intervalStr := os.Getenv("MONITOR_SLA_INTERVAL")
		interval, err := strconv.Atoi(intervalStr)
		if err != nil {
			log.Printf("Invalid MONITOR_SLA_INTERVAL=%s, fallback to 1 min", intervalStr)
			interval = 1
		}

		sleep := time.Duration(interval) * time.Minute
		log.Print("Sleep : ", sleep)
		time.Sleep(sleep)
	}
}

func GetCaseStageData(ctx context.Context, conn *pgx.Conn, orgId string) ([]model.CaseStageInfo, error) {
	maxAlert := os.Getenv("MONITOR_SLA_ALERT_LIMIT")
	alertDurStr := os.Getenv("MONITOR_SLA_NEXT_ALERT")
	alertDur, err := strconv.Atoi(alertDurStr)
	if err != nil {
		log.Printf("Invalid MONITOR_SLA_NEXT_ALERT=%s, fallback 10", alertDurStr)
		alertDur = 10
	}

	statusStr := os.Getenv("MONITOR_SLA")
	statusStr_ := ConvertStatusList(statusStr)
	query := fmt.Sprintf(`
		SELECT c."caseId" ,c."statusId" , s.data , c."createdDate", c.versions, c."overSlaCount"
		FROM tix_cases c
		JOIN tix_case_current_stage s 
		  ON c."caseId" = s."caseId"
		WHERE s."stageType" = 'case'
		AND c."orgId" = '%s'
		AND c."statusId" IN (%s)
		AND c."overSlaCount" < %s
		AND (
          c."overSlaDate" IS NULL
          OR c."overSlaDate" < NOW() - INTERVAL '%d minute'
      	)
		AND c."createdDate" IS NOT NULL;
	`, orgId, statusStr_, maxAlert, alertDur)

	//log.Print(query)

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var results []model.CaseStageInfo

	for rows.Next() {
		var rec model.CaseStageInfo
		if err := rows.Scan(&rec.CaseId, &rec.StatusId, &rec.Data, &rec.CreatedDate, &rec.Versions, &rec.OverSlaCount); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		// Skip NULL createdDate
		if rec.CreatedDate == nil {
			fmt.Printf("Case %s has NULL createdDate, skipping\n", rec.CaseId)
			continue
		}

		// Calculate SLA
		var stage model.NodeData
		if err := json.Unmarshal([]byte(rec.Data), &stage); err != nil {
			fmt.Println("JSON parse error:", err)
			continue
		}

		slaMin, err := strconv.Atoi(stage.Data.Config.SLA)
		if err != nil {
			fmt.Printf("Invalid SLA for case %s: %s\n", rec.CaseId, stage.Data.Config.SLA)
			continue
		}

		expireTime := rec.CreatedDate.Add(time.Duration(slaMin) * time.Minute)

		if time.Now().After(expireTime) {
			fmt.Printf("Case %s SLA expired\n", rec.CaseId)
			results = append(results, rec) // ✅ เก็บทั้ง record
		}
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return results, nil
}

func UpdateCaseSLAPlus(ctx context.Context, conn *pgx.Conn, orgId, caseId string, overSlaFlag bool, overSlaDate time.Time) error {
	query := `
		UPDATE tix_cases
		SET "overSlaFlag" = $1,
		    "overSlaDate" = $2,
		    "overSlaCount" = COALESCE("overSlaCount", 0) + 1,
		    "updatedAt" = NOW()
		WHERE "orgId" = $3
		  AND "caseId" = $4;
	`
	//log.Print("====UpdateCaseSLAPlus===", query)
	_, err := conn.Exec(ctx, query, overSlaFlag, overSlaDate, orgId, caseId)
	if err != nil {
		return fmt.Errorf("update case SLA failed: %w", err)
	}

	return nil
}
