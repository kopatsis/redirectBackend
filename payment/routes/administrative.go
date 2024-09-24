package routes

import (
	"c361main/payment/redisfn"
	"c361main/platform"
	"c361main/specialty/sendgridfn"
	"c361main/user"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/sendgrid/sendgrid-go"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
)

func HandleLogAllOut(rdb *redis.Client, auth *auth.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, err, empty := user.GetSubFromJWT(auth, c)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		} else if empty {
			c.JSON(400, gin.H{"error": "no token"})
			return
		}

		if err := redisfn.AddResetDate(rdb, uid); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		response := map[string]any{"success": true}
		c.JSON(200, response)
	}
}

func HandleDeleteAccount(rdb *redis.Client, auth *auth.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, err, empty := user.GetSubFromJWT(auth, c)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		} else if empty {
			c.JSON(400, gin.H{"error": "no token"})
			return
		}

		if err := redisfn.AddBanned(rdb, uid); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		response := map[string]any{"success": true}
		c.JSON(200, response)
	}
}

func HandleUserLogout() gin.HandlerFunc {
	return func(c *gin.Context) {
		platform.RemoveCookie(c)

		response := map[string]any{"loggedout": true}
		c.JSON(200, response)
	}
}

func HandleInternalAlertEmail(sendgridClient *sendgrid.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		shouldBe := os.Getenv("CHECK_PASSCODE")
		if shouldBe == "" {
			c.JSON(500, gin.H{"error": "no passcode exists on backend"})
			return
		}

		passcode := c.Request.Header.Get("X-Passcode-ID")
		if passcode == "" {
			c.JSON(400, gin.H{"error": "no passcode provided"})
			return
		} else if passcode != shouldBe {
			c.JSON(400, gin.H{"error": "incorrect passcode provided"})
			return
		}

		var payload struct {
			Subject string `json:"subject"`
			Body    string `json:"body"`
		}

		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		if err := sendgridfn.SendSeriousErrorAlert(sendgridClient, payload.Subject, payload.Body); err != nil {
			sendgridfn.SendSeriousErrorAlert(sendgridClient, "Sending the Actual Issue Email", "This error: "+err.Error())
		}

		response := map[string]any{"success": true}
		c.JSON(200, response)
	}
}
