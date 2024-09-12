package entry

import (
	"c361main/convert"
	"c361main/datatypes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"math/rand"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/martian/v3/log"
	"gorm.io/gorm"
)

var RESERVE = []int{
	124092490, 886968310, 210759517, 635537027, 716492954, 714163706,
	1022127667, 1039898767, 548615800, 873858384, 1025089867, 318172894,
	911652002, 369912434, 562716951, 975542319, 209260992, 1048199652,
	1031340735, 768628033,
}

var FILTER = []string{
	"shit", "fuck", "bitch", "cunt", "fag", "whore", "dick", "ass", "nigga", "tard",
}

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

func PostEntryFullDB(db *gorm.DB, entry *datatypes.Entry) error {
	return db.Create(entry).Error
}

func GetTheNewID() (int, string, error) {
	try := 0
	for try < 128 {
		attempt := rand.Intn(1073741822-64+1) + 64
		st, err := convert.ToSixFour(attempt)
		if err != nil {
			try++
			continue
		}
		lower := strings.ToLower(st)
		for _, word := range FILTER {
			if strings.Contains(word, lower) {
				try++
				continue
			}
		}
		return attempt, st, nil
	}
	return 0, "", errors.New("unable to generate a number within constraints")
}

func AttemptToPost(db *gorm.DB, rdb *redis.Client, entry *datatypes.Entry) (string, error) {
	for i := 0; i < 10; i++ {
		id, param, err := GetTheNewID()
		if err != nil {
			continue
		}

		exists, err := rdb.Exists(context.Background(), param).Result()
		if exists > 0 || err != nil {
			continue
		}

		entry.ID = id

		if err := PostEntryFullDB(db, entry); err != nil {
			continue
		}

		return param, nil
	}

	return "", errors.New("unable to create")
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
