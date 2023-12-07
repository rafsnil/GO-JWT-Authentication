package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type user struct {
	Id            primitive.ObjectID `bson: "_id"`
	First_Name    *string            `json:"first_name" validate:"required, min=2, max=100"`
	Last_Name     *string            `json:"last_name" validate:"required, min=2, max=100"`
	Password      *string            `json:"password" validate:"required, min=6"`
	Email         *string            `json:"email" validate:"email, required"`
	Phone         *string            `json:"phone" validate:"required"`
	Token         *string            `json:"token"`
	User_Type     *string            `json:"user_type" validate:"required, eq=ADMIN|eq=USER"`
	Refresh_Token time.Time          `json:"refresh_token"`
	Created_At    time.Time          `json:"created_at"`
	User_Id       string             `json:"user_id"`
}