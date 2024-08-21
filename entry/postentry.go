package entry

import (
	"c361main/convert"
	"c361main/datatypes"
	"context"
	"fmt"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/martian/v3/log"
	"gorm.io/gorm"
)

// Sets error message for post request using gin Context
func errorPost(c *gin.Context, err error, reason string) {
	fmt.Printf("Failed to post: %v", err)
	c.JSON(400, gin.H{
		"Error Type":  reason,
		"Exact Error": err.Error(),
	})
}

func PostEntryDB(db *gorm.DB, entry *datatypes.Entry) (int, error) {
	if err := db.Create(entry).Error; err != nil {
		return 0, err
	}
	return entry.ID, nil
}

func PostEntry(db *gorm.DB, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var entry datatypes.Entry

		if err := c.ShouldBindJSON(&entry); err != nil {
			errorPost(c, err, "Failed to bind entry body")
			return
		}

		entry.InitalizeFormat()

		if _, err := url.Parse(entry.RealURL); err != nil {
			errorPost(c, err, "Not a URL that can be parsed")
			return
		}

		id, err := PostEntryDB(db, &entry)
		if err != nil {
			errorPost(c, err, "Failed to post entry to database")
			return
		}

		sixFour, err := convert.ToSixFour(id)
		if err != nil {
			errorPost(c, err, "Serious issue: int does not work for six four conversion")
			return
		}

		if err := rdb.Set(context.Background(), sixFour, entry.RealURL, 0).Err(); err != nil {
			log.Errorf("Redis didn't post key: %s, val: %s, err: %s\n", sixFour, entry.RealURL, err.Error())
			return
		}

		c.JSON(201, gin.H{
			"parameter": sixFour,
		})
	}
}
