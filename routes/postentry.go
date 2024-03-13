package routes

import (
	"fmt"
	"time"

	"strings"

	"cloud.google.com/go/datastore"
	"github.com/gin-gonic/gin"
)

const baseChars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Represents an entry for a shortened URL
type Entry struct {
	User     int64  `json:"user"`
	RealURL  string `json:"url"`
	Archived bool
	Date     time.Time
}

// Initializes attributes for a user for those not provided
func (entry *Entry) InitalizeFormat() {
	entry.Date = time.Now()
	entry.Archived = false
	entry.RealURL = ensureHttpPrefix(entry.RealURL)
}

// If URL doesn't have a protocol prefix, one is added
func ensureHttpPrefix(url string) string {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "http://" + url
	}
	return url
}

// Converts int64 number to base 62 number with string representation
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

// Reverses a string by by rune rather than by character
func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Sets error message for post request using gin Context
func errorPost(c *gin.Context, err error, reason string) {
	fmt.Printf("Failed to post: %v", err)
	c.JSON(400, gin.H{
		"Error Type":  reason,
		"Exact Error": err.Error(),
	})
}

// Creates a URL entry from provided data and pushes it to datastore, then sending confirmation of post
func PostEntry(client *datastore.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var entry Entry

		if err := c.ShouldBindJSON(&entry); err != nil {
			errorPost(c, err, "Failed to bind entry body")
			return
		}

		entry.InitalizeFormat()

		key := datastore.IncompleteKey("Entry", nil)
		newkey, err := client.Put(c, key, &entry)
		if err != nil {
			errorPost(c, err, "Failed to post entry")
			return
		}

		c.JSON(201, gin.H{
			"parameter": toBase62(newkey.ID),
		})
	}
}
