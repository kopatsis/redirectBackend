package routes

import (
	"fmt"

	"cloud.google.com/go/datastore"
	"github.com/gin-gonic/gin"
)

func PatchEntry(client *datastore.Client) gin.HandlerFunc {
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

		var entry Entry

		getkey := datastore.IDKey("Entry", id10, nil)

		if err := client.Get(c, getkey, &entry); err != nil {
			fmt.Printf("Failed to get: %v", err)
			c.JSON(400, gin.H{
				"Error Type":  "Failed to get entry",
				"Exact Error": err.Error(),
			})
			return
		}

		entry.Archived = false

		if _, err := client.Put(c, getkey, &entry); err != nil {
			fmt.Printf("Failed to post: %v", err)
			c.JSON(400, gin.H{
				"Error Type":  "Failed to post entry",
				"Exact Error": err.Error(),
			})
			return
		}

		c.Status(204)
	}
}
