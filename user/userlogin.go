package user

import (
	"bytes"
	"context"
	"errors"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func OLDhasEmailPasswordAccount(client *auth.Client, uid string) (bool, error) {
	userRecord, err := client.GetUser(context.Background(), uid)
	if err != nil {
		return false, err
	}

	for _, provider := range userRecord.ProviderUserInfo {
		if provider.ProviderID == "password" {
			return true, nil
		}
	}

	return false, nil
}

func hasPasswordAccount(uid string, rdb *redis.Client) (bool, error) {
	val, err := rdb.Get(context.Background(), ":HPA:"+uid).Bytes()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}

	return bytes.Equal(val, []byte{1}), nil
}

func addHasPasswordAccount(uid string, rdb *redis.Client) error {
	return rdb.Set(context.Background(), ":HPA:"+uid, []byte{1}, 0).Err()
}

func HasPasswordPost(firebase *firebase.App, rdb *redis.Client) gin.HandlerFunc {
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

		if err := addHasPasswordAccount(id, rdb); err != nil {
			c.JSON(400, gin.H{
				"Error Type":  "Unable to create entry for has password",
				"Exact Error": err.Error(),
			})
			return
		}

		c.JSON(201, gin.H{"uid": id})
	}
}

func HasPasswordHandler(firebase *firebase.App, rdb *redis.Client) gin.HandlerFunc {
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

		has, err := hasPasswordAccount(id, rdb)
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
