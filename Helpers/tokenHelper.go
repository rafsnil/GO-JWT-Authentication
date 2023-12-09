package helpers

import (
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

type SignedDetails struct {
	Email      string
	First_Name string
	Last_Name  string
	Uid        string
	User_Type  string
	jwt.StandardClaims
}

//var userCollection *mongo.Collection = database.OpenCollection(database.Client, "User")

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
	token, err := jwt.NewWithClaims(jwt.SigningMethodES256, claims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		log.Panic(err)
		return
	}

	//Creating a refresh token
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodES256, refreshClaims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		log.Panic(err)
		return
	}

	return token, refreshToken, err

}
