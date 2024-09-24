package main

import (
	"c361main/database"
	"c361main/entries"
	"c361main/platform"
	"c361main/specialty/firebaseAuth"
	"c361main/specialty/sendgridfn"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		if os.Getenv("APP_ENV") != "production" {
			log.Fatalf("Failed to load the env vars: %v", err)
		}
	}

	db, err := database.Connect()
	if err != nil {
		log.Fatal(err)
	}

	auth := firebaseAuth.InitFirebase()

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

	var httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}

	sendgridClient := sendgridfn.InitSendgrid()

	rtr := platform.New(db, auth, rdb, httpClient, sendgridClient)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := http.ListenAndServe(":"+port, rtr); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}
}
