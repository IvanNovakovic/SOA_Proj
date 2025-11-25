package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TransportDuration struct {
	Walking int `bson:"walking" json:"walking"` // minutes
	Biking  int `bson:"biking" json:"biking"`   // minutes
	Driving int `bson:"driving" json:"driving"` // minutes
}

type Tour struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AuthorID    string             `bson:"authorId" json:"authorId"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Difficulty  string             `bson:"difficulty" json:"difficulty"`
	Tags        []string           `bson:"tags" json:"tags"`
	Status      string             `bson:"status" json:"status"` // draft, published, archived
	Price       float64            `bson:"price" json:"price"`
	Distance    float64            `bson:"distance" json:"distance"` // kilometers
	Durations   TransportDuration  `bson:"durations" json:"durations"`
	PublishedAt *time.Time         `bson:"publishedAt,omitempty" json:"publishedAt,omitempty"`
	ArchivedAt  *time.Time         `bson:"archivedAt,omitempty" json:"archivedAt,omitempty"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
}
