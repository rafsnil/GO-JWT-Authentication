package routes

import (
	"github.com/gin-gonic/gin"
	controllers "github.com/rafsnil/Go-JWT-Authentication/Controllers"
	middleware "github.com/rafsnil/Go-JWT-Authentication/Middlewares"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	/*Using the Authenticate Middleware to see if the user is
	permitted to access the routes*/
	incomingRoutes.Use(middleware.Authenticate())

	//Handling the request to the routes through Gin
	//Gin requries the function to be called rather than just passed
	incomingRoutes.GET("/users", controllers.GetAllUsers())
	incomingRoutes.GET("/users/:user_id", controllers.GetUserByID())
}
