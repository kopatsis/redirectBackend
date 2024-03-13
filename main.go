package main

import (
	"c361main/routes"
	"context"
	"log"
	"os"

	"cloud.google.com/go/datastore"
	"github.com/gin-gonic/gin"
)

const projectID = "cs361a"

// Defines attributes needed for outside domains to access this service while hosted
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		headers := "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With"

		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", headers)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// Specifies functions for each route used in this service
func defineRoutes(rtr *gin.Engine, client *datastore.Client) {
	rtr.POST("/user", routes.PostUser(client))
	rtr.POST("/login", routes.PostLoginUser(client))
	rtr.POST("/entry", routes.PostEntry(client))

	rtr.GET("/entry/:id", routes.GetEntry(client))
	rtr.DELETE("/entry/:id", routes.DeleteEntry(client))
	rtr.PATCH("/entry/:id", routes.PatchEntry(client))
	rtr.GET("/user/:id/entries", routes.GetEntries(client))
}

// Defines the port specified, or 8080 if none is supplied
func createPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}
	return port
}

// Function that is called first when this is ran, all setup and teardown for this gin server
func main() {
	rtr := gin.Default()

	rtr.Use(CORSMiddleware())

	ctx := context.Background()
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	defineRoutes(rtr, client)

	port := createPort()

	log.Printf("Listening on port %s", port)
	rtr.Run(":" + port)
}

func defineRoutesReal(rtr *gin.Engine, client *datastore.Client) {
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
