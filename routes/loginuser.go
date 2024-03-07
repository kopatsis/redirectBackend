package routes

import (
	"fmt"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/gin-gonic/gin"
)

type LoginUser struct {
	AuthSub      string `json:"sub"`
	AuthName     string `json:"name"`
	AuthNickName string `json:"nickname"`
	AuthPicture  string `json:"picture"`
	Date         time.Time
}

func PostLoginUser(client *datastore.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		var user LoginUser

		if err := c.ShouldBindJSON(&user); err != nil {
			fmt.Printf("Failed to get login user: %v", err)
			c.JSON(400, gin.H{
				"Error Type":  "Failed to bind entry body",
				"Exact Error": err.Error(),
			})
			return
		}

		user.Date = time.Now()

		query := datastore.NewQuery("LoginUser").FilterField("AuthSub", "=", user.AuthSub).Limit(1)

		var results []User
		keys, err := client.GetAll(c, query, &results)
		if err != nil {
			c.JSON(400, gin.H{
				"Error Type":  "Failed to query existing login user",
				"Exact Error": err.Error(),
			})
			return
		}

		var key *datastore.Key
		if len(keys) > 0 {
			key = keys[0]
		} else {
			key = datastore.IncompleteKey("LoginUser", nil)
		}

		newkey, err := client.Put(c, key, &user)
		if err != nil {
			fmt.Printf("Failed to post: %v", err)
			c.JSON(400, gin.H{
				"Error Type":  "Failed to post login user",
				"Exact Error": err.Error(),
			})
			return
		}

		c.JSON(201, gin.H{
			"key":  newkey.ID,
			"user": user,
		})
	}
}
