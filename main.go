package main

import (
	"os"

	"github.com/gin-gonic/gin"
	routes "github.com/rafsnil/Go-JWT-Authentication/Routes"
)

func main() {
	//Getting the port from .env file
	port := os.Getenv("PORT")

	if port == "" {
		port = "8000"
	}

	//gin.New() returns a new engine instance without a middleware
	router := gin.New()
	//gin.Use() attaches a global middleware to the router
	router.Use(gin.Logger())

	/*
		Similar To:
		http.HandleFunc("/api-1", func (w http.ResponseWriter, r *http.Request)){
			fmt.Fprintln(w, "Access Granted for API-1")
		}
		⬇⬇⬇⬇⬇⬇⬇⬇⬇
	*/

	//Passing the gin engine (incoming request) to the AuthRoutes/UserRoutes
	routes.AuthRoutes(router)
	routes.UserRoutes(router)

	// c.JSON sets the content type to json automatically
	router.GET("/api-1", func(c *gin.Context) {
		c.JSON(200, gin.H{"Success": "Access Grandted for API-1"})
	})

	router.GET("/api-2", func(c *gin.Context) {
		c.JSON(200, gin.H{"Success": "Access Granted for API-2"})
	})

	//Similar to http.ListenAndServe(":8080",router)
	router.Run(":" + port)

}
