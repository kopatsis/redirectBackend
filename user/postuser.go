package user

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func errorPost(c *gin.Context, err error, reason string) {
	fmt.Printf("Failed to post: %v", err)
	c.JSON(400, gin.H{
		"Error Type":  reason,
		"Exact Error": err.Error(),
	})
}

func PostUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie := GetUserCookie(c)

		c.JSON(201, gin.H{
			"key": cookie,
		})
	}
}

func UserCookie(c *gin.Context) string {
	newUUID := uuid.New().String()

	c.SetCookie("useruuid", newUUID, 2147483647, "/", ".shortentrack.com", false, false)

	return newUUID
}

func GetUserCookie(c *gin.Context) string {
	userUUID, err := c.Cookie("useruuid")
	if err != nil {
		return UserCookie(c)
	}
	return userUUID
}
