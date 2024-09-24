package routes

import (
	"c361main/payment/redisfn"
	"c361main/specialty/cloudflare"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

type TurnstilePost struct {
	Email   string `json:"email"`
	Captcha string `json:"cf-turnstile-response"`
}

func ExternalGetHandler(rdb *redis.Client) gin.HandlerFunc {
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

		id := c.Param("id")
		if id == "" {
			c.JSON(400, gin.H{"error": "no param provided"})
			return
		}

		paying, err := redisfn.CheckUserPaying(rdb, id)
		if err != nil && !paying {
			c.JSON(400, gin.H{"error": "unable to query payment status"})
			return
		}

		c.JSON(200, gin.H{"paying": paying})
	}
}

func VerifyTurnstileHandler(client *http.Client) gin.HandlerFunc {
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

		var form TurnstilePost
		if err := c.ShouldBindJSON(&form); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		} else if form.Email == "" {
			c.JSON(400, gin.H{"message": "Please supply an email along with the turnstile verification."})
			return
		}

		success, err := cloudflare.VerifyTurnstile(client, form.Captcha)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		} else if !success {
			c.JSON(401, gin.H{"message": "Unfortunately, your submission did not pass the Cloudflare verification. Close this window and try again."})
			return
		}

		c.Status(204)
	}
}
