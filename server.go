/*
Program written for converting any url to a tinyurl.
Database Used : MongoDB
Cache : Redis
Functionalities: Will Create a Tiny Url for for any url
--> A Distinct tiny url will be generated
-->It will get expire if will not be used in last 30 days
-->How It is being Used: Converting url hashed bytes using MD5 and then taking hashed value
checking in db if already exists if not create a new tinyurl using those values.

*/

package main

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
	"tinyurl/db"
	Redis "tinyurl/redis"
	"tinyurl/types"

	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

//Function to store tiny url in DB
func StoreTinyURL(dbURLData types.Urls, longURL string, tinyURL string, dbClient *mongo.Collection, redisClient *redis.Client) {
	dbClient.InsertOne(context.TODO(), dbURLData)
	redisClient.HSet("urls", tinyURL, longURL)
}

//Getting tiny URL
func GetTinyHandler(res http.ResponseWriter, req *http.Request, dbClient *mongo.Collection, redisClient *redis.Client) {
	requestParams, err := req.URL.Query()["longUrl"]
	if !err || len(requestParams[0]) < 1 {
		io.WriteString(res, "URL parameter longUrl is missing")
	} else {
		longURL := requestParams[0]
		tinyURL, expireDate := GenerateHashAndInsert(longURL, 0, dbClient, redisClient)
		var data = types.Response{
			Url:      tinyURL,
			ExpireOn: expireDate,
		}
		res.Header().Set("Content-Type", "application/json")
		json.NewEncoder(res).Encode(data)
	}
}

func GenerateHashAndInsert(longURL string, startIndex int, dbClient *mongo.Collection, redisClient *redis.Client) (string, time.Time) {
	byteURLData := []byte(longURL)
	hashedURLData := fmt.Sprintf("%x", md5.Sum(byteURLData))
	tinyURLRegex, err := regexp.Compile("[/+]")
	if err != nil {
		return "Unable to generate tiny URL", time.Now()
	}
	tinyURLData := tinyURLRegex.ReplaceAllString(base64.URLEncoding.EncodeToString([]byte(hashedURLData)), "_")
	if len(tinyURLData) < (startIndex + 6) {
		return "Unable to generate tiny URL", time.Now()
	}
	tinyURL := tinyURLData[startIndex : startIndex+6]
	var dbURLData types.Urls
	_ = dbClient.FindOne(context.TODO(), bson.M{"tinyurl": tinyURL}).Decode(&dbURLData)

	if dbURLData.Tinyurl == "" {
		go StoreTinyURL(types.Urls{ID: primitive.NewObjectID(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(), Tinyurl: tinyURL, Longurl: longURL}, longURL, tinyURL, dbClient, redisClient)
		return tinyURL, time.Now().AddDate(0, 1, 0)
	} else {
		return GenerateHashAndInsert(longURL, startIndex+1, dbClient, redisClient)
	}
}

// GetLongHandler -> Fetches long URL and returns it
func GetLongHandler(res http.ResponseWriter, req *http.Request, dbClient *mongo.Collection, redisClient *redis.Client) {
	requestParams, err := req.URL.Query()["tinyUrl"]
	if !err || len(requestParams[0]) < 1 {
		io.WriteString(res, "URL parameter tinyUrl is missing")
	}
	tinyURL := requestParams[0]
	redisSearchResult := redisClient.HGet("urls", tinyURL)
	if redisSearchResult.Val() != "" {
		var data = types.ResponseLongURL{
			Url: redisSearchResult.Val(),
		}
		res.Header().Set("Content-Type", "application/json")
		json.NewEncoder(res).Encode(data)
	} else {
		var url types.Urls
		_ = dbClient.FindOneAndUpdate(context.TODO(), bson.M{"tinyurl": tinyURL}, bson.M{"updated_at": time.Now()}).Decode(&url)
		if url.Longurl != "" {
			redisClient.HSet("urls", tinyURL, url.Longurl)
			var data = types.ResponseLongURL{
				Url: url.Longurl,
			}
			res.Header().Set("Content-Type", "application/json")
			json.NewEncoder(res).Encode(data)
		} else {
			io.WriteString(res, "Unable to find long URL")
		}
	}
}

func main() {
	redisClient := Redis.RedisClient()

	pong, err := redisClient.Ping().Result()
	fmt.Println("Redis ping", pong, err)
	// MongoDB connection
	dbClient := db.ConnectDB()

	serverInstance := &http.Server{
		Addr: ":8080",
	}

	http.HandleFunc("/long/", func(w http.ResponseWriter, r *http.Request) {
		GetLongHandler(w, r, dbClient, redisClient)
	})

	http.HandleFunc("/tiny/", func(w http.ResponseWriter, r *http.Request) {
		GetTinyHandler(w, r, dbClient, Redis.RedisClient())
	})

	serverInstance.ListenAndServe()
}
