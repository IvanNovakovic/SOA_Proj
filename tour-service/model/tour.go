package model

import (
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Tour struct {
    ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    AuthorID    string             `bson:"authorId" json:"authorId"`
    Name        string             `bson:"name" json:"name"`
    Description string             `bson:"description" json:"description"`
    Difficulty  string             `bson:"difficulty" json:"difficulty"`
    Tags        []string           `bson:"tags" json:"tags"`
    Status      string             `bson:"status" json:"status"`
    Price       float64            `bson:"price" json:"price"`
    CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
}