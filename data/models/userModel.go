package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	Id       primitive.ObjectID `json:"id" bson:"_id"`
	Username string             `json:"username" binding:"required"`
	Email    string             `json:"email" binding:"required,email"`
	Password string             `json:"password" binding:"required"`
}
