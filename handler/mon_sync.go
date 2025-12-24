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
)

func ReSyncCase(c *gin.Context) error {
	username := os.Getenv("INTEGRATION_USR")
	orgId := os.Getenv("INTEGRATION_ORG_ID")

	c.Set("username", username)
	c.Set("orgId", orgId)

	counter := 0

	hostname, err := os.Hostname()
	if err != nil {
		log.Println("Hostname error:", err)
		hostname = "unknown"
	}

	maxRetryStr := os.Getenv("ESB_RE_CREATE_MAX")
	maxRetry, err := strconv.Atoi(maxRetryStr)
	if err != nil {
		log.Printf("Invalid ESB_RE_CREATE_MAX=%s, fallback to 3", maxRetryStr)
		maxRetry = 3 // ค่า default
	}

	log.Printf("Starting Case Sync monitor on host: %s\n", hostname)

	for {
		val, err := utils.OwnerCaseSyncGet()
		if err != nil {
			log.Println("Redis GET error:", err)
			return err
		}

		log.Printf("[Tick %d] Redis sla key = '%s'\n", counter, val)
		val = ""
		// Only this host should write if no one else owns the lock or it owns it
		if val == "" || val == hostname {

			err = utils.OwnerCaseSyncSet(hostname)
			if err != nil {
				log.Println("Redis SET error:", err)
				return err
			}

			items, err := utils.GetAllCaseSync(context.Background())
			if err != nil {
				log.Println(err)

			}

			for _, item := range items {
				var req model.OwnerCaseSyncReq
				if err := json.Unmarshal([]byte(item.Value), &req); err != nil {
					log.Println("unmarshal error:", err)
					continue
				}

				//Update ESB WO
				conn, ctx, cancel := utils.ConnectDB()
				if conn == nil {
					continue
				}
				defer cancel()
				defer conn.Close(ctx)
				defer cancel()

				UnitUser := ""
				unitLists, count, err := GetUnitsWithDispatch(c, conn, orgId, req.CaseId, "S003", "")
				if err != nil {
					panic(err)
				}
				if count > 0 {
					UnitUser = unitLists[0].Username
					fmt.Println("First username:", UnitUser)
				}
				req_ := model.UpdateStageRequest{
					CaseId:   req.CaseId,
					Status:   "S001",
					UnitUser: UnitUser, // หรือ set ค่า default
				}
				log.Print("==xxx==")
				log.Print(req_)
				log.Print(c)

				flagError := true
				if req.Type == "create" {
					res, error := Re_CreateBusKafka_WO(c, conn, req_, orgId)

					if error != nil {
						log.Print("--error--")
						log.Print(error)
						msg := "Re_CreateBusKafka_WO - Error"
						req.Message = &msg
						//count + 1
					} else {
						hasWorkOrderNumber := res.Data != nil && res.Data.WorkOrderNumber != ""

						if hasWorkOrderNumber {
							log.Println("Work order created:", res.Data.WorkOrderNumber)
							flagError = false
						} else {
							log.Println("No work_order_number in response:", res.Message)
						}
						log.Print(res)
						req.Result = res
					}

					if flagError {
						req.Count++
						b_, err := json.Marshal(req)
						if err != nil {
							fmt.Println("Re Case Error_2:", req.CaseId)
						}
						if req.Count > maxRetry {
							utils.CaseSyncTempSet(req.CaseId, string(b_))
							utils.CaseSyncDel(req.CaseId)
						} else {

							utils.CaseSyncSet(req.CaseId, string(b_))
						}

					} else {
						fmt.Println("Re Case Success:", req.CaseId)
						utils.CaseSyncDel(req.CaseId)
					}

				}

				log.Printf("CaseId=%s Count=%d\n", req.CaseId, req.Count)
			}

			err = utils.OwnerCaseSyncSet("")
			if err != nil {
				log.Println("Redis SET error:", err)
				return err
			}
			log.Printf("Case Sync lock acquired/renewed by host: %s\n", hostname)
		} else {
			log.Printf("Case Sync lock held by another host: %s (this host: %s)\n", val, hostname)
		}

		counter++

		intervalStr := os.Getenv("ESB_RE_CREATE_INTERVAL")
		interval, err := strconv.Atoi(intervalStr)
		if err != nil {
			log.Printf("Invalid ESB_RE_CREATE_INTERVAL=%s, fallback to 1 min", intervalStr)
			interval = 1
		}

		sleep := time.Duration(interval) * time.Minute
		log.Print("Sleep : ", sleep)
		time.Sleep(sleep)
	}
}
