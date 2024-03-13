package routes

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"cloud.google.com/go/datastore"
	"github.com/gin-gonic/gin"
)

// Converts string representing base 62 number to base 10 int64
func fromBase62(str string) (int64, error) {
	if len(str) > 11 {
		return 0, errors.New("too large string to be actual int64")
	}
	var result int64
	base := len(baseChars)
	for i, char := range str {
		power := len(str) - i - 1
		index := int64(strings.IndexRune(baseChars, char))
		if index == -1 {
			return 0, errors.New("character used in id parameter not allowed")
		}
		result += index * int64(math.Pow(float64(base), float64(power)))
	}
	return result, nil
}

// Sets error message for get request using gin Context
func errorGet(c *gin.Context, err error, reason string, errorCode int) {
	fmt.Printf("Failed to get: %v", err)
	c.JSON(errorCode, gin.H{
		"Error Type":  reason,
		"Exact Error": err.Error(),
	})
}

// Retrieves the real URL for a URL if one exists in the system using the provided ID (or sends error)
func GetEntry(client *datastore.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		id10, err := fromBase62(c.Param("id"))
		if err != nil {
			errorGet(c, err, "Failed to convert id param", 400)
			return
		}

		getkey := datastore.IDKey("Entry", id10, nil)
		var entry Entry

		if err := client.Get(c, getkey, &entry); err != nil {
			errorGet(c, err, "Failed to get entry", 400)
			return
		}

		if entry.Archived {
			errorGet(c, errors.New("entry cannot be found"), "Link has been archived and deleted", 404)
			return
		}

		c.JSON(200, gin.H{
			"url": entry.RealURL,
		})
	}
}
