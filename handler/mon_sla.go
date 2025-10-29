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
				//fmt.Println("Stage Data JSON:", stage)
				caseId := stage.CaseId
				req := model.UpdateStageRequest{
					CaseId:   caseId,
					Status:   RecheckSLA(stage.StatusId),
					UnitUser: MONITOR_NAME, // หรือ set ค่า default
				}
				//log.Print(req)
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
		SELECT c."caseId", c."statusId", s."data", s."updatedAt", 
		       c."versions", c."overSlaCount", s."wfId", s."nodeId"
		FROM tix_cases c
		JOIN tix_case_current_stage s ON c."caseId" = s."caseId"
		WHERE s."stageType" = 'case'
		AND c."orgId" = '%s'
		AND c."statusId" IN (%s)
		AND c."overSlaCount" < %s
		AND (
          c."overSlaDate" IS NULL
          OR c."overSlaDate" < NOW() - INTERVAL '%d minute'
      	)
		AND s."updatedAt" IS NOT NULL;
	`, orgId, statusStr_, maxAlert, alertDur)

	//For test
	// query = fmt.Sprintf(`
	// 	SELECT c."caseId", c."statusId", s."data", s."updatedAt",
	// 	       c."versions", c."overSlaCount", s."wfId", s."nodeId"
	// 	FROM tix_cases c
	// 	JOIN tix_case_current_stage s ON c."caseId" = s."caseId"
	// 	WHERE s."stageType" = 'case'
	// 	AND c."orgId" = '%s'
	// 	AND c."caseId" = 'D251020-00006';
	// `, orgId)

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	var results, results_data []model.CaseStageInfo
	wfSet := make(map[string]struct{}) // collect unique wfIds

	for rows.Next() {
		var rec model.CaseStageInfo
		if err := rows.Scan(&rec.CaseId, &rec.StatusId, &rec.Data, &rec.UpdatedAt,
			&rec.Versions, &rec.OverSlaCount, &rec.WfId, &rec.NodeId); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		if rec.UpdatedAt == nil {
			fmt.Printf("Case %s has NULL UpdatedAt, skipping\n", rec.CaseId)
			continue
		}
		wfSet[rec.WfId] = struct{}{}
		results = append(results, rec)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	// 🔹 Step 2: Fetch workflow nodes for all wfIds
	wfNodesMap, err := getWorkflow(ctx, conn, orgId, wfSet)
	if err != nil {
		return nil, fmt.Errorf("getWorkflow error: %w", err)
	}

	// 🔹 Step 3: Compute nextNode for each case
	for i, c := range results {
		nodes := wfNodesMap[c.WfId]
		if nodes == nil {
			continue
		}

		var nextNode *model.WorkflowNode

		// find "connections" node
		for _, node := range nodes {
			if node.Section != "connections" {
				continue
			}

			dataBytes, _ := json.Marshal(node.Data)
			var conns []model.WorkFlowConnection
			if err := json.Unmarshal(dataBytes, &conns); err != nil {
				continue
			}

			for _, conn := range conns {
				if conn.Source == c.NodeId {
					candidate := nodes[conn.Target]
					// skip SLA nodes
					for candidate.Type == "sla" {
						found := false
						for _, c2 := range conns {
							if c2.Source == candidate.NodeId && c2.Label == "yes" {
								candidate = nodes[c2.Target]
								found = true
								break
							}
						}
						if !found {
							break
						}
					}
					nextNode = &candidate
					break
				}
			}
			if nextNode != nil {
				break
			}
		}

		// Calculate SLA
		var nodeData model.WorkflowNodeData
		dataBytes, _ := json.Marshal(nextNode.Data)
		if err := json.Unmarshal(dataBytes, &nodeData); err != nil {
			fmt.Printf("❌ Cannot parse nextNode.Data for case %s: %v\n", results[i].CaseId, err)
			continue
		}

		slaMin, err := strconv.Atoi(nodeData.Data.Config.SLA)
		if err != nil {
			fmt.Printf("Invalid SLA for case %s: %s\n", results[i].CaseId, nodeData.Data.Config.SLA)
			continue
		}
		// --- FIX START ---

		now := getTimeNowUTC()                  // always compare in UTC
		updatedAt := results[i].UpdatedAt.UTC() // normalize both
		expireTime := updatedAt.Add(time.Duration(slaMin) * time.Minute)
		// --- FIX END ---

		// fmt.Printf("📋 UpdatedAt : %s\n", updatedAt)
		// fmt.Printf("📋 now : %s\n", now)
		// fmt.Printf("📋 slaMin : %d\n", slaMin)
		// fmt.Printf("📋 expireTime : %s\n", expireTime)

		//fmt.Printf("📋 now : %s\n", now)
		// fmt.Printf("📋 now getTimeNowUTC : %s\n", now)
		//fmt.Printf("📋 now getTimeNow : %s\n", getTimeNow())
		//fmt.Printf("📋 now getTimeNowUTC : %s\n", getTimeNowUTC())
		// fmt.Printf("📋 now getTimeNowBangkok : %s\n", getTimeNowBangkok())
		// fmt.Printf("📋 now displayBangkokTime : %s\n", displayBangkokTime(getTimeNow()))

		// fmt.Printf("📋 expireTime : %s\n", expireTime)
		//fmt.Printf("📋 expireTime Results: %v\n", now.After(expireTime))

		//x := created
		//y := getTimeNowUTC()
		//z := expireTime
		//fmt.Printf("📋 Check  \n%s\n%s\n%s\n", x, y, z)

		//fmt.Printf("📋 expireTime Results: %v\n", now.After(expireTime))
		if now.After(expireTime) {
			fmt.Printf("📋 UpdatedAt: %s | now: %s | slaMin: %d | expireTime: %s\n", updatedAt, now, slaMin, expireTime)
			results[i].NextNode = nextNode
			results_data = append(results_data, results[i])
		}

		//results[i].NextNode = nextNode
	}

	jsonBytes, err := json.MarshalIndent(results_data, "", "  ")
	if err != nil {
		log.Printf("❌ JSON marshal error: %v", err)
	} else {
		fmt.Printf("📋 CaseStageData Results:\n%d\n", len(jsonBytes))
	}

	return results_data, nil
}

