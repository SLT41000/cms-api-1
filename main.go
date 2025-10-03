// @title CMS API
// @version 1.0
// @termsOfService http://somewhere.com/
// @BasePath /
// @contact.name API Support
// @contact.url http://somewhere.com/support
// @contact.email support@somewhere.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @schemes http https

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mainPackage/config"
	_ "mainPackage/docs"
	"mainPackage/handler"
	"mainPackage/model"
	"os"
	"strconv"
	"time"

	"github.com/IBM/sarama"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/ulule/limiter/v3"
	ginlimiter "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"go.uber.org/zap"
)

func init() {
	// Load .env file
	logger := config.GetLog()
	err := godotenv.Load()
	if err != nil {
		logger.Fatal("Error loading .env file")
	}
}
func main() {
	rate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  50,
	}
	go handler.StartAutoDeleteScheduler()

	go func() {
		if err := StartKafkaWorker_Create(); err != nil {
			log.Printf("Kafka worker error: %v", err)
		}
	}()

	go func() {
		if err := StartKafkaWorker_Update(); err != nil {
			log.Printf("Kafka worker error: %v", err)
		}
	}()

	store := memory.NewStore()
	instance := limiter.New(store, rate)
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	// router.Use(cors.Default())
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173",             // dev
			"https://cms-sigma-woad.vercel.app", // production frontend
			"https://cms.welcomedcc.com",
			"https://welcome-service-stg.metthier.ai:65000",
			"https://welcome-cms-stg.metthier.ai:65000",
			"https://mettlink-workorder-service",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.Use(gin.Recovery())
	router.Use(ginlimiter.NewMiddleware(instance))
	auth := router.Group("/api/v1/auth")
	{
		auth.GET("/login", handler.UserLogin)
		auth.POST("/login", handler.UserLoginPost)
		auth.POST("/add", handler.UserAddAuth)
		auth.POST("/refresh", handler.RefreshToken)
	}
	v1 := router.Group("/api/v1")
	{
		v1.Use(handler.ProtectedHandler)
		v1.GET("/area/country_province_districts", handler.GetCountryProvinceDistricts)

		v1.GET("/forms", handler.GetForm)
		v1.GET("/forms/getAllForms", handler.GetAllForm)
		v1.POST("/forms", handler.FormInsert)
		v1.PATCH("/forms/:uuid", handler.FormUpdate)
		v1.PATCH("/forms/active", handler.FormActive)
		v1.PATCH("/forms/lock", handler.FormLock)
		v1.PATCH("/forms/publish", handler.FormPublish)
		v1.POST("/forms/casesubtype", handler.GetFormByCaseSubType)
		v1.GET("/workflows", handler.GetWorkFlowList)
		v1.GET("/workflows/:id", handler.GetWorkFlow)
		v1.POST("/workflows", handler.WorkFlowInsert)
		v1.PATCH("/workflows/:uuid", handler.WorkFlowUpdate)
		v1.DELETE("/workflows/:uuid", handler.WorkflowDelete)

		v1.GET("/case", handler.ListCase)
		v1.GET("/case/:id", handler.CaseById)
		v1.GET("/caseId/:caseId", handler.CaseByCaseId)
		v1.POST("/case/add", handler.InsertCase)
		v1.PATCH("/case/:id", handler.UpdateCase)
		v1.DELETE("/case/:id", handler.DeleteCase)
		v1.GET("/caseResult/", handler.CaseResult)

		v1.GET("/case_status", handler.GetCaseStatus)
		v1.GET("/case_status/:id", handler.GetCaseStatusById)
		v1.POST("/case_status/add", handler.InsertCaseStatus)
		v1.PATCH("/case_status/:id", handler.UpdateCaseStatus)
		v1.DELETE("/case_status/:id", handler.DeleteCaseStatus)

		v1.GET("/casetypes", handler.ListCaseType)
		v1.POST("/casetypes/add", handler.InsertCaseType)
		v1.PATCH("/casetypes/:id", handler.UpdateCaseType)
		v1.DELETE("/casetypes/:id", handler.DeleteCaseType)
		v1.GET("/casetypes_with_subtype", handler.ListCaseTypeWithSubtype)
		v1.GET("/casesubtypes", handler.ListCaseSubType)
		v1.POST("/casesubtypes/add", handler.InsertCaseSubType)
		v1.PATCH("/casesubtypes/:id", handler.UpdateCaseSubType)
		v1.DELETE("/casesubtypes/:id", handler.DeleteCaseSubType)

		v1.GET("/departments", handler.GetDepartment)
		v1.GET("/departments/:id", handler.GetDepartmentbyId)
		v1.POST("/departments/add", handler.InsertDepartment)
		v1.PATCH("/departments/:id", handler.UpdateDepartment)
		v1.DELETE("/departments/:id", handler.DeleteDepartment)

		v1.GET("/commands", handler.GetCommand)
		v1.GET("/commands/:id", handler.GetCommandById)
		v1.POST("/commands/add", handler.InsertCommand)
		v1.PATCH("/commands/:id", handler.UpdateCommand)
		v1.DELETE("/commands/:id", handler.DeleteCommand)

		v1.GET("/department_command_stations", handler.GetDepartmentCommandStation)
		v1.GET("/stations", handler.GetStation)
		v1.GET("/stations/:id", handler.GetStationbyId)
		v1.POST("/stations/add", handler.InsertStations)
		v1.PATCH("/stations/:id", handler.UpdateStations)
		v1.DELETE("/stations/:id", handler.DeleteStations)

		v1.GET("/role", handler.GetRole)
		v1.GET("/role/:id", handler.GetRolebyId)
		v1.POST("/role/add", handler.InsertRole)
		v1.PATCH("/role/:id", handler.UpdateRole)
		v1.DELETE("/role/:id", handler.DeleteRole)

		v1.GET("/permission", handler.GetPermission)
		v1.GET("/permission/:permId", handler.GetPermissionById)
		v1.POST("/permission/add", handler.InsertPermission)
		v1.PATCH("/permission/:permId", handler.UpdatePermission)
		v1.DELETE("/permission/:permId", handler.DeletePermission)

		v1.GET("/role_permission", handler.GetRolePermission)
		v1.GET("/role_permission/:id", handler.GetRolePermissionbyId)
		v1.GET("/role_permission/roleId/:roleId", handler.GetRolePermissionbyroleId)
		v1.POST("/role_permission/add", handler.InsertRolePermission)
		v1.PATCH("/role_permission/:roleId", handler.UpdateRolePermission)
		v1.PATCH("/role_permission/multi", handler.UpdateMultiRolePermission)
		v1.DELETE("/role_permission/:id", handler.DeleteRolePermission)

		v1.GET("/customer", handler.CustomerList)
		v1.POST("/customer/add", handler.CustomerAdd)
		v1.GET("/customer/:id", handler.CustomerById)
		v1.PATCH("/customer/:id", handler.CustomerUpdate)
		v1.DELETE("/customer/:id", handler.CustomerDelete)

		v1.GET("/customer_contacts", handler.CustomerContactList)
		v1.POST("/customer_contacts/add", handler.CustomerContactAdd)
		v1.GET("/customer_contacts/:id", handler.CustomerContactById)
		v1.PATCH("/customer_contacts/:id", handler.CustomerContactUpdate)
		v1.DELETE("/customer_contacts/:id", handler.CustomerContactDelete)

		v1.GET("/customer_with_socials", handler.CustomerSocialList)
		v1.POST("/customer_with_socials/add", handler.CustomerSocialAdd)
		v1.GET("/customer_with_socials/:id", handler.CustomerWithSocialById)
		v1.PATCH("/customer_with_socials/:id", handler.CustomerSocialUpdate)
		v1.DELETE("/customer_with_socials/:id", handler.CustomerSocialDelete)

		v1.GET("/users", handler.GetUmUserList)
		v1.GET("/users/:id", handler.GetUmUserById)
		v1.POST("/users/add", handler.UserAdd)
		v1.PATCH("/users/:id", handler.UserUpdate)
		v1.DELETE("/users/:id", handler.UserDelete)

		v1.PATCH("/users/change_password/:id", handler.ChangeUserPassword)
		v1.GET("/users/username/:username", handler.GetUmUserByUsername)
		v1.PATCH("/users/username/:username", handler.UserUpdateByUsername)
		v1.GET("/users_with_skills", handler.GetUserWithSkills)
		v1.GET("/users_with_skills/:id", handler.GetUserWithSkillsById)
		v1.GET("/users_with_skills/skillId/:skillId", handler.GetUserWithSkillsBySkillId)
		v1.POST("/users_with_skills/add", handler.InsertUserWithSkills)
		v1.PATCH("/users_with_skills/:id", handler.UpdateUserWithSkills)
		v1.DELETE("/users_with_skills/:id", handler.DeleteUserWithSkills)
		v1.GET("/users_with_contacts", handler.GetUserWithContacts)
		v1.GET("/users_with_contacts/:id", handler.GetUserWithContactsById)
		v1.POST("/users_with_contacts/add", handler.InsertUserWithContacts)
		v1.PATCH("/users_with_contacts/:id", handler.UpdateUserWithContacts)
		v1.DELETE("/users_with_contacts/:id", handler.DeleteUserWithContacts)
		v1.GET("/users_with_socials", handler.GetUserWithSocials)
		v1.GET("/users_with_socials/:id", handler.GetUserWithSocialsById)
		v1.POST("/users_with_socials/add", handler.InsertUserWithSocials)
		v1.PATCH("/users_with_socials/:id", handler.UpdateUserWithSocials)
		v1.DELETE("/users_with_socials/:id", handler.DeleteUserWithSocials)
		v1.GET("/user_groups/all", handler.GetUmGroupList)

		v1.GET("/mdm/properties", handler.GetMmdProperty)
		v1.GET("/mdm/properties/:id", handler.GetMmdPropertyById)
		v1.POST("/mdm/properties/add", handler.InsertMmdProperty)
		v1.PATCH("/mdm/properties/:id", handler.UpdateMmdProperty)
		v1.DELETE("/mdm/properties/:id", handler.DeleteMmdProperty)

		v1.GET("/mdm/sources", handler.GetMmdUnitSources)
		v1.GET("/mdm/sources/:id", handler.GetMmdUnitSourcesById)
		v1.POST("/mdm/sources/add", handler.InsertMmdUnitSources)
		v1.PATCH("/mdm/sources/:id", handler.UpdateMmdUnitSources)
		v1.DELETE("/mdm/sources/:id", handler.DeleteMmdUnitSources)

		v1.GET("/mdm/types", handler.GetMmdUnitType)
		v1.GET("/mdm/types/:id", handler.GetMmdUnitTypeById)
		v1.POST("/mdm/types/add", handler.InsertMmdUnitType)
		v1.PATCH("/mdm/types/:id", handler.UpdateMmdUnitType)
		v1.DELETE("/mdm/types/:id", handler.DeleteMmdUnitType)

		v1.GET("/mdm/companies", handler.GetMmdCompanies)
		v1.GET("/mdm/companies/:id", handler.GetMmdCompaniesById)
		v1.POST("/mdm/companies/add", handler.InsertMmdCompanies)
		v1.PATCH("/mdm/companies/:id", handler.UpdateMmdCompanies)
		v1.DELETE("/mdm/companies/:id", handler.DeleteMmdCompanies)

		v1.GET("/mdm/status", handler.GetMmdUnitStatus)
		v1.GET("/mdm/status/:id", handler.GetMmdUnitStatusById)
		v1.POST("/mdm/status/add", handler.InsertMmdUnitStatus)
		v1.PATCH("/mdm/status/:id", handler.UpdateMmdUnitStatus)
		v1.DELETE("/mdm/status/:id", handler.DeleteMmdUnitStatus)

		v1.GET("/mdm/units", handler.GetMmdUnit)
		v1.GET("/mdm/units/:id", handler.GetMmdUnitById)
		v1.POST("/mdm/units/add", handler.InsertMmdUnit)
		v1.PATCH("/mdm/units/:id", handler.UpdateMmdUnit)
		v1.DELETE("/mdm/units/:id", handler.DeleteMmdUnit)

		v1.GET("/mdm/units/properties/unitId", handler.GetMmdUnitWithProperty)

		v1.GET("/dispatch/:caseId/SOP", handler.GetSOP)
		v1.GET("/dispatch/:caseId/units", handler.GetUnit)
		v1.POST("/dispatch/event", handler.UpdateCurrentStage)

		v1.GET("/dispatch/:caseId/SOP/unit/:unitId", handler.GetUnitSOP)

		v1.GET("/audit_log", handler.GetAuditlog)
		v1.GET("/audit_log/:username", handler.GetAuditlogByUsername)

		v1.GET("/case_history", handler.GetCaseHistory)
		v1.GET("/case_history/:caseId", handler.GetCaseHistoryByCaseId)
		v1.POST("/case_history/add", handler.InsertCaseHistory)
		v1.PATCH("/case_history/:id", handler.UpdateCaseHistory)
		v1.DELETE("/case_history/:id", handler.DeleteCaseHistory)

		v1.GET("/devices", handler.GetDeviceIoT)
		v1.GET("/devices/:id", handler.GetDeviceIoTById)

	}

	minimal := router.Group("/api/minimal")
	{
		minimal.POST("/case/create", handler.MinimalCreateCase)
	}

	nonAuth := router.Group("/api/v1")
	{
		nonAuth.POST("/users/reset_password", handler.ResetUserPassword)
		nonAuth.GET("/generate_caseid", handler.GenerateCaseIDHandler)
	}
	health := router.Group("/")
	{
		health.GET("/health", handler.Health)
	}

	notifications := router.Group("/api/v1/notifications")
	{

		v1.Use(handler.ProtectedHandler)
		notifications.GET("/register", handler.WebSocketHandler)
		notifications.POST("/", handler.CreateNotifications)
		notifications.GET("/:orgId/:username", handler.GetNotificationsForUser)
		notifications.PUT("/:id", handler.UpdateNotification)
		notifications.DELETE("/:id", handler.DeleteNotification)
	}
	logger := config.GetLog()
	SERV_ADDR := os.Getenv("SERV_ADDR")
	SERV_PORT := os.Getenv("SERV_PORT")
	URL := "http://" + SERV_ADDR + ":" + SERV_PORT
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	logger.Info("Server started at:" + URL)
	logger.Info("Swagger docs available at: " + URL)
	for _, env := range os.Environ() {
		logger.Info(env)
	}
	if err := router.Run(SERV_ADDR + ":" + SERV_PORT); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}

}

