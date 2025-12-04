package handler

import (
	"log"
	"mainPackage/utils"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func SummaryReport(c *gin.Context) error {
	counter := 0
	//REPORT_NAME := os.Getenv("REPORT_NAME")

	hostname, err := os.Hostname()
	if err != nil {
		log.Println("Hostname error:", err)
		hostname = "unknown"
	}
	hostname = "SummaryReport_" + hostname

	log.Printf("Starting SummaryReport on host: %s\n", hostname)

	for {
		val, err := utils.OwnerReportGet()
		if err != nil {
			log.Println("Redis GET error:", err)
			//return err
		}

		log.Printf("[Tick %d] Redis sla key = '%s'\n", counter, val)
		val = ""
		// Only this host should write if no one else owns the lock or it owns it
		if val == "" || val == hostname {

			err = utils.OwnerReportSet(hostname)
			if err != nil {
				log.Println("Redis SET error:", err)
				//return err
			}

			conn, ctx, cancel := utils.ConnectDB_REPORT()
			if conn == nil {
				log.Printf("DB connection is nil")
				//return nil
			}
			defer cancel()
			defer conn.Close(ctx)

			// ---> Query storeprocedure
			_, err = conn.Exec(ctx, `CALL store_generate_summary_report();`)
			if err != nil {
				log.Println("CALL store_generate_summary_report error:", err)
				//return err
			}

			log.Println("Stored procedure store_generate_summary_report executed successfully")

			err = utils.OwnerReportSet("")
			if err != nil {
				log.Println("Redis SET error:", err)
				return err
			}
			log.Printf("Report lock acquired/renewed by host: %s\n", hostname)
		} else {
			log.Printf("Report lock held by another host: %s (this host: %s)\n", val, hostname)
		}

		counter++

		intervalStr := os.Getenv("MONITOR_REPORT_INTERVAL")
		interval, err := strconv.Atoi(intervalStr)
		if err != nil {
			log.Printf("Invalid MONITOR_REPORT_INTERVAL=%s, fallback to 1 min", intervalStr)
			interval = 1
		}

		sleep := time.Duration(interval) * time.Minute
		log.Print("Sleep : ", sleep)
		time.Sleep(sleep)
	}
}

func CaseHistory(c *gin.Context) error {
	counter := 0
	//REPORT_NAME := os.Getenv("REPORT_NAME")

	hostname, err := os.Hostname()
	if err != nil {
		log.Println("Hostname error:", err)
		hostname = "unknown"
	}

	hostname = "CaseHistory_" + hostname

	log.Printf("Starting SummaryReport on host: %s\n", hostname)

	for {
		val, err := utils.OwnerReportGet()
		if err != nil {
			log.Println("Redis GET error:", err)
			//return err
		}

		log.Printf("[Tick %d] Redis sla key = '%s'\n", counter, val)
		val = ""
		// Only this host should write if no one else owns the lock or it owns it
		if val == "" || val == hostname {

			err = utils.OwnerReportSet(hostname)
			if err != nil {
				log.Println("Redis SET error:", err)
				//return err
			}

			conn, ctx, cancel := utils.ConnectDB_REPORT()
			if conn == nil {
				log.Printf("DB connection is nil")
				//return nil
			}
			defer cancel()
			defer conn.Close(ctx)

			// ---> Query storeprocedure
			_, err = conn.Exec(ctx, `CALL store_generate_case_history();`)
			if err != nil {
				log.Println("CALL store_generate_case_history error:", err)
				//return err
			}

			log.Println("Stored procedure store_generate_case_history executed successfully")

			err = utils.OwnerReportSet("")
			if err != nil {
				log.Println("Redis SET error:", err)
				return err
			}
			log.Printf("Report lock acquired/renewed by host: %s\n", hostname)
		} else {
			log.Printf("Report lock held by another host: %s (this host: %s)\n", val, hostname)
		}

		counter++

		intervalStr := os.Getenv("MONITOR_REPORT_INTERVAL")
		interval, err := strconv.Atoi(intervalStr)
		if err != nil {
			log.Printf("Invalid MONITOR_REPORT_INTERVAL=%s, fallback to 1 min", intervalStr)
			interval = 1
		}

		sleep := time.Duration(interval) * time.Minute
		log.Print("Sleep : ", sleep)
		time.Sleep(sleep)
	}
}
