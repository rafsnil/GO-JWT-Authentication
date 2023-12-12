package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
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
	fmt.Println("Pass Given by user: " + userPasswrod)
	fmt.Println("Pass Stored in DB: " + providedPassword)
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPasswrod))
	fmt.Println(err)
	check := true
	msg := ""

	if err != nil {
		msg = "Password is Incorrect!"
		check = false
	}

	return check, msg

}

// SIGN UP HANDLER
// 2
func ExecuteSignUp() gin.HandlerFunc {
	return func(requestCntxt *gin.Context) {
		var cntxt, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User

		//Mapping the json data in the request to the user struct
		err := requestCntxt.BindJSON(&user)
		if err != nil {
			requestCntxt.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			// defer cancel()
			return
		}

		// err1 := validate.StructPartial(user, "First_Name")
		// if err1 != nil {
		// 	fmt.Println("Could not validate first name")
		// 	log.Panic(err1)
		// }
		// err2 := validate.StructPartial(user, "Last_Name")
		// if err2 != nil {
		// 	fmt.Println("Could not validate last name")
		// 	log.Panic(err2)
		// }
		// err3 := validate.StructPartial(user, "Password")
		// if err3 != nil {
		// 	fmt.Println("Could not validate password")
		// 	log.Panic(err3)
		// }
		// err4 := validate.StructPartial(user, "Email")
		// if err4 != nil {
		// 	fmt.Println("Could not validate email")
		// 	log.Panic(err4)
		// }
		// err5 := validate.StructPartial(user, "Phone")
		// if err5 != nil {
		// 	fmt.Println("Could not validate phone")
		// 	log.Panic(err5)
		// }
		// err6 := validate.StructPartial(user, "User_Type")
		// if err6 != nil {
		// 	fmt.Println("Could not validate user_type")
		// 	log.Panic(err6)
		//}

		/*Validating the user info according to the validation
		rules mentioned in userModels.go*/

		validationErr := validate.Struct(user)

		if validationErr != nil {
			requestCntxt.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			// defer cancel()
			return
		}
		//fmt.Println("Running till here: After validation")
		/*
			After successfull validation, FIRST OF ALL hash and store
			the password so that it is in plaintext form for the minimal
			time in the systen checking if the email and phonenumber already exist in the database
		*/

		password := HashPassword(*user.Password)
		user.Password = &password

		count, err := userCollection.CountDocuments(cntxt, bson.M{"email": user.Email})
		if err != nil {
			// defer cancel()
			log.Panic(err)
			requestCntxt.JSON(http.StatusInternalServerError, gin.H{"Error": "Error occured while checking the email."})
			return
		}
		if count > 0 {
			requestCntxt.JSON(http.StatusInternalServerError, gin.H{"Error": "This Email Already Exists!"})
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
			requestCntxt.JSON(http.StatusInternalServerError, gin.H{"Error": "This Phone Number Already Exists!"})
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
		token, refreshToken, err := helpers.GenerateAllTokens(*user.Email, *user.First_Name, *user.Last_Name, *user.User_Type, user.User_Id)
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
			requestCntxt.JSON(http.StatusInternalServerError, gin.H{"Error": "User item was not created."})
			return
		}

		// defer cancel()
		requestCntxt.JSON(http.StatusOK, resultInsertionNumber)
	}
}

func ExecuteLogin() gin.HandlerFunc {
	return func(requestCntxt *gin.Context) {
		var cntxt, cancel = context.WithTimeout(context.Background(), 100*time.Second)
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
		if !passIsValid {
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
		token, refreshtoken, err := helpers.GenerateAllTokens(*foundUser.Email, *foundUser.First_Name, *foundUser.Last_Name, *foundUser.User_Type, foundUser.User_Id)
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

// GET USERS HANDLER
// GetAllUsers() cann only be used by admin
func GetAllUsers() gin.HandlerFunc {
	return func(requestCntxt *gin.Context) {
		//Checking if the user is admin
		err := helpers.CheckUserType(requestCntxt, "ADMIN")
		if err != nil {
			requestCntxt.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
		var cntxt, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		/*
			The purpose of recordPerPage in this function is to determine the number of users
			to retrieve per page in the user list. It is used to handle pagination, which
			allows users to browse through a large list of users without having to load all
			of them at once.
		*/

		//Specifying how many users will be listed in one page if recordPerPage is missing in the url
		recordPerPage, err := strconv.Atoi(requestCntxt.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}
		// fmt.Println("Ran after recordPerPage")
		//Specifying how many pages will be created with the records if page is missing in the url
		page, err := strconv.Atoi(requestCntxt.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}
		// fmt.Println("Ran after Page")

		startIndex := (page - 1) * recordPerPage
		parsedStartIndex, err := strconv.Atoi(requestCntxt.Query("startIndex"))
		if err == nil {
			startIndex = parsedStartIndex
		}
		// fmt.Println("Ran after startIndex")

		//Setting up a match stage with no matching criteria
		//allowing all docs to pass through without filering
		matchStage := bson.D{{Key: "$match", Value: bson.D{{}}}}

		//Settin up the group stage where all docs will be grouped together as id is null
		//Then total number of focuments are counted and then
		//all the docs are pushed into data
		groupStage := bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{{Key: "_id", Value: "null"}}},
			{Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}}}}}

		/*
			$project tells that this is the project stage.
			Id field is excluded
			total_count is included
			new field "user_items" is created which will hold a subset of the data
			It is called $data because it was name "data" in the groupStage
			$slice is used to get recordPerPage number of data.
			If startIndex is provided, slicing will start from startIndex. Otherwise, slicing will start from the beginning of the data array.
		*/
		projectStage := bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "total_count", Value: 1},
				{Key: "user_items", Value: bson.D{
					{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}}}},
			}}}

		//Using aggregate which returns a cursor, to get my required data in the required style
		result, err := userCollection.Aggregate(cntxt, mongo.Pipeline{
			matchStage,
			groupStage,
			projectStage,
		})

		if err != nil {
			requestCntxt.JSON(http.StatusInternalServerError, gin.H{"Error": "Error occured while listing user items"})
		}
		// fmt.Println("Ran after aggregate")
		// fmt.Println("\n\n\n\nPrinting the curcsor:", result, "\n\n\n\n")

		var allUsers []bson.M
		//Decoding all the docs into allUsers which is a bson.M slice.
		err = result.All(cntxt, &allUsers)
		// fmt.Println("\n\n\n\nPrinting the stuff:", allUsers, "\n\n\n\n")
		if err != nil {
			log.Fatal(err)
			// fmt.Println("\n\nError here:\n\n", err, "\n\n")
		}

		requestCntxt.JSON(http.StatusOK, allUsers[0])
	}
}

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

		cntxt, cancelContext := context.WithTimeout(context.Background(), 100*time.Second)
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
