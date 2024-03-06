package routes

import (
	"fmt"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/gin-gonic/gin"
)

type User struct {
	Date time.Time
}

func PostUser(client *datastore.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := User{
			Date: time.Now(),
		}

		key := datastore.IncompleteKey("User", nil)
		newkey, err := client.Put(c, key, &user)
		if err != nil {
			fmt.Printf("Failed to post: %v", err)
			c.JSON(400, gin.H{
				"Error Type":  "Failed to post user",
				"Exact Error": err.Error(),
			})
			return
		}

		c.JSON(201, gin.H{
			"key": newkey.ID,
		})
	}
}
