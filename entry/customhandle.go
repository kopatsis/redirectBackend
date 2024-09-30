package entry

import (
	"c361main/convert"
	"c361main/datatypes"
	"c361main/payment/redisfn"
	"c361main/user"
	"context"
	"encoding/json"
	"errors"
	"os"
	"regexp"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func CreateCustomHandleStruct(url, userid string, param int64) string {
	custom := datatypes.Custom{
		URL:    url,
		UserID: userid,
		Param:  param,
	}

	str, _ := json.Marshal(custom)
	return string(str)
}

func CheckCustomHandleExists(db *gorm.DB, handle string) (bool, error) {
	var count int64
	err := db.Model(&datatypes.Entry{}).Where("custom_handle = ?", handle).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func UpdateCustomHandle(db *gorm.DB, handle string, id int64) error {
	var entry datatypes.Entry
	if err := db.First(&entry, id).Error; err != nil {
		return err
	}
	entry.CustomHandle = handle
	return db.Save(&entry).Error
}

func GetEntryByID(db *gorm.DB, id int64) (*datatypes.Entry, error) {
	var entry datatypes.Entry
	if err := db.First(&entry, id).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

func CheckCustomHandle(db *gorm.DB, auth *auth.Client, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		handle := c.Param("id")

		if len(handle) < 7 || len(handle) > 128 {
			errorPatch(c, errors.New("wrong size for new handle"), "Invalid or missing JSON input", 400)
			return
		}

		re := regexp.MustCompile(`^[a-zA-Z0-9_-]*$`)
		if !re.MatchString(handle) {
			errorPatch(c, errors.New("wrong characters for new handle"), "Invalid or missing JSON input", 400)
			return
		}

		userid, _, err := user.GetUserID(auth, c)
		if err != nil {
			errorPatch(c, err, "failed to get jwt or header user id", 400)
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

		existing, err := rdb.Get(context.Background(), handle).Result()
		if err != nil && err != redis.Nil {
			errorPatch(c, err, "failed to check from redis if handle exists", 400)
			return
		}

		c.JSON(200, gin.H{
			"available": !exists && (existing == "" || existing == ":e:"+userid),
		})
	}
}

func PatchCustomHandle(db *gorm.DB, auth *auth.Client, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userid, inFirebase, err := user.GetUserID(auth, c)
		if err != nil {
			errorPatch(c, err, "failed to get jwt or header user id", 400)
			return
		}

		paying := false
		if inFirebase {
			paying, err = redisfn.CheckUserPaying(rdb, userid)
			if err != nil {
				errorPatch(c, err, "failed to correctly check status of user payment", 400)
				return
			}
		}

		if !paying {
			errorPatch(c, errors.New("must be a paying membership to modify custom handle"), "not a valid paying member", 400)
			return
		}

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

		if userid != entry.User {
			errorPatch(c, errors.New("unauthorized"), "Doesn't belong to that user for entry", 400)
			return
		}

		existing, err := rdb.Get(context.Background(), json.Handle).Result()
		if err != nil && err != redis.Nil {
			errorPatch(c, err, "failed to check from redis if handle exists", 400)
			return
		} else if existing != "" && existing != ":e:"+userid {
			errorPatch(c, errors.New("not allowed to use hanlde"), "handle already exists and is not :e: + userid", 400)
			return
		}

		if err := UpdateCustomHandle(db, json.Handle, id10); err != nil {
			errorPatch(c, err, "could not save to db", 400)
			return
		}

		if err := rdb.Set(context.Background(), json.Handle, CreateCustomHandleStruct(entry.RealURL, userid, entry.ID), 0).Err(); err != nil {
			errorPatch(c, err, "could not save to redis", 400)
			return
		}

		if entry.CustomHandle != "" {
			if err := rdb.Set(context.Background(), entry.CustomHandle, ":e:"+userid, 0).Err(); err != nil {
				errorPatch(c, err, "could not save old handle to redis", 400)
				return
			}
		}

		c.JSON(200, gin.H{
			"custom": json.Handle,
		})

	}
}

func DeleteCustomHandle(db *gorm.DB, auth *auth.Client, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		userid, inFirebase, err := user.GetUserID(auth, c)
		if err != nil {
			errorPatch(c, err, "failed to get jwt or header user id", 400)
			return
		}

		paying := false
		if inFirebase {
			paying, err = redisfn.CheckUserPaying(rdb, userid)
			if err != nil {
				errorPatch(c, err, "failed to correctly check status of user payment", 400)
				return
			}
		}

		if !paying {
			errorPatch(c, errors.New("must be a paying membership to modify custom handle"), "not a valid paying member", 400)
			return
		}

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

		if userid != entry.User {
			errorPatch(c, errors.New("unauthorized"), "Doesn't belong to that user for entry", 400)
			return
		}

		if err := UpdateCustomHandle(db, "", id10); err != nil {
			errorPatch(c, err, "could not save to db", 400)
			return
		}

		if err := rdb.Set(context.Background(), entry.CustomHandle, ":e:"+userid, 0).Err(); err != nil {
			errorPatch(c, err, "could not save old handle to redis", 400)
			return
		}

		c.Status(204)
	}
}
