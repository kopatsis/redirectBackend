package routes

import (
	"fmt"

	"cloud.google.com/go/datastore"
	"github.com/gin-gonic/gin"
)

// Sets error message for delete request using gin Context
func errorDelete(c *gin.Context, err error, reason string, errorCode int) {
	fmt.Printf("Failed to delete: %v", err)
	c.JSON(errorCode, gin.H{
		"Error Type":  reason,
		"Exact Error": err.Error(),
	})
}

// If it exists, sets the status of a URL entry to archived so it cannot be retrieved
func DeleteEntry(client *datastore.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		id10, err := fromBase62(c.Param("id"))
		if err != nil {
			errorDelete(c, err, "Failed to convert id param", 400)
			return
		}

		var entry Entry

		getkey := datastore.IDKey("Entry", id10, nil)

		if err := client.Get(c, getkey, &entry); err != nil {
			errorDelete(c, err, "Failed to get entry", 400)
			return
		}

		entry.Archived = true

		if _, err := client.Put(c, getkey, &entry); err != nil {
			errorDelete(c, err, "Failed to post entry", 400)
			return
		}

		c.Status(204)
	}
}
