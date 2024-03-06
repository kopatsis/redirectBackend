package routes

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"cloud.google.com/go/datastore"
	"github.com/gin-gonic/gin"
)

func fromBase62(str string) (int64, error) {
	if len(str) > 11 {
		return 0, errors.New("Too large string to be actual int64")
	}
	var result int64
	base := len(baseChars)
	for i, char := range str {
		power := len(str) - i - 1
		index := int64(strings.IndexRune(baseChars, char))
		if index == -1 {
			return 0, errors.New("Character used in id parameter not allowed")
		}
		result += index * int64(math.Pow(float64(base), float64(power)))
	}
	return result, nil
}

func GetEntry(client *datastore.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		id62 := c.Param("id")
		id10, err := fromBase62(id62)
		if err != nil {
			fmt.Printf("Failed to get: %v", err)
			c.JSON(400, gin.H{
				"Error Type":  "Failed to convert id param",
				"Exact Error": err.Error(),
			})
			return
		}

		getkey := datastore.IDKey("Entry", id10, nil)
		var entry Entry

		if err := client.Get(c, getkey, &entry); err != nil {
			fmt.Printf("Failed to get: %v", err)
			c.JSON(400, gin.H{
				"Error Type":  "Failed to get entry",
				"Exact Error": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"url": entry.RealURL,
		})
	}
}
