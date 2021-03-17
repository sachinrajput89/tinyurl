/*
Package types : Package for defining MongoDB Schema
*/
package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

//Urls Struct to define
type Urls struct {
	ID        primitive.ObjectID `bson:"_id"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
	Tinyurl   string             `bson:"tinyurl"`
	Longurl   string             `bson:"longurl"`
}

// Response : Output json response
type Response struct {
	URL      string    `json:"url`
	ExpireOn time.Time `json:"expireOn`
}

// ResponseLongURL : to define ResponseLongURL
type ResponseLongURL struct {
	URL string `json:"url`
}
