package helpers

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func CheckUserType(requestCntxt *gin.Context, role string) (err error) {
	userType := requestCntxt.GetString("user_type")
	err = nil
	if userType != role {
		err = errors.New("unauthorized to access this resource❗")
	}
	return err
}

func MatchUserTypeToUid(requestCntxt *gin.Context, userId string) (err error) {
	//Getting the client's user type
	userType := requestCntxt.GetString("user_type")
	//This uid is in the gin.Context through the Authenticate() middleware
	//"uid" is the user id of the person who is sending the request
	uid := requestCntxt.GetString("uid")
	err = nil

	//Checking if the client is user type 'USER' and if the client's uid
	//is not equal to the 'user_id' the client is trying to get

	//If the client is USER and the client's uid is not equal to the userId he is looing for
	//then show an error
	if userType == "USER" && uid != userId {
		err = errors.New("unauthorized to access this resource❗")
		return err
	}

	// err = CheckUserType(requestCntxt, userType)

	return err
}
