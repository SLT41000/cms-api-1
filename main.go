// @title Customers API
// @version 1.0
// @description This is the Customers API server.
// @termsOfService http://somewhere.com/
// @BasePath /
// @contact.name API Support
// @contact.url http://somewhere.com/support
// @contact.email support@somewhere.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @schemes http

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
	dbConn, _, _ := config.ConnectDB()
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Use(cors.Default())
	router.Use(gin.Recovery())
	auth := router.Group("/api/v1/authAPI")
	{

		authHandler := handler.AuthHandler{DB: dbConn}
		auth.POST("/token", authHandler.GetToken)
	}
	// customers := router.Group("/api/v1/customers")
	// customers.Use(handler.ProtectedHandler)
	// {

	// 	customerHandler := handler.CustomerHandler{DB: dbConn}
	// 	customers.GET("/:id", customerHandler.GetCustomer)
	// 	customers.GET("", customerHandler.ListCustomers)
	// 	customers.POST("", customerHandler.CreateCustomer)
	// 	customers.DELETE("/:id", customerHandler.DeleteCustomer)
	// 	customers.PATCH("/:id", customerHandler.UpdateCustomer)
	// }

	cases := router.Group("/api/v1/cases")
	{
		cases.Use(handler.ProtectedHandler)
		caseHandler := handler.CaseHandler{DB: dbConn}
		cases.GET("", caseHandler.ListCase)
		cases.GET("/search", caseHandler.SearchCase)
		cases.GET("/:id", caseHandler.SearchCaseById)
		cases.PATCH("/:id", caseHandler.UpdateCase)
		cases.DELETE("/:id", caseHandler.DeleteCase)
		cases.PATCH("/close/:id", caseHandler.UpdateCaseClose)
		cases.GET("/detail/:id", caseHandler.SearchCaseByCaseCode)
		cases.POST("", caseHandler.CreateCase)
	}
	logger := config.GetLog()
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	logger.Info("Server started at: http://localhost:8080")
	logger.Info("Swagger docs available at: http://localhost:8080/swagger/index.html")
	if err := router.Run(":8080"); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}

}
