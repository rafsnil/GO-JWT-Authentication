package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	database "github.com/rafsnil/Go-JWT-Authentication/Database"
	helpers "github.com/rafsnil/Go-JWT-Authentication/Helpers"
	models "github.com/rafsnil/Go-JWT-Authentication/Models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "Users")
var validate = validator.New()

// HASH PASSWORD HANDLER
func HashPassword(userPassword string) string {
	//Hashing the password
	password, err := bcrypt.GenerateFromPassword([]byte(userPassword), 14)
	if err != nil {
		log.Panic(err)
		return ""
	}
	return string(password)
}

// VERIFY PASSWORD HANDLER
func VerifyPassword(userPasswrod string, providedPassword string) (bool, string) {
	//Comparing the hashed password with the given password
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPasswrod))
	check := true
	msg := ""

	if err != nil {
		msg = fmt.Sprintf("Password is Incorrect!")
		check = false
	}

	return check, msg

}

// SIGN UP HANDLER
// 2
func ExecuteSignUp() gin.HandlerFunc {
	return func(requestCntxt *gin.Context) {
		var cntxt, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var user models.User

		//Mapping the json data in the request to the user struct
		err := requestCntxt.BindJSON(&user)
		if err != nil {
			requestCntxt.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			// defer cancel()
			return
		}

		/*Validating the user info according to the validation
		rules mentioned in userModels.go*/
		validationErr := validate.Struct(user)
		if validationErr != nil {
			requestCntxt.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			// defer cancel()
			return
		}

		/*
			After successfull validation, FIRST OF ALL hash and store
			the password so that it is in plaintext form for the minimal
			time in the systen checking if the email and phonenumber already exist in the database
		*/
		count, err := userCollection.CountDocuments(cntxt, bson.M{"email": user.Email})
		if err != nil {
			// defer cancel()
			log.Panic(err)
			requestCntxt.JSON(http.StatusInternalServerError, gin.H{"Error": "Error occured while checking the email."})
			return
		}

		count, err = userCollection.CountDocuments(cntxt, bson.M{"phone": user.Phone})
		if err != nil {
			// defer cancel()
			log.Panic(err)
			requestCntxt.JSON(http.StatusInternalServerError, gin.H{"Error": "Error occured while checking the phone number"})
			return
		}

		if count > 0 {
			requestCntxt.JSON(http.StatusInternalServerError, gin.H{"Error": "This Email or Phone Number Already Exists!"})
			return
		}

		//Assigning the created and updated time
		//time.RFC3339 is a layout of time
		//time.Now() returns the present time, which is then formatted to the time.RFC3339 layout
		//                   time.Parse(    LAYOUT  ,             TIME VALUE         )
		user.Created_At, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if err != nil {
			requestCntxt.JSON(http.StatusInternalServerError, gin.H{"Error": "Error while parsing time for created_at"})
			return
		}

		user.Updated_At, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		if err != nil {
			requestCntxt.JSON(http.StatusInternalServerError, gin.H{"Error": "Error while parsing time for created_at"})
			return
		}

		user.Id = primitive.NewObjectID()

		//Hex() converts the user.Id to string
		user.User_Id = user.Id.Hex()

		/*
			As in the model, email, fname, lname, usertype are all string pointers
			so we use * to dereference the pointer and get the actual string value.
			However, as user.User_Id is not a string pointer, we can directly send the value
			or we can send *&user.User_Id (if we are a bitch about memory waste)
		*/
		token, refreshToken, err := helpers.GenerateAllTokens(*user.Email, *user.First_Name, *user.Last_Name, *user.User_Type, *&user.User_Id)
		// token, refreshToken, _ := helpers.GenerateAllTokens(*user.Email, *user.First_Name, *user.Last_Name, *user.User_Type, user.User_Id)

		if err != nil {
			log.Panic(err)
			requestCntxt.JSON(http.StatusInternalServerError, gin.H{"Error": "Error while generating token"})
			return
		}

		//& is used as both token and refresh token are string pointers in the user model.
		user.Token = &token
		user.Refresh_Token = &refreshToken

		//Unserting the info to the database
		resultInsertionNumber, insertErr := userCollection.InsertOne(cntxt, user)
		if insertErr != nil {
			// defer cancel()
			msg := fmt.Sprintf("User item was not created.")
			requestCntxt.JSON(http.StatusInternalServerError, gin.H{"Error": msg})
			return
		}

		// defer cancel()
		requestCntxt.JSON(http.StatusOK, resultInsertionNumber)
	}
}

func ExecuteLogin() gin.HandlerFunc {
	return func(requestCntxt *gin.Context) {
		var cntxt, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		var user models.User
		var foundUser models.User

		if err := requestCntxt.BindJSON(&user); err != nil {
			requestCntxt.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}

		//userCollection.FindOne() returns nil if the required doc is not found
		requiredUser := userCollection.FindOne(cntxt, bson.M{"email": user.Email})

		err := requiredUser.Decode(&foundUser)
		if err != nil {
			requestCntxt.JSON(http.StatusInternalServerError, gin.H{"Error": "Email or Password is Incorrect"})
			return
		}

		passIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		if passIsValid != true {
			requestCntxt.JSON(http.StatusInternalServerError, gin.H{"Error": msg})
			return
		}

		if foundUser.Email == nil {
			requestCntxt.JSON(http.StatusInternalServerError, gin.H{"Error": "User not found!"})
		}
		/*
			The expression *foundUser.Email accesses the actual email address of a user in Go. The variable foundUser
			holds an object of type User, with an Email field defined as a pointer to a string. The * operator
			dereferences the pointer to retrieve the underlying string value, which is the user's email address.
		*/
		token, refreshtoken, err := helpers.GenerateAllTokens(*foundUser.Email, *foundUser.First_Name, *foundUser.Last_Name, *foundUser.User_Type, *&foundUser.User_Id)
		if err != nil {
			log.Panic(err)
			requestCntxt.JSON(http.StatusInternalServerError, gin.H{"Error": "Error while generating tokens!"})
		}

		helpers.UpdateAllTokens(token, refreshtoken, foundUser.User_Id)
		//The user is looked for again, to ensure the foundUser has the token and refresh token
		err = userCollection.FindOne(cntxt, bson.M{"user_id": foundUser.User_Id}).Decode(&foundUser)
		if err != nil {
			requestCntxt.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		requestCntxt.JSON(http.StatusOK, foundUser)
	}
}

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
			return
		}

		requestCntxt.JSON(http.StatusOK, user)

	}
}
