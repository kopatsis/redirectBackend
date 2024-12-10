package entry

import (
	"c361main/convert"
	"c361main/datatypes"
	"c361main/specialty/sendgridfn"
	"c361main/user"
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"math/rand"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sendgrid/sendgrid-go"
	"gorm.io/gorm"
)

var RESERVE = []int64{
	124092490, 886968310, 210759517, 635537027, 716492954, 714163706,
	1022127667, 1039898767, 548615800, 873858384, 1025089867, 318172894,
	911652002, 369912434, 562716951, 975542319, 209260992, 1048199652,
	1031340735, 768628033,
}

var FILTER = []string{
	"shit", "fuck", "bitch", "cunt", "fag", "whore", "dick", "ass", "nigga", "tard", "nig",
}

// Sets error message for post request using gin Context
func errorPost(c *gin.Context, err error, reason string) {
	fmt.Printf("Failed to post: %v", err)
	c.JSON(400, gin.H{
		"Error Type":  reason,
		"Exact Error": err.Error(),
	})
}

func IsValidURL(rawURL string) bool {
	urlPattern := `^(http|https)://((localhost|[0-9]{1,3}(\.[0-9]{1,3}){3})|([a-zA-Z0-9.-]+\.[a-zA-Z]{2,}))(:[0-9]+)?(/.*)?$`
	matched, _ := regexp.MatchString(urlPattern, rawURL)
	return matched
}

func PostEntryFullDB(db *gorm.DB, entry *datatypes.Entry) (id int, uniqueIss bool, actualErr error) {
	err := db.Create(entry).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return 0, true, err
		} else {
			return 0, false, err
		}
	}
	return entry.ID, false, nil
}

func GetTheNewID() (string, error) {
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

		if slices.Contains(RESERVE, attempt) {
			try++
			continue
		}

		return st, nil
	}
	return "", errors.New("unable to generate a number within constraints")
}

func AttemptToPost(db *gorm.DB, rdb *redis.Client, sendgridClient *sendgrid.Client, entry *datatypes.Entry) (string, int, error) {
	try := 0

	var idRet int
	for try < 10 {
		param, err := GetTheNewID()
		if err != nil {
			try++
			continue
		}

		exists, err := rdb.Exists(context.Background(), param).Result()
		if exists > 0 || err != nil {
			try++
			continue
		}

		entry.Param = param

		if id, uniqueIssue, err := PostEntryFullDB(db, entry); err != nil {
			if uniqueIssue {
				try += 5
				continue
			} else {
				return "", 0, err
			}
		} else {
			idRet = id
		}

		return param, idRet, nil
	}

	newID := RESERVE[rand.Intn(20)]
	st, err := convert.ToSixFour(newID)
	if err != nil {
		return "", 0, err
	}

	if id, _, err := PostEntryFullDB(db, entry); err != nil {
		if err := ErrorAlertEmail(sendgridClient, newID, true); err != nil {
			log.Println("Couldn't send error alert email for reserve fail: " + err.Error())
		}
		return "", 0, err
	} else {
		idRet = id
	}

	if err := ErrorAlertEmail(sendgridClient, newID, false); err != nil {
		log.Println("Couldn't send error alert email for reserve success: " + err.Error())
	}
	return st, idRet, nil
}

func ErrorAlertEmail(sendgridClient *sendgrid.Client, id int64, failed bool) error {
	idSt := strconv.FormatInt(id, 10)

	subject := "HAD TO USE RESERVE FOR ADD ID"
	if failed {
		subject = "FAILED TO USE RESERVE FOR ADD ID"
	}

	body := "ID :" + idSt + "\nPlease redo the RESERVED and redeploy ASAP"

	return sendgridfn.SendSeriousErrorAlert(sendgridClient, subject, body)
}

func PostEntry(db *gorm.DB, rdb *redis.Client, auth *auth.Client, sendgridClient *sendgrid.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		userid, _, err := user.GetUserID(auth, c)
		if err != nil {
			errorPatch(c, err, "Failed to get user id", 400)
			return
		}

		var entry datatypes.Entry

		if err := c.ShouldBindJSON(&entry); err != nil {
			errorPost(c, err, "Failed to bind entry body")
			return
		}

		entry.User = userid

		entry.InitalizeFormat()

		if !IsValidURL(entry.RealURL) {
			errorPost(c, errors.New("not real url"), "Not a URL that can be parsed")
			return
		}

		if entry.Custom && entry.Param == "" {
			errorPost(c, errors.New("custom handle missing"), "Error with custom Handle")
			return
		}
		if entry.Custom && len(entry.Param) < 7 || len(entry.Param) > 128 {
			errorPost(c, errors.New("custom handle length incorrect"), "Error with custom Handle")
			return
		}
		if entry.Custom && !regexp.MustCompile(`^[a-zA-Z0-9_-]*$`).MatchString(entry.Param) {
			errorPost(c, errors.New("custom handle chars incorrect"), "Error with custom Handle")
			return
		}
		if !entry.Custom && entry.Param != "" {
			errorPost(c, errors.New("custom bool field incorrect"), "Error with custom Handle")
			return
		}

		var sixFour string
		var retID int
		if entry.Custom {
			newID, uniqueIssue, err := PostEntryFullDB(db, &entry)
			if err != nil {
				if uniqueIssue {
					errorPost(c, err, "Not a unique custom handle")
				} else {
					errorPost(c, err, "Error other than unique custom handle")
				}
				return
			}
			sixFour = entry.Param
			if err := rdb.Set(context.Background(), sixFour, CreateCustomHandleStruct(entry.RealURL, userid, newID), 0).Err(); err != nil {
				errorPost(c, err, "Could not post to redis")
				return
			}
			retID = newID
		} else {
			var newID int
			sixFour, newID, err = AttemptToPost(db, rdb, sendgridClient, &entry)
			if err != nil {
				errorPost(c, err, "Could not post to db")
				return
			}
			if err := rdb.Set(context.Background(), sixFour, CreateCustomHandleStruct(entry.RealURL, userid, newID), 0).Err(); err != nil {
				errorPost(c, err, "Could not post to redis")
				return
			}
			retID = newID
		}

		c.JSON(201, gin.H{
			"id": retID,
		})
	}
}
