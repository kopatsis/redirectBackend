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

var RESERVE = []int64{
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

func PostEntryDB(db *gorm.DB, entry *datatypes.Entry) (int64, error) {
	if err := db.Create(entry).Error; err != nil {
		return 0, err
	}
	return entry.ID, nil
}

func PostEntryFullDB(db *gorm.DB, entry *datatypes.Entry) (uniqueIss bool, actualErr error) {
	err := db.Create(entry).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return true, err
		} else {
			return false, err
		}
	}
	return false, nil
}

func GetTheNewID() (int64, string, error) {
	try := 0
	for try < 128 {
		attempt := int64(rand.Intn(convert.LIMIT-64+1) + 64)
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
	try := 0
	for try < 10 {
		id, param, err := GetTheNewID()
		if err != nil {
			try++
			continue
		}

		exists, err := rdb.Exists(context.Background(), param).Result()
		if exists > 0 || err != nil {
			try++
			continue
		}

		entry.ID = id

		if uniqueIssue, err := PostEntryFullDB(db, entry); err != nil {
			if uniqueIssue {
				try += 5
				continue
			} else {
				return "", err
			}
		}

		return param, nil
	}

	newID := RESERVE[rand.Intn(20)]
	st, err := convert.ToSixFour(newID)
	if err != nil {
		return "", err
	}

	if _, err := PostEntryFullDB(db, entry); err != nil {
		return "", err
	}

	return st, nil

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

		sixFour, err := AttemptToPost(db, rdb, &entry)
		if err != nil {
			errorPost(c, err, "Could not post to db")
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
