package routes

import (
	"fmt"
	"time"

	// "strconv"
	"strings"

	"cloud.google.com/go/datastore"
	"github.com/gin-gonic/gin"
)

const baseChars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type Entry struct {
	User     int64  `json:"user"`
	RealURL  string `json:"url"`
	Archived bool
	Date     time.Time
}

func ensureHttpPrefix(url string) string {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "http://" + url
	}
	return url
}

func toBase62(num int64) string {
	if num == 0 {
		return string(baseChars[0])
	}
	var result strings.Builder
	base := int64(len(baseChars))
	for num > 0 {
		result.WriteString(string(baseChars[num%base]))
		num /= base
	}
	return reverse(result.String())
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func PostEntry(client *datastore.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var entry Entry

		if err := c.ShouldBindJSON(&entry); err != nil {
			fmt.Printf("Failed to post: %v", err)
			c.JSON(400, gin.H{
				"Error Type":  "Failed to bind entry body",
				"Exact Error": err.Error(),
			})
			return
		}

		entry.Date = time.Now()
		entry.Archived = false
		entry.RealURL = ensureHttpPrefix(entry.RealURL)

		key := datastore.IncompleteKey("Entry", nil)
		newkey, err := client.Put(c, key, &entry)
		if err != nil {
			fmt.Printf("Failed to post: %v", err)
			c.JSON(400, gin.H{
				"Error Type":  "Failed to post entry",
				"Exact Error": err.Error(),
			})
			return
		}

		// param := strconv.FormatInt(newkey.ID, 36)
		param := toBase62(newkey.ID)

		c.JSON(201, gin.H{
			"parameter": param,
		})
	}
}
