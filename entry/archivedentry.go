package entry

import (
	"c361main/convert"
	"c361main/datatypes"
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func errorDelete(c *gin.Context, err error, reason string, errorCode int) {
	fmt.Printf("Failed to delete: %v", err)
	c.JSON(errorCode, gin.H{
		"Error Type":  reason,
		"Exact Error": err.Error(),
	})
}

func ArchivedEntryDB(db *gorm.DB, id int) error {
	return db.Model(&datatypes.Entry{}).Where("id = ?", id).Updates(datatypes.Entry{
		Archived:     true,
		ArchivedDate: time.Now(),
	}).Error
}

func ArchivedEntry(db *gorm.DB, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		id10, err := convert.FromSixFour(c.Param("id"))
		if err != nil {
			errorDelete(c, err, "Failed to convert id param", 400)
			return
		}

		if err := ArchivedEntryDB(db, id10); err != nil {
			errorDelete(c, err, "Failed to archive on the database", 400)
			return
		}

		if err := rdb.Del(context.Background(), c.Param("id")).Err(); err != nil {
			errorDelete(c, err, "Failed to archive on actual redis", 400)
			return
		}

		c.Status(204)
	}
}
