package user

import (
	"context"
	"errors"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
)

func hasEmailPasswordAccount(client *auth.Client, uid string) (bool, error) {
	userRecord, err := client.GetUser(context.Background(), uid)
	if err != nil {
		return false, err
	}

	fmt.Println(userRecord.Email)

	for _, provider := range userRecord.ProviderUserInfo {
		fmt.Println(provider.DisplayName)
		fmt.Println(provider.ProviderID)
		if provider.ProviderID == "password" {
			return true, nil
		}
	}

	return false, nil
}

func HasPasswordHandler(firebase *firebase.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err, empty := GetSubFromJWT(firebase, c)
		if empty {
			c.JSON(401, gin.H{
				"Error Type":  "Incorrect auth",
				"Exact Error": errors.New("no token"),
			})
			return
		} else if err != nil {
			c.JSON(401, gin.H{
				"Error Type":  "Incorrect auth",
				"Exact Error": err.Error(),
			})
			return
		}

		authClient, err := firebase.Auth(context.Background())
		if err != nil {
			c.JSON(401, gin.H{
				"Error Type":  "Incorrect auth",
				"Exact Error": err.Error(),
			})
			return
		}

		has, err := hasEmailPasswordAccount(authClient, id)
		if err != nil {
			c.JSON(400, gin.H{
				"Error Type":  "Firebase error",
				"Exact Error": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"HasPassword": has,
		})
	}
}
