package routes

import (
	"errors"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/gin-gonic/gin"
)

// Represents a login using an email or SSO provider such as Google or Facebook
type LoginUser struct {
	AuthSub      string `json:"sub"`
	AuthName     string `json:"name"`
	AuthNickName string `json:"nickname"`
	AuthPicture  string `json:"picture"`
	Date         time.Time
}

// If a user with the Auth0 sub id exists, provides the key for this user, else provides a blank key to create a new user
func createKey(client *datastore.Client, user LoginUser, c *gin.Context) *datastore.Key {

	query := datastore.NewQuery("LoginUser").FilterField("AuthSub", "=", user.AuthSub).Limit(1)

	var results []LoginUser
	keys, err := client.GetAll(c, query, &results)
	if err != nil {
		errorPost(c, err, "Failed to query existing login user")
		return nil
	}

	var key *datastore.Key
	if len(keys) > 0 {
		key = keys[0]
	} else {
		key = datastore.IncompleteKey("LoginUser", nil)
	}
	return key
}

// Posts a login instance for a user, creating a new one if one doesn't already exist, and sends confirmation
func PostLoginUser(client *datastore.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		var user LoginUser

		if err := c.ShouldBindJSON(&user); err != nil {
			errorPost(c, err, "Failed to bind entry body")
			return
		}

		originalKey := createKey(client, user, c)
		if originalKey == nil {
			errorPost(c, errors.New("datastore error querying user"), "Failed to query existing login user")
			return
		}

		user.Date = time.Now()
		newkey, err := client.Put(c, originalKey, &user)
		if err != nil {
			errorPost(c, err, "Failed to post login user")
			return
		}

		c.JSON(201, gin.H{
			"key":  newkey.ID,
			"user": user,
		})
	}
}
