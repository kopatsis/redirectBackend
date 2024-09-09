package entry

import (
	"c361main/convert"
	"c361main/datatypes"
	"c361main/user"
	"context"
	"errors"
	"net/http"
	"os"
	"regexp"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func CheckCustomHandleExists(db *gorm.DB, handle string) (bool, error) {
	var count int64
	err := db.Model(&datatypes.Entry{}).Where("custom_handle = ?", handle).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func UpdateCustomHandle(db *gorm.DB, handle string, id int) error {
	var entry datatypes.Entry
	if err := db.First(&entry, id).Error; err != nil {
		return err
	}
	entry.CustomHandle = handle
	return db.Save(&entry).Error
}

func GetEntryByID(db *gorm.DB, id int) (*datatypes.Entry, error) {
	var entry datatypes.Entry
	if err := db.First(&entry, id).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

func CheckCustomHandle(db *gorm.DB, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		handle := c.Param("id")

		if len(handle) < 7 || len(handle) > 256 {
			errorPatch(c, errors.New("wrong size for new handle"), "Invalid or missing JSON input", 400)
			return
		}

		re := regexp.MustCompile(`^[a-zA-Z0-9_-]*$`)
		if !re.MatchString(handle) {
			errorPatch(c, errors.New("wrong characters for new handle"), "Invalid or missing JSON input", 400)
			return
		}

		if c.GetHeader("X-Passcode-ID") != os.Getenv("CHECK_PASSCODE") {
			errorPatch(c, errors.New("no or wrong passcode id attacked"), "Invalid or missing passcode", 400)
			return
		}

		exists, err := CheckCustomHandleExists(db, handle)
		if err != nil {
			errorPatch(c, err, "failed to check from db if handle exists", 400)
			return
		}

		count, err := rdb.Exists(context.Background(), handle).Result()
		if err != nil {
			errorPatch(c, err, "failed to check from redis if handle exists", 400)
			return
		}

		c.JSON(200, gin.H{
			"available": !exists && count == 0,
		})
	}
}

func PatchCustomHandle(db *gorm.DB, app *firebase.App, rdb *redis.Client, httpClient *http.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		id10, err := convert.FromSixFour(c.Param("id"))
		if err != nil {
			errorPatch(c, err, "Failed to convert id param", 400)
			return
		}

		entry, err := GetEntryByID(db, id10)
		if err != nil {
			errorPatch(c, err, "Could not get actual entry", 400)
			return
		}

		var json struct {
			Handle string `json:"handle" binding:"required"`
		}

		if err := c.ShouldBindJSON(&json); err != nil {
			errorPatch(c, err, "Invalid or missing JSON input", 400)
			return
		}

		if len(json.Handle) < 7 || len(json.Handle) > 256 {
			errorPatch(c, errors.New("wrong size for new handle"), "Invalid or missing JSON input", 400)
			return
		}

		re := regexp.MustCompile(`^[a-zA-Z0-9_-]*$`)
		if !re.MatchString(json.Handle) {
			errorPatch(c, errors.New("wrong characters for new handle"), "Invalid or missing JSON input", 400)
			return
		}

		userid, inFirebase, err := user.GetUserID(app, c)
		if err != nil {
			errorPatch(c, err, "failed to get jwt or header user id", 400)
			return
		}

		if userid != entry.User {
			errorPatch(c, errors.New("unauthorized"), "Doesn't belong to that user for entry", 400)
			return
		}

		paying := false
		if inFirebase {
			paying, err = user.CheckPaymentStatus(userid, httpClient)
			if err != nil {
				errorPatch(c, err, "failed to correctly check status of user payment", 400)
				return
			}
		}

		if !paying {
			errorPatch(c, errors.New("must be a paying membership to modify custom handle"), "not a valid paying member", 400)
			return
		}

		count, err := rdb.Exists(context.Background(), json.Handle).Result()
		if err != nil {
			errorPatch(c, err, "failed to check from redis if handle exists", 400)
			return
		} else if count > 0 {
			errorPatch(c, err, "exists in redis already", 400)
			return
		}

		if err := UpdateCustomHandle(db, json.Handle, id10); err != nil {
			errorPatch(c, err, "could not save to db", 400)
			return
		}

		if err := rdb.Set(context.Background(), json.Handle, entry.RealURL, 0).Err(); err != nil {
			errorPatch(c, err, "could not save to redis", 400)
			return
		}

	}
}

func DeleteCustomHandle(db *gorm.DB, app *firebase.App, rdb *redis.Client, httpClient *http.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		id10, err := convert.FromSixFour(c.Param("id"))
		if err != nil {
			errorPatch(c, err, "Failed to convert id param", 400)
			return
		}

		entry, err := GetEntryByID(db, id10)
		if err != nil {
			errorPatch(c, err, "Could not get actual entry", 400)
			return
		}

		userid, inFirebase, err := user.GetUserID(app, c)
		if err != nil {
			errorPatch(c, err, "failed to get jwt or header user id", 400)
			return
		}

		if userid != entry.User {
			errorPatch(c, errors.New("unauthorized"), "Doesn't belong to that user for entry", 400)
			return
		}

		paying := false
		if inFirebase {
			paying, err = user.CheckPaymentStatus(userid, httpClient)
			if err != nil {
				errorPatch(c, err, "failed to correctly check status of user payment", 400)
				return
			}
		}

		if !paying {
			errorPatch(c, errors.New("must be a paying membership to modify custom handle"), "not a valid paying member", 400)
			return
		}

		if err := UpdateCustomHandle(db, "", id10); err != nil {
			errorPatch(c, err, "could not save to db", 400)
			return
		}

		if err := rdb.Del(context.Background(), entry.CustomHandle).Err(); err != nil {
			errorPatch(c, err, "could not save to redis", 400)
			return
		}
	}
}
