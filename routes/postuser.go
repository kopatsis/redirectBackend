package routes

import (
	"time"

	"cloud.google.com/go/datastore"
	"github.com/gin-gonic/gin"
)

// Represents a local user
type User struct {
	Date time.Time
}

// Creates an instance of a local user, adds it to datastore, and returns the key to use in URL entry requests
func PostUser(client *datastore.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user User
		user.Date = time.Now()

		key := datastore.IncompleteKey("User", nil)
		newkey, err := client.Put(c, key, &user)
		if err != nil {
			errorPost(c, err, "Failed to post user")
			return
		}

		c.JSON(201, gin.H{
			"key": newkey.ID,
		})
	}
}
