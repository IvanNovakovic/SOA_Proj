package model

import (
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

type KeyPoint struct {
    ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    TourID      primitive.ObjectID `bson:"tourId" json:"tourId"`
    Name        string             `bson:"name" json:"name"`
    Description string             `bson:"description,omitempty" json:"description,omitempty"`
    ImageURL    string             `bson:"imageUrl,omitempty" json:"imageUrl,omitempty"`
    Latitude    float64            `bson:"latitude" json:"latitude"`
    Longitude   float64            `bson:"longitude" json:"longitude"`
    CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
}
