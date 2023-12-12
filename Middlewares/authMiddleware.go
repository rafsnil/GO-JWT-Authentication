package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	helpers "github.com/rafsnil/Go-JWT-Authentication/Helpers"
)

func Authenticate() gin.HandlerFunc {
	return func(requestCntxt *gin.Context) {
		//Getting the token from the header
		clientToken := requestCntxt.Request.Header.Get("token")
		if clientToken == "" {
			requestCntxt.JSON(http.StatusInternalServerError, gin.H{"Error": "No Authorization Header Provided!"})
			requestCntxt.Abort()
			return
		}

		//Validating the token
		//If the token is valid, it extracts and stores the claims from the token into the claims variable.
		claims, err := helpers.ValidateToken(clientToken)
		if err != "" {
			requestCntxt.JSON(http.StatusInternalServerError, gin.H{"Error": err})
			requestCntxt.Abort()
			return
		}

		//Setting the values of the field extracted from the token
		requestCntxt.Set("email", claims.Email)
		requestCntxt.Set("first_name", claims.First_Name)
		requestCntxt.Set("last_name", claims.Last_Name)
		requestCntxt.Set("uid", claims.Uid)
		requestCntxt.Set("user_type", claims.User_Type)

		//Going forward with the middleware chain
		requestCntxt.Next()
	}
}
