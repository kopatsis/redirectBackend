package entries

import (
	"c361main/convert"
	"c361main/datatypes"
	"c361main/user"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func errorGet(c *gin.Context, err error, reason string) {
	fmt.Printf("Failed to get: %v", err)
	c.JSON(400, gin.H{
		"Error Type":  reason,
		"Exact Error": err.Error(),
	})
}

func GetEntriesDB(db *gorm.DB, user string) ([]datatypes.ShortenedEntry, error) {
	var entries []datatypes.Entry
	var shortenedEntries []datatypes.ShortenedEntry

	if err := db.Where("user = ? AND archived = ?", user, false).Find(&entries).Error; err != nil {
		return nil, err
	}

	for _, entry := range entries {

		param, err := convert.ToSixFour(entry.ID)
		if err != nil {
			return nil, err
		}

		shortenedEntry := datatypes.ShortenedEntry{
			Param:   param,
			User:    entry.User,
			RealURL: entry.RealURL,
			Date:    entry.Date,
		}
		shortenedEntries = append(shortenedEntries, shortenedEntry)
	}

	return shortenedEntries, nil
}

func GetEntries(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userid, _, err := user.GetUserID(c)
		if err != nil {
			errorGet(c, err, "failed to get jwt or header user id")
			return
		}

		entries, err := GetEntriesDB(db, userid)
		if err != nil {
			errorGet(c, err, "failed to query actual entries")
			return
		}

		c.JSON(200, gin.H{
			"entries": &entries,
		})
	}
}