func getWorkflow(ctx context.Context, conn *pgx.Conn, orgId string, wfSet map[string]struct{}) (map[string]map[string]model.WorkflowNode, error) {
	// Collect wfIds from wfSet
	var wfIds []string
	for id := range wfSet {
		wfIds = append(wfIds, id)
	}

	// If no wfIds, return empty map
	if len(wfIds) == 0 {
		return map[string]map[string]model.WorkflowNode{}, nil
	}

	query := `
		SELECT "wfId", "nodeId", "type", "section", "data"
		FROM wf_nodes
		WHERE "orgId" = $1 AND "wfId" = ANY($2)
		ORDER BY "wfId", "nodeId";
	`

	rows, err := conn.Query(ctx, query, orgId, wfIds)
	if err != nil {
		return nil, fmt.Errorf("fetch workflow nodes error: %w", err)
	}
	defer rows.Close()

	wfNodesMap := make(map[string]map[string]model.WorkflowNode)

	for rows.Next() {
		var wfId, nodeId, nodeType, section string
		var data any

		if err := rows.Scan(&wfId, &nodeId, &nodeType, &section, &data); err != nil {
			return nil, fmt.Errorf("row scan error: %w", err)
		}

		if _, ok := wfNodesMap[wfId]; !ok {
			wfNodesMap[wfId] = make(map[string]model.WorkflowNode)
		}

		wfNodesMap[wfId][nodeId] = model.WorkflowNode{
			NodeId:  nodeId,
			Type:    nodeType,
			Section: section,
			Data:    data,
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return wfNodesMap, nil
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
