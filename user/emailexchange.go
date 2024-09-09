package user

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

func AddExchange(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email" binding:"required"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		if c.GetHeader("X-Passcode-ID") != os.Getenv("CHECK_PASSCODE") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		id, err := AddEmail(req.Email, rdb)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store data"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"id": id})
	}
}

func GetExchange(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-Passcode-ID") != os.Getenv("CHECK_PASSCODE") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		id := c.Param("id")

		email, err := GetAndDeleteEmail(id, rdb)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Key not found or failed to retrieve"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"email": email})
	}
}

func AddEmail(email string, rdb *redis.Client) (string, error) {
	id := uuid.New().String()
	key := ":e:" + id
	err := rdb.Set(context.Background(), key, email, 0).Err()
	if err != nil {
		return "", err
	}
	return id, nil
}

func GetAndDeleteEmail(uuid string, rdb *redis.Client) (string, error) {
	key := ":e:" + uuid
	email, err := rdb.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return "", errors.New("dne before deletion")
	}
	if err != nil {
		return "", err
	}
	err = rdb.Del(context.Background(), key).Err()
	if err == redis.Nil {
		return email, errors.New("dne after deletion")
	}
	if err != nil {
		return email, err
	}
	return email, nil
}
