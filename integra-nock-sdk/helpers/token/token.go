package token

import (
	"fmt"
	"log"
	"os"

	jwt "github.com/dgrijalva/jwt-go"
)

//check token is valid or not
func TokenValid(receivedToken string) error {
	log.Println("Inside token valid function of client side", receivedToken)
	// tokenString := ExtractToken(receivedToken)
	tokenString := receivedToken
	log.Println("tokenString in client token valid check: ", tokenString)

	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("API_SECRET")), nil
	})
	if err != nil {
		return err
	}
	return nil
}

//extract particular token
// func ExtractToken(bearerToken string) string {
// 	log.Println("inside extract token fx of client side", bearerToken)

// 	// token := c.Query("token")
// 	// log.Println("token in extract fx: ", token)
// 	// if token != "" {
// 	// 	return token
// 	// }
// 	// bearerToken := c.Request.Header.Get("Authorization")
// 	if len(strings.Split(bearerToken, " ")) == 2 {
// 		return strings.Split(bearerToken, " ")[1]
// 	}
// 	return ""
// }
