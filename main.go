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

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Use(cors.Default())
	router.Use(gin.Recovery())
	auth := router.Group("/api/v1/AuthAPI")
	{

		auth.GET("/token", handler.GetToken)
		auth.GET("/login", handler.UserLogin)
		auth.POST("/add", handler.UserAdd)
	}
	forms := router.Group("/api/v1/forms")
	{

		forms.GET("/:id", handler.GetForm)
	}
	workflows := router.Group("/api/v1/workflows")
	{

		workflows.GET("/:id", handler.GetWorkFlow)
	}
	casetypes := router.Group("/api/v1/casetypes")
	{

		casetypes.GET("", handler.ListCaseType)
	}

	cases := router.Group("/api/v1/cases")
	{
		cases.Use(handler.ProtectedHandler)
		cases.GET("", handler.ListCase)
		cases.GET("/search", handler.SearchCase)
		cases.GET("/:id", handler.SearchCaseById)
		cases.PATCH("/:id", handler.UpdateCase)
		cases.DELETE("/:id", handler.DeleteCase)
		cases.PATCH("/close/:id", handler.UpdateCaseClose)
		cases.GET("/detail/:id", handler.SearchCaseByCaseCode)
		cases.POST("", handler.CreateCase)
	}
	trans := router.Group("/api/v1/trans")
	{
		// trans.Use(handler.ProtectedHandler)
		trans.GET("", handler.ListTransaction)
		trans.GET("/:id", handler.SearchTransaction)
		trans.POST("", handler.CreateTransaction)
		trans.PATCH("/:id", handler.UpdateTransaction)
		trans.DELETE("/:id", handler.DeleteTransaction)
	}
	notes := router.Group("/api/v1/notes")
	{
		// trans.Use(handler.ProtectedHandler)
		notes.GET("/:id", handler.ListTransactionNote)
		notes.POST("", handler.CreateTransactionNote)
	}
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
