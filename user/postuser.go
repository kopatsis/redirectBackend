package user

import (
	"c361main/datatypes"
	"fmt"
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/gin-gonic/gin"
)

func errorPost(c *gin.Context, err error, reason string) {
	fmt.Printf("Failed to post: %v", err)
	c.JSON(400, gin.H{
		"Error Type":  reason,
		"Exact Error": err.Error(),
	})
}

func PostUser(client *datastore.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user datatypes.User
		user.Date = time.Now()

		key := datastore.IncompleteKey("User", nil)
		newkey, err := client.Put(c, key, &user)
		if err != nil {
			errorPost(c, err, "Failed to post user")
			return
		}

		c.JSON(201, gin.H{
			"key": "DS-" + strconv.FormatInt(newkey.ID, 10),
		})
	}
}
