package entry

import (
	"bytes"
	"c361main/convert"
	"c361main/datatypes"
	"c361main/user"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"math/rand"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
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

func IsValidURL(rawURL string) bool {
	urlPattern := `^(http|https)://((localhost|[0-9]{1,3}(\.[0-9]{1,3}){3})|([a-zA-Z0-9.-]+\.[a-zA-Z]{2,}))(:[0-9]+)?(/.*)?$`
	matched, _ := regexp.MatchString(urlPattern, rawURL)
	return matched
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

		if slices.Contains(RESERVE, attempt) {
			try++
			continue
		}

		return attempt, st, nil
	}
	return 0, "", errors.New("unable to generate a number within constraints")
}

func AttemptToPost(db *gorm.DB, rdb *redis.Client, httpClient *http.Client, entry *datatypes.Entry) (string, error) {
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
		if err := ErrorAlertEmail(httpClient, newID, true); err != nil {
			log.Println("Couldn't send error alert email for reserve fail: " + err.Error())
		}
		return "", err
	}

	if err := ErrorAlertEmail(httpClient, newID, false); err != nil {
		log.Println("Couldn't send error alert email for reserve success: " + err.Error())
	}
	return st, nil
}

func ErrorAlertEmail(httpClient *http.Client, id int64, failed bool) error {
	idSt := strconv.FormatInt(id, 10)

	subject := "HAD TO USE RESERVE FOR ADD ID"
	if failed {
		subject = "FAILED TO USE RESERVE FOR ADD ID"
	}

	body := "ID :" + idSt + "\nPlease redo the RESERVED and redeploy ASAP"

	passcode := os.Getenv("CHECK_PASSCODE")
	if passcode == "" {
		return errors.New("no passcode exists in the environment")
	}

	payload := map[string]string{
		"subject": subject,
		"body":    body,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	checkURL := os.Getenv("PAY_API_URL")
	if checkURL == "" {
		checkURL = "https://pay.shortentrack.com"
	}

	checkURL += "/administrative/internalemail"

	req, err := http.NewRequest("POST", checkURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	req.Header.Set("X-Passcode-ID", passcode)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to send email alert")
	}

	return nil
}

func PostEntry(db *gorm.DB, rdb *redis.Client, app *firebase.App, httpClient *http.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		userid, _, err := user.GetUserID(app, c)
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

		sixFour, err := AttemptToPost(db, rdb, httpClient, &entry)
		if err != nil {
			errorPost(c, err, "Could not post to db")
			return
		}

		if err := rdb.Set(context.Background(), sixFour, entry.RealURL, 0).Err(); err != nil {
			errorPost(c, err, "Could not post to redis")
			return
		}

		c.JSON(201, gin.H{
			"parameter": sixFour,
		})
	}
}
