package routes

import (
	"fmt"

	"cloud.google.com/go/datastore"
	"github.com/gin-gonic/gin"
)

func DeleteEntry(client *datastore.Client) gin.HandlerFunc {
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

		if err := client.Delete(c, getkey); err != nil {
			fmt.Printf("Failed to delete: %v", err)
			c.JSON(400, gin.H{
				"Error Type":  "Failed to delete entry",
				"Exact Error": err.Error(),
			})
			return
		}

		c.Status(204)
	}
}
