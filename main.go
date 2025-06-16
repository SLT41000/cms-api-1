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
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Use(cors.Default())
	router.Use(gin.Recovery())
	auth := router.Group("/api/v1/AuthAPI")
	{

		auth.POST("/token", handler.GetToken)
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
	logger := config.GetLog()
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	logger.Info("Server started at: http://localhost:8080")
	logger.Info("Swagger docs available at: http://localhost:8080/swagger/index.html")
	if err := router.Run(":8080"); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}

}
