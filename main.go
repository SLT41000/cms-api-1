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

	store := memory.NewStore()
	instance := limiter.New(store, rate)
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	// router.Use(cors.Default())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://example.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.Use(gin.Recovery())
	router.Use(ginlimiter.NewMiddleware(instance))
	auth := router.Group("/api/v1/AuthAPI")
	{
		auth.GET("/token", handler.GetToken)
		auth.GET("/login", handler.UserLogin)
		auth.POST("/add", handler.UserAdd)
	}
	v1 := router.Group("/api/v1")
	{
		// 	v1.Use(handler.ProtectedHandler)
		v1.GET("/forms/:id", handler.GetForm)
		v1.GET("/workflows/:id", handler.GetWorkFlow)
		v1.GET("/casetypes", handler.ListCaseType)
		v1.GET("/casesubtypes", handler.ListCaseSubType)
		v1.GET("/notes/:id", handler.ListTransactionNote)
		v1.POST("/notes", handler.CreateTransactionNote)
		v1.GET("/departments", handler.GetDepartment)
		v1.GET("/commands", handler.GetCommand)
		v1.GET("/stations", handler.GetStation)
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
		notifications.POST("/new", handler.PostNotificationCustom)
		notifications.GET("/recipient/:username", handler.GetNotificationsByRecipient)
		notifications.GET("/noti/:id", handler.GetNotificationByID)
		notifications.GET("/ws", handler.WebSocketHandler)
		notifications.PUT("/edit/:id", handler.UpdateNotificationByID)
		notifications.DELETE("/delete/:id", handler.DeleteNotificationByID)
	}
	logger := config.GetLog()
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	logger.Info("Server started at: http://localhost:8080")
	logger.Info("Swagger docs available at: http://localhost:8080/swagger/index.html")
	if err := router.Run(":8080"); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}

}
