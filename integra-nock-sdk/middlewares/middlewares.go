package middlewares

import (
	"log"
	"strings"

	"integra-nock-sdk/helpers/token"

	"github.com/gin-gonic/gin"
	"integra-nock-sdk/utils"
)

//middleware for token validation
func JwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		receivedToken := c.Request.Header.Get("Authorization")
		if len(strings.Split(receivedToken, " ")) == 2 {
			receivedToken = strings.Split(receivedToken, " ")[1]
		}
		log.Println("received token inside client middleware from client", receivedToken)
		err := token.TokenValid(receivedToken)
		if err != nil {
			log.Println("err in token ", err)
			c.IndentedJSON(401, utils.Response{
				Status:  401,
				Message: "Find unauthorized token in middleware check at client side",
				Data:    nil,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}


	
