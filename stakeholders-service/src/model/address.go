package model

type GeoPoint struct {
	Type        string     `bson:"type" json:"type"`
	Coordinates [2]float64 `bson:"coordinates" json:"coordinates"` // [longitude, latitude]
}

type Address struct {
	Street     string   `bson:"street" json:"street"`
	City       string   `bson:"city" json:"city"`
	State      string   `bson:"state" json:"state"`
	PostalCode string   `bson:"postal_code" json:"postal_code"`
	Country    string   `bson:"country" json:"country"`
	Location   GeoPoint `bson:"location,omitempty" json:"location,omitempty"` // GeoJSON format
}
