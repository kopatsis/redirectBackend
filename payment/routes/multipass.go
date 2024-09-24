package routes

import (
	"c361main/payment/redisfn"
	"c361main/platform"
	"c361main/user"
	"net/http"
	"os"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/go-redis/redis/v8"

	"github.com/gin-gonic/gin"
)

func Multipass(rdb *redis.Client, auth *auth.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		originalURL := os.Getenv("OG_URL")
		if originalURL == "" {
			originalURL = "https://shortentrack.com"
		}

		uid, err := user.VerifyMultipassRetUID(c, auth)
		if err != nil {
			c.Redirect(http.StatusSeeOther, originalURL+"/login?circleRedir=t")
			return
		}

		ban, _, err := redisfn.CheckCookeLimit(rdb, uid, time.Now())
		if err != nil {
			c.Redirect(http.StatusSeeOther, "/loginerror")
			return
		}

		if ban {
			c.Redirect(http.StatusSeeOther, "/loginerror")
			return
		}

		platform.CreateCookie(c, uid)
		c.Redirect(http.StatusSeeOther, "/")
	}
}
