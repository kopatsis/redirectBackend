package routes

import (
	"fmt"
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/gin-gonic/gin"
)

func Redirect(client *datastore.Client) gin.HandlerFunc {
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

		c.Redirect(http.StatusSeeOther, entry.RealURL)
	}
}
