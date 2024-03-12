package routes

import (
	"fmt"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/gin-gonic/gin"
)

type Click struct {
	Param string    `json:"param" binding:"required"`
	Date  time.Time `json:"date"`
}

func PostClick(client *datastore.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var click Click

		if err := c.ShouldBindJSON(&click); err != nil {
			fmt.Printf("Failed to post: %v", err)
			c.JSON(400, gin.H{
				"Error Type":  "Failed to bind Click body",
				"Exact Error": err.Error(),
			})
			return
		}

		if click.Date.IsZero() {
			click.Date = time.Now()
		}

		key := datastore.IncompleteKey("Click", nil)
		newkey, err := client.Put(c, key, &click)
		if err != nil {
			fmt.Printf("Failed to post: %v", err)
			c.JSON(400, gin.H{
				"Error Type":  "Failed to post click",
				"Exact Error": err.Error(),
			})
			return
		}

		c.JSON(201, gin.H{
			"key": newkey.ID,
		})
	}
}
