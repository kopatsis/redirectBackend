package routes

import (
	"fmt"
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

func GetEntries(client *datastore.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		idstr := c.Param("id")
		id10, err := strconv.ParseInt(idstr, 10, 64)
		if err != nil {
			fmt.Printf("Failed to get all: %v", err)
			c.JSON(400, gin.H{
				"Error Type":  "Failed to convert id param",
				"Exact Error": err.Error(),
			})
			return
		}

		var entries []EntryWithParam

		query := datastore.NewQuery("Entry").FilterField("User", "=", id10).FilterField("Archived", "=", false)
		keys, err := client.GetAll(c, query, &entries)
		if err != nil {
			fmt.Printf("Failed to get all: %v", err)
			c.JSON(400, gin.H{
				"Error Type":  "Failed to get entries for user",
				"Exact Error": err.Error(),
			})
			return
		}

		for i, key := range keys {
			entries[i].Param = toBase62(key.ID)
		}

		sort.Slice(entries, func(i, j int) bool {
			return entries[j].Date.Before(entries[i].Date)
		})

		ret := []gin.H{}

		for _, entry := range entries {
			current := gin.H{
				"param": entry.Param,
				"url":   entry.RealURL,
				"user":  entry.User,
				"date":  entry.Date,
			}
			ret = append(ret, current)
		}

		c.JSON(200, gin.H{
			"results": ret,
		})
	}
}
