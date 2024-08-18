package main

import (
	"c361main/database"
	"c361main/platform"
	"c361main/routes"
	"context"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/datastore"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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

	rtr := platform.New(db, client)

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
