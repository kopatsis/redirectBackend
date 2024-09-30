package entries

import (
	"bytes"
	"c361main/convert"
	"c361main/datatypes"
	"c361main/payment/redisfn"
	"c361main/user"
	"encoding/csv"
	"errors"
	"os"
	"strconv"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// type ClickCount struct {
// 	ParamKey int
// 	Count    int64
// }

// func countClicksByParamKey(db *gorm.DB, entries []datatypes.Entry) ([]ClickCount, error) {

// 	var paramKeys []int
// 	for _, ent := range entries {
// 		paramKeys = append(paramKeys, ent.ID)
// 	}

// 	var result []ClickCount

// 	err := db.Model(&datatypes.Click{}).
// 		Select("param_key, COUNT(*) as count").
// 		Where("param_key IN ?", paramKeys).
// 		Group("param_key").
// 		Find(&result).Error
// 	if err != nil {
// 		return nil, err
// 	}
// 	return result, nil
// }

func GetEntriesRaw(db *gorm.DB, user string) ([]datatypes.Entry, error) {
	var entries []datatypes.Entry

	if err := db.Where("user = ? AND archived = ?", user, false).Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

func createLengthed(entries []datatypes.Entry) ([]datatypes.ShortenedEntry, error) {
	ret := []datatypes.ShortenedEntry{}

	for _, ent := range entries {
		param, err := convert.ToSixFour(ent.ID)
		if err != nil {
			return nil, err
		}

		current := datatypes.ShortenedEntry{
			Param:        param,
			User:         ent.User,
			RealURL:      ent.RealURL,
			Date:         ent.Date,
			Count:        ent.Count,
			CustomHandle: ent.CustomHandle,
		}
		ret = append(ret, current)
	}

	return ret, nil
}

func ServeEntriesCSV(c *gin.Context, entries []datatypes.ShortenedEntry) {
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)

	shortDomain := os.Getenv("SHORT_DOMAIN")
	headers := []string{
		"Shortened URL", "Date Created", "Real URL", "Click Count", "Custom Shortened URL",
	}
	writer.Write(headers)

	for _, ent := range entries {
		custom := ""
		if ent.CustomHandle != "" {
			custom = shortDomain + "/" + ent.CustomHandle
		}
		record := []string{
			shortDomain + "/" + ent.Param,
			ent.Date.Format("2006-01-02 15:04:05"),
			ent.RealURL,
			strconv.Itoa(ent.Count),
			custom,
		}

		writer.Write(record)
	}

	writer.Flush()

	filename := "entry_data.csv"
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "text/csv")
	c.Writer.Write(buffer.Bytes())
}

func GetEntriesCSV(db *gorm.DB, auth *auth.Client, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		startTimer := time.Now()

		userid, inFirebase, err := user.GetUserID(auth, c)
		if err != nil {
			errorGet(c, err, "failed to get jwt or header user id")
			return
		}

		paying := false
		if inFirebase {
			paying, err = redisfn.CheckUserPaying(rdb, userid)
			if err != nil {
				errorGet(c, err, "failed to correctly check status of user payment")
				return
			}
		}

		if !paying {
			errorGet(c, errors.New("must be a paying membership to get CSVs"), "not a valid paying member")
			return
		}

		entries, err := GetEntriesRaw(db, userid)
		if err != nil {
			errorGet(c, err, "unable to query entries")
			return
		}

		ret, err := createLengthed(entries)
		if err != nil {
			errorGet(c, err, "unable to convert to lengthed entry structs")
			return
		}

		ServeEntriesCSV(c, ret)

		elapsed := time.Since(startTimer)
		if elapsed < 6*time.Second {
			time.Sleep(6*time.Second - elapsed)
		}
	}
}
