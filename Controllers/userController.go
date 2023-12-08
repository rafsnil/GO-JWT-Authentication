package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	database "github.com/rafsnil/Go-JWT-Authentication/Database"
	helpers "github.com/rafsnil/Go-JWT-Authentication/Helpers"
	models "github.com/rafsnil/Go-JWT-Authentication/Models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "Users")
var validate = validator.New()

func HashPassword()

func VerifyPassword()

// SIGN UP HANDLER
// 2
func ExecuteSignUp() gin.HandlerFunc {
	return func(requestCntxt *gin.Context) {
		var cntxt, cancel = context.WithTimeout(context.Background(), 10*time.Second)

		var user models.User

		//Mapping the json data in the request to the user struct
		err := requestCntxt.BindJSON(&user)
		if err != nil {
			requestCntxt.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			defer cancel()
			return
		}

		/*Validating the user info according to the validation
		rules mentioned in userModels.go*/
		validationErr := validate.Struct(user)
		if validationErr != nil {
			requestCntxt.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			defer cancel()
			return
		}

		/*
			After successfull validation, checking if the email and phone
			number already exist in the database
		*/
		count, err := userCollection.CountDocuments(cntxt, bson.M{"email": user.Email})
		defer cancel()
		if err != nil {
			log.Panic(err)
			requestCntxt.JSON(http.StatusInternalServerError, gin.H{"Error": "Error occured while checking the email."})
		}

		count, err = userCollection.CountDocuments(cntxt, bson.M{"phone": user.Phone})
		defer cancel()
		if err != nil {
			log.Panic(err)
			requestCntxt.JSON(http.StatusInternalServerError, gin.H{"Error": "Error occured while checking the phone number"})
		}

		if count > 0 {
			requestCntxt.JSON(http.StatusInternalServerError, gin.H{"Error": "This Email or Phone Number Already Exists!"})
		}

	}
}

func ExecuteLogin()

func GetAllUsers()

// GET USER BY ID HANDLER
// This can be accessed by only ADMINS
// 1
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

		cntxt, cancelContext := context.WithTimeout(context.Background(), 10*time.Second)
		var user models.User

		//Looking for the doc which has the same user_id
		//Then populating the "user" with the info of that doc
		userDoc := userCollection.FindOne(cntxt, bson.M{"user_id": userId})
		defer cancelContext()
		// userDoc := userCollection.FindOne(requestCntxt, bson.M{"user_id": userId})
		err1 := userDoc.Decode(&user)

		if err1 != nil {
			requestCntxt.JSON(http.StatusInternalServerError, gin.H{"Error": err1.Error()})
		}

		requestCntxt.JSON(http.StatusOK, user)

	}
}
