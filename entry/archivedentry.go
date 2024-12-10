package entry

import (
	"c361main/datatypes"
	"c361main/user"
	"context"
	"errors"
	"fmt"
	"time"

	"firebase.google.com/go/v4/auth"
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

func ArchivedEntryDB(db *gorm.DB, id string) error {
	val := time.Now()
	return db.Model(&datatypes.Entry{}).Where("param = ?", id).Updates(datatypes.Entry{
		Archived:     true,
		ArchivedDate: &val,
	}).Error
}

func ArchivedEntry(db *gorm.DB, auth *auth.Client, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		userid, _, err := user.GetUserID(auth, c)
		if err != nil {
			errorDelete(c, err, "Failed to get user id", 400)
			return
		}

		id := c.Param("id")
		if id == "" {
			errorDelete(c, errors.New("no id param"), "Failed to convert id param", 400)
			return
		}

		var entry datatypes.Entry
		err = db.Where("param = ?", id).First(&entry).Error
		if err != nil {
			errorDelete(c, err, "Failed to retrieve entry", 400)
			return
		} else if entry.User != userid {
			errorDelete(c, errors.New("unauthorized for current entry"), "Wrong user for current entry", 400)
			return
		}

		if err := ArchivedEntryDB(db, id); err != nil {
			errorDelete(c, err, "Failed to archive on the database", 400)
			return
		}

		if err := rdb.Set(context.Background(), c.Param("id"), ":a:", 0).Err(); err != nil {
			errorDelete(c, err, "Failed to archive on redis", 400)
			return
		}

		c.Status(204)
	}
}
