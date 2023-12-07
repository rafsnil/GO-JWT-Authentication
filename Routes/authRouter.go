package routes

import (
	"github.com/gin-gonic/gin"
	controllers "github.com/rafsnil/Go-JWT-Authentication/Controllers"
)

func AuthRoutes(incomingRoutes *gin.Engine) {

	//Handling the incoming routes
	incomingRoutes.POST("/users/signup", controllers.ExecuteSignUp())
	incomingRoutes.POST("/users/login", controllers.ExecuteLogin())
}
