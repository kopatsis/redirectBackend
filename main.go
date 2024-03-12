package main

import (
	"c361main/routes"
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/datastore"
	"github.com/gin-gonic/gin"
)

const projectID = "cs361a"

// func indexHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.URL.Path != "/" {
// 		http.NotFound(w, r)
// 		return
// 	}
// 	fmt.Fprint(w, "Hello, World!")
// }

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func main() {

	rtr := gin.Default()

	rtr.Use(CORSMiddleware())

	ctx := context.Background()
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
		fmt.Println(err)
	}

	rtr.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"tik": "tok",
		})
	})

	rtr.POST("/user", routes.PostUser(client))
	rtr.POST("/login", routes.PostLoginUser(client))
	rtr.POST("/entry", routes.PostEntry(client))
	rtr.GET("/entry/:id", routes.GetEntry(client))
	rtr.DELETE("/entry/:id", routes.DeleteEntry(client))
	rtr.PATCH("/entry/:id", routes.PatchEntry(client))
	rtr.GET("/user/:id/entries", routes.GetEntries(client))
	rtr.GET("/r/:id", routes.Redirect(client))

	rtr.POST("/analyze", routes.PostClick(client))
	rtr.GET("/analyze/hourly/:id", routes.GetClicksHourly(client))
	rtr.GET("/analyze/daily/:id", routes.GetClicksDaily(client))
	rtr.GET("/analyze/weekly/:id", routes.GetClicksWeekly(client))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	rtr.Run(":" + port)

}
