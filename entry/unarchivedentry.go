package entry

import (
	"c361main/convert"
	"c361main/datatypes"
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/martian/v3/log"
	"gorm.io/gorm"
)

func errorPatch(c *gin.Context, err error, reason string, errorCode int) {
	fmt.Printf("Failed to patch: %v", err)
	c.JSON(errorCode, gin.H{
		"Error Type":  reason,
		"Exact Error": err.Error(),
	})
}

func UnarchivedEntryDB(db *gorm.DB, id int) (string, error) {
	var entry datatypes.Entry
	err := db.Model(&datatypes.Entry{}).
		Where("id = ?", id).
		Updates(datatypes.Entry{Archived: false}).
		First(&entry).Error
	if err != nil {
		return "", err
	}
	return entry.RealURL, nil
}

func UnarchivedEntry(db *gorm.DB, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		id10, err := convert.FromSixFour(c.Param("id"))
		if err != nil {
			errorPatch(c, err, "Failed to convert id param", 400)
			return
		}

		url, err := UnarchivedEntryDB(db, id10)
		if err != nil {
			errorPatch(c, err, "Failed to archive on the database", 400)
			return
		}

		if err := rdb.Set(context.Background(), c.Param("id"), url, 0).Err(); err != nil {
			log.Errorf("Redis didn't post key: %s, val: %s, err: %s\n", c.Param("id"), url, err.Error())
			return
		}

		c.Status(204)
	}
}
