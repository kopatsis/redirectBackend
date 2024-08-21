package main

import (
	"c361main/database"
	"c361main/entries"
	"c361main/platform"
	"c361main/routes"
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/datastore"
	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

const projectID = "cs361a"

func main() {

	if err := godotenv.Load(); err != nil {
		if os.Getenv("APP_ENV") != "production" {
			log.Fatalf("Failed to load the env vars: %v", err)
		}
	}

	ctx := context.Background()
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	db, err := database.Connect()
	if err != nil {
		log.Fatal(err)
	}

	firebaseConfigBase64 := os.Getenv("FIREBASE_CONFIG_BASE64")
	if firebaseConfigBase64 == "" {
		log.Fatal("FIREBASE_CONFIG_BASE64 environment variable is not set.")
	}

	configJSON, err := base64.StdEncoding.DecodeString(firebaseConfigBase64)
	if err != nil {
		log.Fatalf("Error decoding FIREBASE_CONFIG_BASE64: %v", err)
	}

	sa := option.WithCredentialsJSON(configJSON)
	firebase, err := firebase.NewApp(context.Background(), nil, sa)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	scheduler := gocron.NewScheduler(time.UTC)

	_, err = scheduler.Every(8).Hours().Do(entries.DeleteArchivedEntries, db)
	if err != nil {
		log.Fatalf("Error scheduling cleanup job: %v", err)
	}

	scheduler.StartAsync()

	redisAddr, redisPass := os.Getenv("REDIS_ADDR"), os.Getenv("REDIS_PASSWORD")
	if redisAddr == "" || redisPass == "" {
		log.Fatalf("cannot connect to redis as no addr and/or pass present in env")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Username: "default",
		Password: redisPass,
		DB:       0,
	})

	rtr := platform.New(db, client, firebase, rdb)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := http.ListenAndServe(":"+port, rtr); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}
}

func DefineRoutesReal(rtr *gin.Engine, client *datastore.Client) {
	rtr.GET("/", func(c *gin.Context) { //
		c.JSON(200, gin.H{ //
			"tik": "tok", //
		}) //
	}) //

	rtr.POST("/user", routes.PostUser(client))
	rtr.POST("/login", routes.PostLoginUser(client))
	rtr.POST("/entry", routes.PostEntry(client))
	rtr.GET("/entry/:id", routes.GetEntry(client))
	rtr.DELETE("/entry/:id", routes.DeleteEntry(client))
	rtr.PATCH("/entry/:id", routes.PatchEntry(client))
	rtr.GET("/user/:id/entries", routes.GetEntries(client))
	rtr.GET("/r/:id", routes.Redirect(client)) //

	rtr.POST("/analyze", routes.PostClick(client))                 //
	rtr.GET("/analyze/hourly/:id", routes.GetClicksHourly(client)) //
	rtr.GET("/analyze/daily/:id", routes.GetClicksDaily(client))   //
	rtr.GET("/analyze/weekly/:id", routes.GetClicksWeekly(client)) //
}
