package middleware

import (
	"c361main/payment/redisfn"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func CreateCookie(c *gin.Context, userID string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "user_id",
		Value:    userID,
		Expires:  time.Now().Add(200 * 365 * 24 * time.Hour),
		Path:     "/",
		HttpOnly: true,
	})
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "date",
		Value:    time.Now().Format(time.RFC3339),
		Expires:  time.Now().Add(200 * 365 * 24 * time.Hour),
		Path:     "/",
		HttpOnly: true,
	})
}

func RemoveCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:    "user_id",
		Value:   "",
		Expires: time.Unix(0, 0),
		Path:    "/",
	})
	http.SetCookie(c.Writer, &http.Cookie{
		Name:    "date",
		Value:   "",
		Expires: time.Unix(0, 0),
		Path:    "/",
	})
}

func GetCookieUserID(c *gin.Context) (string, error) {
	cookie, err := c.Cookie("user_id")
	if err != nil {
		return "", err
	} else if cookie == "" {
		return "", errors.New("user_id cookie not found")
	}
	return cookie, nil
}

func GetCookie(c *gin.Context) (string, time.Time, error) {
	userID, err := c.Cookie("user_id")
	if err != nil {
		return "", time.Time{}, err
	} else if userID == "" {
		return "", time.Time{}, errors.New("user_id cookie not found")
	}

	dateStr, err := c.Cookie("date")
	if err != nil {
		return "", time.Time{}, err
	} else if dateStr == "" {
		return "", time.Time{}, errors.New("date cookie not found")
	}

	date, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return "", time.Time{}, errors.New("failed to parse date cookie")
	}

	return userID, date, nil
}

func errorSplit(c *gin.Context, err error, banned bool) {
	method := c.Request.Method
	if method == "GET" {
		if !banned {
			loginURL := os.Getenv("OG_URL")
			if loginURL == "" {
				loginURL = "https://shortentrack.com"
			}
			c.Redirect(http.StatusSeeOther, loginURL+"/login?circleRedir=t")
			return
		} else {
			c.Redirect(http.StatusSeeOther, "/loginerror")
			return
		}
	}
	c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
}

func CookieMiddleware(rdb *redis.Client) gin.HandlerFunc {
	paths := []string{"/user", "/merge", "/entry", "/search", "/entriescsv", "/clicks", "/clickcsv", "/haspassword", "/emailexchange", "/customcheck", "/multipass", "/webhook", "/administrative", "/check", "/verifyturn", "/temporarytest", "/emailsubbed"}

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		for _, p := range paths {
			if strings.HasPrefix(path, p) {
				c.Next()
				return
			}
		}

		userID, date, err := GetCookie(c)
		if err != nil {
			errorSplit(c, errors.New("unauthorized: missing cookie data"), false)
			c.Abort()
			return
		}

		ban, reset, err := redisfn.CheckCookeLimit(rdb, userID, date)
		if err != nil {
			errorSplit(c, errors.New("unauthorized: failure with redis"), false)
			c.Abort()
			return
		}

		if reset {
			errorSplit(c, errors.New("unauthorized: not logged in"), false)
			c.Abort()
			return
		}

		if ban {
			RemoveCookie(c)
			errorSplit(c, errors.New("unauthorized: user does not exist"), true)
			c.Abort()
			return
		}

		c.Next()
	}
}