func StartKafkaWorker_Create() error {
	maxRetryStr := os.Getenv("KAFKA_RETRY")
	intervalStr := os.Getenv("KAFKA_INTERVAL")
	maxRetryInt, err_ := strconv.Atoi(maxRetryStr)
	if err_ != nil {
		fmt.Println("Invalid KAFKA_INTERVAL, using default of 10 seconds")
		maxRetryInt = 10 // default fallback
	}
	intervalInt, err_ := strconv.Atoi(intervalStr)
	if err_ != nil {
		fmt.Println("Invalid KAFKA_INTERVAL, using default of 10 seconds")
		intervalInt = 10 // default fallback
	}
	var (
		brokers       = []string{os.Getenv("KAFKA_SERVER")}
		topic         = os.Getenv("KAFKA_TOPIC_WO_CREATE")
		maxRetry      = maxRetryInt
		retryInterval = time.Duration(intervalInt) * time.Second
	)

	var consumer sarama.Consumer
	var err error

	for attempt := 1; attempt <= maxRetry; attempt++ {
		log.Printf("Attempt %d to connect to "+topic+" brokers: %v", attempt, brokers)
		consumer, err = sarama.NewConsumer(brokers, nil)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to "+topic+": %v", err)
		time.Sleep(retryInterval)
	}

	if err != nil {
		return fmt.Errorf("could not connect to "+topic+" after %d attempts: %w", maxRetry, err)
	}

	defer func() {
		if err := consumer.Close(); err != nil {
			log.Printf("Error closing "+topic+" consumer: %v", err)
		}
	}()

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Printf("Error starting partition consumer: %v", err)
		return err
	}
	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			log.Printf("Error closing StartKafkaWorker_Create() consumer: %v", err)
		}
	}()

	log.Println("Kafka StartKafkaWorker_Create() started. Listening for messages...")

	for msg := range partitionConsumer.Messages() {
		go handleMessage_WO_Create(msg.Value)
	}

	return nil
}

