package main

import (

	//admin route file path

	//client packages
	clientRouter "integra-nock-sdk/client/client_route"

	"integra-nock-sdk/middlewares"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/spacemonkeygo/openssl"
)

func main() {
	router := gin.Default()
	router.Use(CORSMiddleware())

	router.POST("/user/login", clientRouter.Login)
	router.GET("/user/test1", clientRouter.TestApi)
	router.Use(middlewares.JwtAuthMiddleware())
	// router.Use(CORSMiddleware())
	router.GET("/user/home", clientRouter.UserDashboard)
	router.GET("/user/test2", clientRouter.CheckApi)
	// router.GET("/user/home", middlewares.JwtAuthMiddleware(), clientRouter.UserDashboard)

	// ============================= CLIENT CHAINCODE PART =====================================

	router.POST("/chaincode/install", clientRouter.ChaincodeInstall)

	//get chaincode list
	router.GET("/chaincode/list", clientRouter.ChaincodeList)

	//get route to get chaincode info by id
	router.GET("/chaincode/checkforupdates", clientRouter.ChaincodeUpdateCheck)

	//api to download and install updates of chaincode
	router.POST("/chaincode/update/:chaincodeid", clientRouter.InstallUpdate)

	// ============================== CLIENT ORGANIZATION PART ===========================================
	router.GET("/organization", clientRouter.ListOrganizations)

	//route for sign newly added org
	//orgid is id of organization to be signed
	router.POST("/organization/sign/:orgid", clientRouter.SignOrganization)

	router.Run("localhost:4000")
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
