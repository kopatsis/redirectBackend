package entry

import (
	"c361main/convert"
	"c361main/datatypes"
	"c361main/user"
	"context"
	"errors"
	"fmt"

	"firebase.google.com/go/v4/auth"
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

func UnarchivedEntryDB(db *gorm.DB, id int64) (string, error) {
	err := db.Model(&datatypes.Entry{}).
		Where("id = ?", id).
		Update("archived", false).
		Error
	if err != nil {
		return "", err
	}

	var entry datatypes.Entry
	err = db.First(&entry, id).Error
	if err != nil {
		return "", err
	}
	return entry.RealURL, nil
}

func UnarchivedEntry(db *gorm.DB, auth *auth.Client, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		userid, _, err := user.GetUserID(auth, c)
		if err != nil {
			errorPatch(c, err, "Failed to get user id", 400)
			return
		}

		id10, err := convert.FromSixFour(c.Param("id"))
		if err != nil {
			errorPatch(c, err, "Failed to convert id param", 400)
			return
		}

		var entry datatypes.Entry
		err = db.First(&entry, id10).Error
		if err != nil {
			errorDelete(c, err, "Failed to retrieve entry", 400)
			return
		} else if entry.User != userid {
			errorPatch(c, errors.New("unauthorized for current entry"), "Wrong user for current entry", 400)
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
