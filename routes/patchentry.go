package routes

import (
	"fmt"

	"cloud.google.com/go/datastore"
	"github.com/gin-gonic/gin"
)

// Sets error message for patch request using gin Context
func errorPatch(c *gin.Context, err error, reason string) {
	fmt.Printf("Failed to patch: %v", err)
	c.JSON(400, gin.H{
		"Error Type":  reason,
		"Exact Error": err.Error(),
	})
}

// Retrieves URL entry provided and patches status to unarchived, then sends confirmation
func PatchEntry(client *datastore.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		id10, err := fromBase62(c.Param("id"))
		if err != nil {
			errorPatch(c, err, "Failed to convert id param")
			return
		}

		var entry Entry

		getkey := datastore.IDKey("Entry", id10, nil)

		if err := client.Get(c, getkey, &entry); err != nil {
			errorPatch(c, err, "Failed to get entry")
			return
		}

		entry.Archived = false

		if _, err := client.Put(c, getkey, &entry); err != nil {
			errorPatch(c, err, "Failed to post entry")
			return
		}

		c.Status(204)
	}
}