func handleMessage_WO_Create(message []byte) {
	log.Printf("Raw message: %s", string(message))
	var wo model.WorkOrder
	if err := json.Unmarshal(message, &wo); err != nil {
		log.Printf("Error unmarshalling message: %v", err)
		return
	}

	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		log.Printf("DB connection is nil")
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	username := os.Getenv("INTEGRATION_USR")
	orgId := os.Getenv("INTEGRATION_ORG_ID")

	if err := handler.IntegrateCreateCaseFromWorkOrder(ctx, conn, wo, username, orgId); err != nil {
		log.Printf("Error creating case from WorkOrder: %v", err)
	}
}

func StartKafkaWorker_Update() error {
	maxRetryStr := os.Getenv("KAFKA_RETRY")
	intervalStr := os.Getenv("KAFKA_INTERVAL")
	maxRetryInt, err_ := strconv.Atoi(maxRetryStr)
	if err_ != nil {
		fmt.Println("Invalid KAFKA_INTERVAL, using default of 10 seconds")
		maxRetryInt = 10 // default fallback
	}
	intervalInt, err_ := strconv.Atoi(intervalStr)
	if err_ != nil {
		fmt.Println("Invalid KAFKA_INTERVAL, using default of 10 seconds")
		intervalInt = 10 // default fallback
	}
	var (
		brokers       = []string{os.Getenv("KAFKA_SERVER")}
		topic         = os.Getenv("KAFKA_TOPIC_WO_UPDATE")
		maxRetry      = maxRetryInt
		retryInterval = time.Duration(intervalInt) * time.Second
	)

	var consumer sarama.Consumer
	var err error

	for attempt := 1; attempt <= maxRetry; attempt++ {
		log.Printf("Attempt %d to connect to "+topic+" brokers: %v", attempt, brokers)
		consumer, err = sarama.NewConsumer(brokers, nil)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to "+topic+": %v", err)
		time.Sleep(retryInterval)
	}

	if err != nil {
		return fmt.Errorf("could not connect to "+topic+" after %d attempts: %w", maxRetry, err)
	}

	defer func() {
		if err := consumer.Close(); err != nil {
			log.Printf("Error closing "+topic+" consumer: %v", err)
		}
	}()

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Printf("Error starting partition consumer: %v", err)
		return err
	}
	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			log.Printf("Error closing StartKafkaWorker_Update() consumer: %v", err)
		}
	}()

	log.Println("Kafka StartKafkaWorker_Update() started. Listening for messages...")

	for msg := range partitionConsumer.Messages() {
		go handleMessage_WO_Update(&gin.Context{}, msg.Value)
	}

	return nil
}

func handleMessage_WO_Update(c *gin.Context, message []byte) {
	var wo model.WorkOrder
	if err := json.Unmarshal(message, &wo); err != nil {
		log.Printf("Error unmarshalling message: %v", err)
		return
	}

	conn, ctx, cancel := config.ConnectDB()
	if conn == nil {
		log.Printf("DB connection is nil")
		return
	}
	defer cancel()
	defer conn.Close(ctx)

	log.Print("==handleMessage_WO_Update==")
	log.Print(wo)

	//Workorder Status
	// •	NEW
	// •	ASSIGNED
	// •	ACKNOWLEDGE
	// •	INPROGRESS
	// •	DONE
	// •	ONHOLD
	// •	CANCEL

	if wo.Status == "NEW" || wo.Status == "ASSIGNED" {
		return
	}
	log.Print("====1===")
	username := os.Getenv("INTEGRATION_USR")
	orgId := os.Getenv("INTEGRATION_ORG_ID")
	c.Set("username", username)
	c.Set("orgId", orgId)

	if err := handler.IntegrateUpdateCaseFromWorkOrder(c, conn, wo, username, orgId); err != nil {
		log.Printf("Error creating case from WorkOrder: %v", err)
	}
}
