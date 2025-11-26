package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChangeEvent struct {
	ID            bson.M `bson:"_id"`
	OperationType string `bson:"operationType"`
	FullDocument  Chat   `bson:"fullDocument"`
	Ns            bson.M `bson:"ns"`
}

type Chat struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Room      string             `bson:"room" json:"room"`
	User      string             `bson:"user" json:"user"`
	Message   string             `bson:"message" json:"message"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
