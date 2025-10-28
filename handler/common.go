package handler

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mainPackage/model"
	"mainPackage/utils"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

func ToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64:
		return strconv.FormatInt(reflect.ValueOf(v).Int(), 10)
	case uint, uint8, uint16, uint32, uint64:
		return strconv.FormatUint(reflect.ValueOf(v).Uint(), 10)
	case float32, float64:
		return strconv.FormatFloat(reflect.ValueOf(v).Float(), 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", value)
	}
}

func ToInt(value interface{}) int {
	switch v := value.(type) {
	case int:
		return v
	case int8, int16, int32, int64:
		return int(reflect.ValueOf(v).Int())
	case uint, uint8, uint16, uint32, uint64:
		return int(reflect.ValueOf(v).Uint())
	case float32, float64:
		return int(reflect.ValueOf(v).Float())
	case string:
		num, _ := strconv.Atoi(strings.TrimSpace(v))
		return num // Returns 0 if string isn't a number
	case bool:
		if v {
			return 1
		}
		return 0
	default:
		return 0
	}
}

func deriveKey(passphrase string) []byte {
	hash := sha256.Sum256([]byte(passphrase))
	return hash[:] // 32 bytes
}

func encrypt(plaintext string) (string, error) {
	key := deriveKey(os.Getenv("SECRET_KEY"))

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)

	// base64-encode so it's printable
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decrypt(ciphertextBase64 string) (string, error) {
	// logger := config.GetLog()
	key := deriveKey(os.Getenv("SECRET_KEY"))

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {

		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func unmarshalToMap(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func unmarshalToSliceOfMaps(data []byte) ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	err := json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func CaseCurrentStageInsert(conn *pgx.Conn, ctx context.Context, c *gin.Context, req model.CustomCaseCurrentStage) error {
	logger := utils.GetLog()

	username := GetVariableFromToken(c, "username")
	orgId := GetVariableFromToken(c, "orgId")
	now := time.Now()

	log.Print("===CaseCurrentStageInsert===")
	// 1. Insert responder
	_, err := conn.Exec(ctx, `
        INSERT INTO tix_case_responders ("orgId","caseId","unitId","userOwner","statusId","createdAt","createdBy")
        VALUES ($1,$2,$3,$4,$5,NOW(),$6)
    `, orgId, req.CaseID, "case", username, req.StatusID, username)
	if err != nil {
		log.Print(err)
		return err
	}
	//log.Print("===CaseCurrentStageInsert===")
	// Step 2: Load workflow node from DB
	query := `
	SELECT t1.id, t1."orgId", t1."wfId", t1."nodeId", t1.versions, t1.type, t1.section, t1.data,
	       t1.pic, t1."group", t1."formId", t1."createdAt", t1."updatedAt", t1."createdBy", t1."updatedBy"
	FROM public.wf_nodes t1
	JOIN public.wf_definitions t2
	  ON t1."versions" = t2."versions" AND t1."wfId" = t2."wfId"
	WHERE t2."wfId" = $1 AND t1."nodeId" = $2 AND t2."orgId" = $3
	`

	logger.Debug("Loading workflow node",
		zap.String("query", query),
		zap.Any("params", []any{req.WfID, req.NodeID, orgId}),
	)

	var workflow model.WfNode
	err = conn.QueryRow(ctx, query, req.WfID, req.NodeID, orgId).Scan(
		&workflow.ID, &workflow.OrgID, &workflow.WfID, &workflow.NodeID,
		&workflow.Versions, &workflow.Type, &workflow.Section,
		&workflow.Data, &workflow.Pic, &workflow.Group, &workflow.FormID,
		&workflow.CreatedAt, &workflow.UpdatedAt, &workflow.CreatedBy, &workflow.UpdatedBy,
	)

	log.Print("===workflow")
	log.Print(workflow)

	if err != nil {
		if err == pgx.ErrNoRows {
			logger.Warn("No workflow node found")
			return fmt.Errorf("workflow node not found")
		}
		logger.Error("Failed to load workflow node", zap.Error(err))
		return err
	}

	// Step 3: Insert into tix_case_current_stage
	insertQuery := `
	INSERT INTO public.tix_case_current_stage(
		"orgId", "caseId", "wfId", "nodeId", "stageType", "unitId", "username", versions, type, section, data, pic, "group", "formId",
		"createdAt", "updatedAt", "createdBy", "updatedBy"
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,
		$12, $13, $14, $15, $16, $17, $18
	)
	`

	args := []interface{}{
		workflow.OrgID, req.CaseID, workflow.WfID, req.NodeID, "case", "", "", workflow.Versions,
		workflow.Type, workflow.Section, workflow.Data, workflow.Pic,
		workflow.Group, workflow.FormID, now, now, username, username,
	}

	logger.Debug("Inserting current stage",
		zap.String("query", insertQuery),
		zap.Any("args", args),
	)

	_, err = conn.Exec(ctx, insertQuery, args...)
	if err != nil {
		logger.Error("Insert failed", zap.Error(err))
		return err
	}

	logger.Info("Insert success", zap.String("caseId", req.CaseID))
	return nil
}

// UpdateCurrentStage replaces fn_dispatch_unit_stage
func UpdateCurrentStageCore(ctx *gin.Context, conn *pgx.Conn, req model.UpdateStageRequest) (model.Response, error) {
	var result model.Response
	logger := utils.GetLog()
	username := GetVariableFromToken(ctx, "username")
	orgId := GetVariableFromToken(ctx, "orgId")
	log.Print("====9===")
	// // 1. Insert responder
	// _, err := conn.Exec(ctx, `
	//     INSERT INTO tix_case_responders ("orgId","caseId","unitId","userOwner","statusId","createdAt","createdBy")
	//     VALUES ($1,$2,$3,$4,$5,NOW(),$6)
	// `, orgId, req.CaseId, req.UnitId, req.UnitUser, req.Status, username)
	// if err != nil {
	// 	return result, err
	// }

	log.Print(username)
	log.Print(orgId)
	log.Print("---INSERT---")
	log.Print("====10===")
	query := `
	SELECT  "caseId"::text, "wfId", "nodeId", "stageType", "unitId", "username", "versions", "type", "section", "data",
  "pic", "group", "formId"
		FROM tix_case_current_stage
		WHERE "caseId"=$1 and ( "stageType" = 'case' OR "unitId" = $2)
	`

	rows, err := conn.Query(ctx, query, req.CaseId, req.UnitId)
	if err != nil {
		log.Println("Query failed:", err)
		return model.Response{Status: "-1", Msg: "Failure.1", Desc: err.Error()}, err
	}
	defer rows.Close()

	log.Print("---XXXX---")
	log.Print(rows)

	var caseStages model.CurrentStage
	var unitStages model.CurrentStage
	for rows.Next() {
		var stage model.CurrentStage
		if err := rows.Scan(
			&stage.CaseId,
			&stage.WfID,
			&stage.NodeId,
			&stage.StageType, // order ต้องตรงกับ SELECT
			&stage.UnitID,
			&stage.UserOwner,
			&stage.Versions,
			&stage.Type,
			&stage.Section,
			&stage.Data,
			&stage.Pic,
			&stage.Group,
			&stage.FormId,
		); err != nil {
			log.Println("Row scan failed:", err)
			continue
		}
		log.Println("---CURRENT---", stage.StageType)
		//stages = append(stages, stage)
		if stage.StageType == "case" {
			log.Println("---CASE---")
			caseStages = stage
		}
		if stage.StageType == "unit" {
			log.Println("---UNIT--->>")
			log.Println(stage.UnitID)
			log.Println(req.UnitId)
			if stage.UnitID == req.UnitId {
				log.Println("---UNIT--->>2")
				unitStages = stage
			}

		}
	}

	log.Print("------CASE---")
	log.Print(caseStages)
	log.Print("------UNIT---")
	log.Print(unitStages)
	log.Print("---NEXT---")
	//return model.Response{}, err
	// 🔹 Step 2: Get all workflow nodes using wfId
	//wfId := caseStages.WfID
	//version := caseStages.Versions

	allNodes, nodeConn, allNodesId, dispatchNode, err := GetAllNodes(ctx, conn, orgId.(string), caseStages.WfID, caseStages.Versions, logger)
	if err != nil {
		return model.Response{Status: "-1", Msg: "Failure.2", Desc: err.Error()}, err
	}

	log.Print("Total Nodes:", len(allNodes))
	log.Print("Total Connections:", len(nodeConn))
	log.Print("Total NodeId:", len(allNodesId))

	log.Print("------NEXT---")
	log.Print("------nodeConn---")
	log.Print(nodeConn)
	log.Print("------allNodesId---")
	log.Print(allNodesId)

	log.Print("------CHECK---")
	// 🔹 Step 3: Check next node
	CaseNextNode, UnitNextNode, caseCount, unitCount := GetNextNode(allNodesId, nodeConn, caseStages, unitStages, logger)

	fmt.Println("Case Next Node:", CaseNextNode)
	fmt.Println("Unit Next Node:", UnitNextNode)
	fmt.Println("caseCount:", caseCount)
	fmt.Println("unitCount:", unitCount)

	//Check Stage
	dataMaps := UnitNextNode.Data.(map[string]interface{})
	data2 := dataMaps["data"].(map[string]interface{})
	config := data2["config"].(map[string]interface{})
	log.Print("======dataMaps==")
	log.Print(config["action"])

	//Check Close Case
	if req.ResID != "" {
		if config["action"] == req.Status {
			log.Print("--> 1. Case Close")
			Result, err := UpdateCaseCurrentStage(ctx, conn, req, CaseNextNode, "case", username.(string))
			if err != nil {
				return Result, err
			}

			//--Update tix_cases on time (Group status)
			Result_, err := DispatchUpdateCaseStatus(ctx, conn, req, username.(string))
			if err != nil {

				log.Printf("Update status failed: %v", err)
			} else {
				log.Print(Result_)
				log.Println("Case status updated successfully-1")
			}
			GenerateNotiAndComment(ctx, conn, req, orgId.(string), "0", &req.ResDetail)
			//-->New Function for close
			// UpdateBusKafka_WO(ctx, conn, req)
			return Result, err
		} else {
			log.Println("Status worng number-1 :", err)
			return model.Response{Status: "-1", Msg: "Failure.3", Desc: "Status worng number!"}, err
		}

	}

	if unitCount == 0 || (config["action"] == req.Status) {
		// 1. Insert responder
		log.Print("--> 1. Insert responder")
		_, err := conn.Exec(ctx, `
		    INSERT INTO tix_case_responders ("orgId","caseId","unitId","userOwner","statusId","createdAt","createdBy")
		    VALUES ($1,$2,$3,$4,$5,NOW(),$6)
		`, orgId, req.CaseId, req.UnitId, req.UnitUser, req.Status, username)
		if err != nil {
			return result, err
		}
	} else {
		log.Println("Status worng number-2 :", err)
		return model.Response{Status: "-2", Msg: "Failure.3", Desc: "Status worng number!"}, err
	}

	//return model.Response{}, err
	// 🔹 Step 4:  Update data
	if unitCount == 0 && CaseNextNode.Type == "dispatch" { //--First Unit for case
		//--Update current stage :  case
		Result, err := UpdateCaseCurrentStage(ctx, conn, req, CaseNextNode, "case", username.(string))
		if err != nil {
			return Result, err
		}

		//--insert current stage : unit
		Result, err = InsertUnitCurrentStage(ctx, conn, req, UnitNextNode, "unit", username.(string))
		if err != nil {
			return Result, err
		}

		//--Update tix_cases on time (Group status)
		Result, err = DispatchUpdateCaseStatus(ctx, conn, req, username.(string))
		if err != nil {
			log.Printf("Update status failed: %v", err)
		} else {
			log.Println("Case status updated successfully-2")
		}
		GenerateNotiAndComment(ctx, conn, req, orgId.(string), "0")
		UpdateBusKafka_WO(ctx, conn, req)
		return Result, err

	} else if unitCount == caseCount { //-- Unit relate Case
		//--Update current stage :  case
		Result, err := UpdateCaseCurrentStage(ctx, conn, req, CaseNextNode, "case", username.(string))
		if err != nil {
			return Result, err
		}

		//--Update current stage :  unit
		Result, err = UpdateCaseCurrentStage(ctx, conn, req, UnitNextNode, "unit", username.(string))
		if err != nil {
			return Result, err
		}

		//--Update tix_cases on time (Group status)
		Result, err = DispatchUpdateCaseStatus(ctx, conn, req, username.(string))
		if err != nil {
			log.Printf("Update status failed: %v", err)
		} else {
			log.Println("Case status updated successfully-3")
		}

		GenerateNotiAndComment(ctx, conn, req, orgId.(string), "0")
		UpdateBusKafka_WO(ctx, conn, req)

		return Result, err

	} else if unitCount > 0 && unitCount < caseCount { //--Second Unit follow SOP

		//--Update current stage :  unit
		log.Print("--> --Second Unit follow SOP ")
		Result, err := UpdateCaseCurrentStage(ctx, conn, req, UnitNextNode, "unit", username.(string))
		if err != nil {
			return Result, err
		}

		GenerateNotiAndComment(ctx, conn, req, orgId.(string), "0")
		UpdateBusKafka_WO(ctx, conn, req)
		return Result, err

	} else if unitCount == 0 { //--Second Unit - First dispatch
		//--insert current stage : unit
		log.Print("--> --Second Unit - First dispatch ")
		UnitNextNode = dispatchNode
		Result, err := InsertUnitCurrentStage(ctx, conn, req, UnitNextNode, "unit", username.(string))
		if err != nil {
			return Result, err
		}

		GenerateNotiAndComment(ctx, conn, req, orgId.(string), "0")
		UpdateBusKafka_WO(ctx, conn, req)
		return Result, err
	}

	return result, err

}

func GetAllNodes(
	ctx context.Context,
	conn *pgx.Conn,
	orgId, wfId, version string,
	logger *zap.Logger,
) ([]model.WorkflowNode, []model.WorkFlowConnection, map[string]model.WorkflowNode, model.WorkflowNode, error) {

	query := `
		SELECT "nodeId", "type", "section", "data"
		FROM wf_nodes
		WHERE "orgId"=$1 AND "wfId"=$2 AND "versions"=$3
		ORDER BY 
			CASE 
				WHEN "section" = 'nodes' THEN 1 
				WHEN "section" = 'connections' THEN 2 
				ELSE 3 
			END
	`

	rows, err := conn.Query(ctx, query, orgId, wfId, version)
	if err != nil {
		logger.Error("Failed to fetch workflow nodes", zap.Error(err))
		return nil, nil, nil, model.WorkflowNode{}, err
	}
	defer rows.Close()

	var dispatchNode model.WorkflowNode
	var allNodes []model.WorkflowNode
	var nodeConn []model.WorkFlowConnection
	allNodesId := make(map[string]model.WorkflowNode)
	nodeStart := ""
	for rows.Next() {
		var node model.WorkflowNode
		if err := rows.Scan(&node.NodeId, &node.Type, &node.Section, &node.Data); err != nil {
			logger.Error("Row scan failed", zap.Error(err))
			return nil, nil, nil, model.WorkflowNode{}, err
		}

		// เก็บ node ทั้งหมด
		allNodes = append(allNodes, node)
		allNodesId[node.NodeId] = node
		if node.Type == "dispatch" {
			dispatchNode = node
		}
		if node.Type == "start" {
			nodeStart = node.NodeId
		}
		// ถ้า section เป็น connections → parse data เป็น []WorkFlowConnection
		if node.Section == "connections" {
			dataBytes, err := json.Marshal(node.Data)
			if err != nil {
				logger.Error("Failed to marshal connection data", zap.Error(err))
				continue
			}

			var conns []model.WorkFlowConnection
			if err := json.Unmarshal(dataBytes, &conns); err != nil {
				logger.Error("Unmarshal connection failed", zap.Error(err))
				continue
			}

			nodeConn = append(nodeConn, conns...)
		}
	}

	order, err := OrderConnection(nodeConn, nodeStart)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("===order==")
	log.Print(order)
	log.Print("En===order==")
	return allNodes, order, allNodesId, dispatchNode, nil
}

func GetNextNode(
	allNodesId map[string]model.WorkflowNode,
	nodeConn []model.WorkFlowConnection,
	caseStages model.CurrentStage,
	unitStages model.CurrentStage,
	logger *zap.Logger,
) (model.WorkflowNode, model.WorkflowNode, int, int) {

	var CaseNextNode model.WorkflowNode
	var UnitNextNode model.WorkflowNode
	var unitCount = 0
	var caseCount = 0
	var rec = 0
	for _, wfConn := range nodeConn {

		//----- For Unit Stage
		logger.Info("---Unit Stage---", zap.Any("node", wfConn))
		if wfConn.Source == unitStages.NodeId {
			candidateCase := allNodesId[wfConn.Target]

			for candidateCase.Type == "sla" {
				found := false
				for _, c := range nodeConn {
					if c.Source == candidateCase.NodeId && c.Label == "yes" {
						candidateCase = allNodesId[c.Target]
						//logger.Info("---candidate---", zap.Any("node", candidateCase))
						if candidateCase.Type == "process" {
							found = true
							break
						}
					}
				}
				if !found {
					break
				}
			}

			UnitNextNode = candidateCase

			logger.Info("UNIT Next node (non-SLA)", zap.Any("node", UnitNextNode))
			break
		}
		unitCount = rec
		rec++
	}

	rec = 0
	for _, wfConn := range nodeConn {

		//----- For Case Stage
		logger.Info("---Case Stage---", zap.Any("node", wfConn))
		if wfConn.Source == caseStages.NodeId {
			candidateCase := allNodesId[wfConn.Target]

			// ถ้า node type เป็น SLA → ข้ามไปหา target ต่อไป
			for candidateCase.Type == "sla" {
				found := false
				for _, c := range nodeConn {
					if c.Source == candidateCase.NodeId && c.Label == "yes" {
						candidateCase = allNodesId[c.Target]
						logger.Info("---candidate--CASE-", zap.Any("node", candidateCase))
						if candidateCase.Type == "process" {
							found = true
							break
						}
					}
				}
				if !found {
					break
				}
			}

			CaseNextNode = candidateCase

			logger.Info("CASE Next node (non-SLA)", zap.Any("node", CaseNextNode))
			break
		}
		caseCount = rec
		rec++
	}

	// ✅ ถ้า UnitNextNode ยังว่าง ให้ใช้ CaseNextNode
	if UnitNextNode.NodeId == "" {
		UnitNextNode = CaseNextNode
		unitCount = 0
		logger.Info("UnitNextNode empty → fallback to CaseNextNode", zap.Any("node", UnitNextNode))
	}

	return CaseNextNode, UnitNextNode, caseCount, unitCount
}

func InsertUnitCurrentStage(
	ctx context.Context,
	conn *pgx.Conn,
	req model.UpdateStageRequest,
	nextStage model.WorkflowNode,
	stageType string,
	username string,
) (model.Response, error) {

	var node model.WfNode
	nodeQuery := `
		SELECT n."orgId", n."wfId", n."nodeId", d."versions", n."type", n."section", 
       n."formId", n."pic", n."group"
		FROM public."wf_nodes" n
		JOIN public."wf_definitions" d 
		ON n."wfId" = d."wfId"
		AND n."versions" = d."versions"
		WHERE n."nodeId" = $1  
	`
	log.Print("-----SELECT-NODE--")
	log.Print(nextStage.NodeId)
	err := conn.QueryRow(ctx, nodeQuery, nextStage.NodeId).Scan(
		&node.OrgID, &node.WfID, &node.NodeID, &node.Versions, &node.Type,
		&node.Section, &node.FormID, &node.Pic, &node.Group,
	)
	if err != nil {
		return model.Response{Status: "-1", Msg: "Failure.InsertUnitCurrentStage.1", Desc: err.Error()}, err
	}
	log.Print(nextStage.Data)

	// Marshal nextStage.Data to JSON for jsonb column
	dataBytes, err := json.Marshal(nextStage.Data)
	if err != nil {
		return model.Response{Status: "-1", Msg: "Failure.InsertUnitCurrentStage.2", Desc: err.Error()}, err
	}

	// Ensure optional string fields are non-nil
	// pic := ""
	// if node.Pic != nil {
	// 	pic = *node.Pic
	// }

	// group := ""
	// if node.Group != nil {
	// 	group = *node.Group
	// }

	// formId := ""
	// if node.FormID != nil {
	// 	formId = *node.FormID
	// }

	now := time.Now()

	insertQuery := `
	INSERT INTO public."tix_case_current_stage"
	("orgId", "caseId", "wfId", "nodeId", "versions", "type", "section", "data",
	  "stageType", "unitId",
	 "username", "updatedAt", "createdAt", "createdBy", "updatedBy")
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14, $15)
	`
	_, err = conn.Exec(ctx, insertQuery,
		node.OrgID, req.CaseId, node.WfID, req.NodeId, node.Versions, node.Type, node.Section, dataBytes,
		stageType, req.UnitId,
		req.UnitUser, now, now, username, username,
	)
	if err != nil {
		return model.Response{Status: "-1", Msg: "Failure.InsertUnitCurrentStage.3", Desc: err.Error()}, err
	}

	return model.Response{Status: "0", Msg: "Success", Desc: "InsertUnitCurrentStage"}, nil

}

func UpdateCaseCurrentStage(
	ctx *gin.Context,
	conn *pgx.Conn,
	req model.UpdateStageRequest,
	nextStage model.WorkflowNode,
	stageType string,
	username string,
) (model.Response, error) {

	var node model.WfNode
	nodeQuery := `
		SELECT n."orgId", n."wfId", n."nodeId", d."versions", n."type", n."section", 
       n."formId", n."pic", n."group"
		FROM public."wf_nodes" n
		JOIN public."wf_definitions" d 
		ON n."wfId" = d."wfId"
		AND n."versions" = d."versions"
		WHERE n."nodeId" = $1  
	`
	log.Print("-----SELECT-NODE--")
	log.Print(nextStage.NodeId)
	err := conn.QueryRow(ctx, nodeQuery, nextStage.NodeId).Scan(
		&node.OrgID, &node.WfID, &nextStage.NodeId, &node.Versions, &node.Type,
		&node.Section, &node.FormID, &node.Pic, &node.Group,
	)
	if err != nil {
		return model.Response{Status: "-1", Msg: "Failure.UpdateCaseCurrentStage.1-" + stageType, Desc: err.Error()}, err
	}

	// Marshal nextStage.Data to JSON for jsonb column
	dataBytes, err := json.Marshal(nextStage.Data)
	if err != nil {
		return model.Response{Status: "-1", Msg: "Failure.UpdateCaseCurrentStage.2-" + stageType, Desc: err.Error()}, err
	}

	now := time.Now()

	log.Print("---Update---")

	if stageType == "case" {
		req.UnitId = ""
		req.UnitUser = ""
	}
	updateQuery := `
	UPDATE public."tix_case_current_stage"
	SET "wfId" = $1,
	    "type" = $2,
	    "section" = $3,
	    "data" = $4,
	     
	    
	    "username" = $6,
	    "updatedAt" = $7,
	    "updatedBy" = $8,
		"nodeId" = $12,
		"versions" = $11 
	WHERE "caseId" = $9
	  AND "stageType" = $10 
	  AND "unitId" = $5
	`

	_, err = conn.Exec(ctx, updateQuery,
		node.WfID, node.Type, node.Section, dataBytes,

		req.UnitId, req.UnitUser, now, username,
		req.CaseId, stageType, node.Versions, nextStage.NodeId,
	)

	if err != nil {
		return model.Response{Status: "-1", Msg: "Failure.UpdateCaseCurrentStage.3-" + stageType, Desc: err.Error()}, err
	}

	//GenerateNotiAndComment(ctx, conn, req, node.OrgID, "0")

	return model.Response{Status: "0", Msg: "Success", Desc: "UpdateCaseCurrentStage-" + stageType}, nil
}

func OrderConnection(connections []model.WorkFlowConnection, nodeStart string) ([]model.WorkFlowConnection, error) {
	// Build graph with adjacency list of connections
	graph := buildGraph(connections)

	startNode := nodeStart
	visited := make(map[string]bool)
	var orderedConns []model.WorkFlowConnection

	dfsConnections(graph, startNode, visited, &orderedConns)

	return orderedConns, nil
}

func buildGraph(conns []model.WorkFlowConnection) map[string][]model.WorkFlowConnection {
	graph := make(map[string][]model.WorkFlowConnection)
	for _, c := range conns {
		graph[c.Source] = append(graph[c.Source], c)
	}
	return graph
}

func dfsConnections(
	graph map[string][]model.WorkFlowConnection,
	node string,
	visited map[string]bool,
	order *[]model.WorkFlowConnection,
) {
	if visited[node] {
		return
	}
	visited[node] = true

	// Traverse each outgoing connection
	for _, conn := range graph[node] {
		*order = append(*order, conn)
		dfsConnections(graph, conn.Target, visited, order)
	}
}

func DispatchUpdateCaseStatus(ctx *gin.Context, conn *pgx.Conn, req model.UpdateStageRequest, username string) (model.Response, error) {

	orgId := GetVariableFromToken(ctx, "orgId")
	log.Print("===DispatchUpdateCaseStatus===")
	// 1. Insert responder
	_, err := conn.Exec(ctx, `
        INSERT INTO tix_case_responders ("orgId","caseId","unitId","userOwner","statusId","createdAt","createdBy")
        VALUES ($1,$2,$3,$4,$5,NOW(),$6)
    `, orgId, req.CaseId, "case", username, req.Status, username)
	if err != nil {
		log.Print(err)
		return model.Response{Status: "-1", Msg: "Failure.DispatchUpdateCaseStatus.0-" + req.CaseId, Desc: err.Error()}, err
	}

	now := time.Now()

	if req.ResID != "" {
		log.Print("===DispatchUpdateCaseStatus--A===")
		query := `
    UPDATE public."tix_cases"
    SET "statusId" = $1,
        "updatedAt" = $2,
        "updatedBy" = $3,
		"resId" = $4,
		"resDetail" = $5,
		"overSlaFlag" = $7,
		"overSlaDate" = $8,
		"overSlaCount" = 0
    WHERE "caseId" = $6;
    `
		log.Print("====XXXXXXXX===")
		cmd, err := conn.Exec(ctx, query, req.Status, now, username, req.ResID, req.ResDetail, req.CaseId, false, nil)
		if err != nil {
			return model.Response{Status: "-1", Msg: "Failure.DispatchUpdateCaseStatus.1-" + req.CaseId, Desc: err.Error()}, err
		}

		if cmd.RowsAffected() == 0 {
			return model.Response{Status: "-1", Msg: "Failure.DispatchUpdateCaseStatus.2-" + req.CaseId, Desc: err.Error()}, err
		}

	} else {
		log.Print("===DispatchUpdateCaseStatus--B===")
		query := `
    UPDATE public."tix_cases"
    SET "statusId" = $1,
        "updatedAt" = $2,
        "updatedBy" = $3,
		"overSlaFlag" = $5,
		"overSlaDate" = $6,
		"overSlaCount" = 0
    WHERE "caseId" = $4;
    `

		cmd, err := conn.Exec(ctx, query, req.Status, now, username, req.CaseId, false, nil)
		if err != nil {
			return model.Response{Status: "-1", Msg: "Failure.DispatchUpdateCaseStatus.1-" + req.CaseId, Desc: err.Error()}, err
		}

		if cmd.RowsAffected() == 0 {
			return model.Response{Status: "-1", Msg: "Failure.DispatchUpdateCaseStatus.2-" + req.CaseId, Desc: err.Error()}, err
		}
	}

	return model.Response{Status: "0", Msg: "Success", Desc: "DispatchUpdateCaseStatus-" + req.CaseId}, nil
}

func GetUserSkills(ctx context.Context, conn *pgx.Conn, orgID string) ([]model.GetSkills, error) {
	query := `
		SELECT "skillId", "en", "th"
		FROM public.um_skills
		WHERE "orgId" = $1 AND "active" = true
		ORDER BY id ASC;
	`

	rows, err := conn.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var skills []model.GetSkills
	for rows.Next() {
		var s model.GetSkills
		if err := rows.Scan(&s.SkillID, &s.En, &s.Th); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		skills = append(skills, s)
	}

	// If no records, return empty slice instead of nil
	if skills == nil {
		return []model.GetSkills{}, nil
	}

	return skills, nil
}

// ConvertSkills returns only the skills whose SkillId exists in data
func ConvertSkills(skills []model.GetSkills, data []string) []model.GetSkills {
	result := []model.GetSkills{}
	// build a quick lookup map for data
	dataMap := make(map[string]bool)
	for _, id := range data {
		dataMap[id] = true
	}

	// filter skills
	for _, s := range skills {
		if dataMap[s.SkillID] {
			result = append(result, s)
		}
	}
	return result
}

func GetUnitProp(ctx context.Context, conn *pgx.Conn, orgID string) ([]model.GetUnisProp, error) {
	query := `
		SELECT "propId", "en", "th"
		FROM public.mdm_properties
		WHERE "orgId" = $1 AND "active" = true
		ORDER BY id ASC;
	`

	rows, err := conn.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var props []model.GetUnisProp
	for rows.Next() {
		var s model.GetUnisProp
		if err := rows.Scan(&s.PropId, &s.En, &s.Th); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		props = append(props, s)
	}

	// If no records, return empty slice instead of nil
	if props == nil {
		return []model.GetUnisProp{}, nil
	}

	return props, nil
}

// ConvertProps returns only the skills whose PropId exists in data
func ConvertProps(props []model.GetUnisProp, data []string) []model.GetUnisProp {
	result := []model.GetUnisProp{}
	// build a quick lookup map for data
	dataMap := make(map[string]bool)
	for _, id := range data {
		dataMap[id] = true
	}

	// filter skills
	for _, s := range props {
		if dataMap[s.PropId] {
			result = append(result, s)
		}
	}
	return result
}

func GenerateNotiAndComment(ctx *gin.Context,
	conn *pgx.Conn,
	req model.UpdateStageRequest,
	orgId string,
	delay string,
	additionalMessage ...*string,
) error {
	statuses, err := utils.GetCaseStatusList(ctx, conn, orgId)
	if err != nil {
		return err
	}
	statusMap := make(map[string]model.CaseStatus)
	for _, s := range statuses {
		statusMap[*s.StatusID] = s
	}
	statusName := statusMap[req.Status]

	if statusName.Th == nil {
		return fmt.Errorf("status %s has nil Th field", req.Status)
	}

	log.Print("====statusName===")
	log.Print(statusName)
	provID, wfId, versions, err := GetInfoFromCase(ctx, conn, orgId, req.CaseId)
	if err != nil {
		log.Printf("error getting provId: %v", err)
	} else {
		log.Printf("provId = %s", provID)
		log.Print(provID, wfId, versions)
	}

	data := []model.Data{
		{Key: "delay", Value: delay}, //0=white, 1=yellow , 2=red
	}
	recipients := []model.Recipient{
		{Type: "provId", Value: provID},
	}

	username := GetVariableFromToken(ctx, "username")
	msg := *statusName.Th
	if username != req.UnitUser {
		if req.Status == os.Getenv("REQUESTCLOSE") {
			if len(additionalMessage) > 0 {
				msg = msg + " : " + *additionalMessage[0]
			}
		} else if req.Status != os.Getenv("ASSIGNED") {
			msg = *statusName.Th + "( แทน " + req.UnitUser + ")"
		} else {
			msg = *statusName.Th + "(" + req.UnitUser + ")"
		}
	}

	msg_alert := msg + " :: " + req.CaseId
	var event = "CASE-STATUS-UPDATE"
	additionalJsonMap := map[string]interface{}{
		"event":    event,
		"caseId":   req.CaseId,
		"status":   req.Status,
		"ms_alert": msg_alert,
	}
	additionalJSON, err := json.Marshal(additionalJsonMap)
	if err != nil {
		log.Print("covent additionalData Error :", err)
	}
	additionalData := json.RawMessage(additionalJSON)
	genNotiCustom(ctx, conn, orgId, username.(string), username.(string), "", *statusName.Th, data, msg_alert, recipients, "/case/"+req.CaseId, "User", event, &additionalData)

	evt := model.CaseHistoryEvent{
		OrgID:     orgId,
		CaseID:    req.CaseId,
		Username:  username.(string),
		Type:      "event",
		FullMsg:   msg,
		JsonData:  "",
		CreatedBy: username.(string),
	}

	log.Print("====InsertCaseHistoryEvent===")
	err = InsertCaseHistoryEvent(ctx, conn, evt)
	if err != nil {
		log.Fatalf("Insert failed: %v", err)
	}

	return nil
}

func GetInfoFromCase(ctx context.Context, conn *pgx.Conn, orgId string, caseID string) (string, string, string, error) {
	var provID, wfId, versions string

	query := `
        SELECT "provId", "wfId", "versions"
        FROM public.tix_cases
        WHERE "orgId" = $1
		AND "caseId" = $2
        LIMIT 1
    `
	err := conn.QueryRow(ctx, query, orgId, caseID).Scan(&provID, &wfId, &versions)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", "", "", fmt.Errorf("no provId found for caseId: %s", caseID)
		}
		return "", "", "", err
	}

	return provID, wfId, versions, nil
}

func InsertCaseHistoryEvent(ctx context.Context, conn *pgx.Conn, evt model.CaseHistoryEvent) error {
	// แปลง JsonData เป็น JSON string
	var jsonDataStr *string
	if evt.JsonData != nil {
		b, err := json.Marshal(evt.JsonData)
		if err != nil {
			return err
		}
		s := string(b)
		jsonDataStr = &s
	}

	query := `
        INSERT INTO tix_case_history_events (
            "orgId", "caseId", username, type, "fullMsg", "jsonData", "createdBy"
        ) VALUES ($1, $2, $3, $4, $5, $6, $7)
    `

	_, err := conn.Exec(ctx, query,
		evt.OrgID,
		evt.CaseID,
		evt.Username,
		evt.Type,
		evt.FullMsg,
		jsonDataStr,
		evt.CreatedBy,
	)

	return err
}

func CalDashboardCaseSummary(ctx context.Context, conn *pgx.Conn, orgId string, recipients []model.Recipient, username, groupTypeId, countryId, provId, distId string) error {
	now := time.Now()
	currentDate := now.Format("2006/01/02")
	currentHour := now.Hour()
	currentTime := fmt.Sprintf("%02d:00:00", currentHour) // e.g., "01:00:00"

	// ✅ โหลดข้อมูล groupType จาก Redis หรือ DB
	groupTypes, err := utils.GroupTypeGetOrLoad(conn)
	if err != nil {
		return fmt.Errorf("cannot load group types: %v", err)
	}

	// ✅ ตรวจสอบ groupTypeId ภายใน field groupTypeLists
	found := false
	for _, g := range groupTypes {
		for _, id := range g.GroupTypeLists {
			if id == groupTypeId {
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return fmt.Errorf("groupTypeId not found in any groupTypeLists: %s", groupTypeId)
	}

	// ✅ UPSERT (เพิ่ม +1 ถ้ามีอยู่แล้ว)
	query := `
		INSERT INTO d_case_summary 
			("orgId", date, time, "groupTypeId", "countryId", "provId", "distId", total)
		VALUES ($1, $2, $3, $4, $5, $6, $7, '1')
		ON CONFLICT ("orgId", date, time, "groupTypeId", "countryId", "provId", "distId")
		DO UPDATE SET total = (d_case_summary.total::int + 1)::varchar;
	`

	_, err = conn.Exec(context.Background(), query, orgId, currentDate, currentTime, groupTypeId, countryId, provId, distId)
	if err != nil {
		return fmt.Errorf("failed to upsert summary: %v", err)
	}

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
