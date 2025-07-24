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
	"mainPackage/config"
	_ "mainPackage/docs"
	"mainPackage/handler"
	"time"

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
	store := memory.NewStore()
	instance := limiter.New(store, rate)
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	// router.Use(cors.Default())
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173",             // dev
			"https://cms-sigma-woad.vercel.app", // production frontend
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE , PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.Use(gin.Recovery())
	router.Use(ginlimiter.NewMiddleware(instance))
	auth := router.Group("/api/v1/auth")
	{
		auth.GET("/login", handler.UserLogin)
		auth.POST("/add", handler.UserAddAuth)
		auth.POST("/refresh", handler.RefreshToken)
	}
	v1 := router.Group("/api/v1")
	{
		v1.Use(handler.ProtectedHandler)
		v1.GET("/forms", handler.GetForm)
		v1.POST("/forms", handler.FormInsert)
		v1.PATCH("/forms/:uuid", handler.FormUpdate)
		v1.PATCH("/forms/active", handler.FormActive)
		v1.PATCH("/forms/lock", handler.FormLock)
		v1.PATCH("/forms/publish", handler.FormPublish)
		v1.GET("/workflows/:id", handler.GetWorkFlow)

		v1.GET("/casetypes", handler.ListCaseType)
		v1.POST("/casetypes/add", handler.InsertCaseType)
		v1.PATCH("/casetypes/:id", handler.UpdateCaseType)
		v1.DELETE("/casetypes/:id", handler.DeleteCaseType)
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

		v1.GET("/role_permission", handler.GetRolePermission)
		v1.GET("/role_permission/:id", handler.GetRolePermissionbyId)
		v1.POST("/role_permission/add", handler.InsertRolePermission)
		v1.PATCH("/role_permission/:roleId", handler.UpdateRolePermission)
		v1.DELETE("/role_permission/:id", handler.DeleteRolePermission)

		v1.GET("/users", handler.GetUmUserList)
		v1.GET("/users/:id", handler.GetUmUserById)
		v1.POST("/users/add", handler.UserAdd)
		v1.PATCH("/users/:id", handler.UserUpdate)
		v1.DELETE("/users/:id", handler.UserDelete)
		v1.GET("/users/username/:username", handler.GetUmUserByUsername)
		v1.PATCH("/users/username/:username", handler.UserUpdateByUsername)
		v1.GET("/users_with_skills", handler.GetUserWithSkills)
		v1.GET("/users_with_skills/:id", handler.GetUserWithSkillsById)
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
	}
	// cases := router.Group("/api/v1/cases")
	// {
	// 	cases.Use(handler.ProtectedHandler)
	// 	cases.GET("", handler.ListCase)
	// 	cases.GET("/search", handler.SearchCase)
	// 	cases.GET("/:id", handler.SearchCaseById)
	// 	cases.PATCH("/:id", handler.UpdateCase)
	// 	cases.DELETE("/:id", handler.DeleteCase)
	// 	cases.PATCH("/close/:id", handler.UpdateCaseClose)
	// 	cases.GET("/detail/:id", handler.SearchCaseByCaseCode)
	// 	cases.POST("", handler.CreateCase)
	// }

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
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	logger.Info("Server started at: http://localhost:8080")
	logger.Info("Swagger docs available at: http://localhost:8080/swagger/index.html")
	if err := router.Run(":8080"); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}

}
