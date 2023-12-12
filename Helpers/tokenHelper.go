package helpers

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	database "github.com/rafsnil/Go-JWT-Authentication/Database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct {
	Email      string
	First_Name string
	Last_Name  string
	Uid        string
	User_Type  string
	jwt.StandardClaims
}

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "User")

var SECRET_KEY string = os.Getenv("JWT_SECRET_KEY")

func GenerateAllTokens(userEmail string, userFname string, userLname string, userType string, userId string) (signedToken string, signedRefreshToken string, err error) {
	claims := &SignedDetails{
		Email:      userEmail,
		First_Name: userFname,
		Last_Name:  userLname,
		Uid:        userId,
		User_Type:  userType,
		StandardClaims: jwt.StandardClaims{
			/*
				This gives the current local time
				currentTime := time.Now().Local()
				Add() adds local time and 24hrs which is the expiration time
				expirationTime := currentTime.Add(time.Hour * time.Duration(24))
				Unix() converts the whole time to a Unix timestamp
				unixExpiration := expirationTime.Unix()
			*/
			//ALL OF THE ABOVE IN ONE LINE
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}

	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}

	//Creating a Token
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		log.Panic(err)
		return
	}

	//Creating a refresh token
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		log.Panic(err)
		return
	}

	return token, refreshToken, err

}

func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {

	//Checking the clains with the token
	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)
	// fmt.Println("Success in parsing the claims")
	if err != nil {
		msg = err.Error()
		return
	}
	//Claim is basically all info that the user has in his token
	//Converting the tokenClaims to signedDetails using type assertion
	claim, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = "The Token is Invalid!"
		return
	}
	// fmt.Println("Success in converting the claims")
	//Checking for the token expiration
	if claim.ExpiresAt < time.Now().Unix() {
		msg = "Token is Expired"
		return
	}
	// fmt.Println("Success in checking expiration date of the claims")
	return claim, msg
}

func UpdateAllTokens(signedToken string, signedRefreshToken string, userId string) {
	//Creating a context to interact with the MongoDB
	cntxt, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	/*
		Primitive.D is a type alias for []bson.E in the Go programming
		language. This means that it represents an ordered slice of bson.E
		elements. bson.E, in turn, represents a single key-value pair in the BSON format, which is the native data format used by MongoDB.
	*/
	var updateObj primitive.D

	/*
		The method below used to work in older version of GO, but
		now is depreciated.
		updateObj = append(updateObj, bson.E{"token", signedToken})
		Instead do this to get the exact same result as above
	*/
	//bson.E{} is used here, as we are trying to maitain the oder of the info in the doc
	updateObj = append(updateObj, bson.E{Key: "token", Value: signedToken})
	updateObj = append(updateObj, bson.E{Key: "refresh_token", Value: signedRefreshToken})

	updateAt, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{Key: "updated_at", Value: updateAt})

	//Here Upsert means both updating and inserting
	upsert := true

	//Setting filter to look for the doc that needs to be updated
	//Typically bson.M is used when the order of the filter(the info inside it) does not matter.
	//If the order massters, then use bson.D for filter
	filter := bson.M{"user_id": userId}

	opt := options.UpdateOptions{
		Upsert: &upsert,
	}
	/*
		Using "$set" in bson.D specifies that the provided values
		should update the corresponding fields in the existing document.
	*/
	//bson.D ensures that the order of the updated doc is maintained
	update := bson.D{
		{Key: "$set", Value: updateObj},
	}
	_, err := userCollection.UpdateOne(
		cntxt,
		filter,
		update,
		&opt,
	)

	if err != nil {
		log.Panic(err)
		return
	}

}
