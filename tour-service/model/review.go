package model

import (
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Review struct {
    ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    TourID     primitive.ObjectID `bson:"tourId" json:"tourId"`
    AuthorID   string             `bson:"authorId" json:"authorId"`
    AuthorName string             `bson:"authorName,omitempty" json:"authorName,omitempty"`
    Rating     int                `bson:"rating" json:"rating"` // 1-5
    Comment    string             `bson:"comment,omitempty" json:"comment,omitempty"`
    Images     []string           `bson:"images,omitempty" json:"images,omitempty"`
    VisitedAt  *time.Time         `bson:"visitedAt,omitempty" json:"visitedAt,omitempty"` // when tourist visited
    CreatedAt  time.Time          `bson:"createdAt" json:"createdAt"`                     // when review posted
}
