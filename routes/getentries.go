package routes

import (
	"sort"
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/gin-gonic/gin"
)

type EntryWithParam struct {
	User     int64  `json:"user"`
	RealURL  string `json:"url"`
	Date     time.Time
	Archived bool
	Param    string
}

// Formats a gin-friendly slice with the necessary attributes for an entry, including base 62 parameter
func formatReturnArray(entries []EntryWithParam) []gin.H {
	returnArray := []gin.H{}

	for _, entry := range entries {
		current := gin.H{
			"param": entry.Param,
			"url":   entry.RealURL,
			"user":  entry.User,
			"date":  entry.Date,
		}
		returnArray = append(returnArray, current)
	}

	return returnArray
}

// Retrieves a slice of all entries for a user and sends a slice with all necessary attributes
func GetEntries(client *datastore.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		id10, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			errorPost(c, err, "Failed to convert id param")
			return
		}

		var entries []EntryWithParam

		query := datastore.NewQuery("Entry").FilterField("User", "=", id10).FilterField("Archived", "=", false)
		keys, err := client.GetAll(c, query, &entries)
		if err != nil {
			errorPost(c, err, "Failed to get entries for user")
			return
		}

		for i, key := range keys {
			entries[i].Param = toBase62(key.ID)
		}

		sort.Slice(entries, func(i, j int) bool {
			return entries[j].Date.Before(entries[i].Date)
		})

		c.JSON(200, gin.H{
			"results": formatReturnArray(entries),
		})
	}
}
