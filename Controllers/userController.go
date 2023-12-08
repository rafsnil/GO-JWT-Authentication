package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	database "github.com/rafsnil/Go-JWT-Authentication/Database"
	helpers "github.com/rafsnil/Go-JWT-Authentication/Helpers"
	"go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "User")
var validate = validator.New()

func HashPassword()

func VerifyPassword()

func ExecuteSignUp()

func ExecuteLogin()

func GetAllUsers()

// GET USER BY ID HANDLER
// This can be accessed by only ADMINS
func GetUserByID() gin.HandlerFunc {
	return func(requestCntxt *gin.Context) {
		//Getting the userId from the url parameter
		//Do NOT USE PARAMS, it is different function.
		userId := requestCntxt.Param("user_id")

		//Checking if the user type, if its an admin
		err := helpers.MatchUserTypeToUid(requestCntxt, userId)
		if err != nil {
			requestCntxt.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
	}
}
