package clicks

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

func GetJustClicksDB(db *gorm.DB, paramkey int64, user string) ([]datatypes.Click, error, bool) {
	err := db.Model(&datatypes.Entry{}).Where("id = ? AND user = ?", paramkey, user).Select("id").First(&datatypes.Entry{}).Error
	if err != nil {
		return nil, err, true
	}

	var clicks []datatypes.Click
	err = db.Where("param_key = ?", paramkey).Find(&clicks).Error
	if err != nil {
		return nil, err, false
	}

	return clicks, nil, false
}

func AnonymizeIPAddresses(allClicks []datatypes.Click) {
	ipSet := map[string]int{}
	for _, click := range allClicks {
		if _, exists := ipSet[click.IPAddress]; !exists {
			ipSet[click.IPAddress] = len(ipSet) + 1
		}
	}

	for i, click := range allClicks {
		number := ipSet[click.IPAddress]
		allClicks[i].IPAddress = "Unique Visitor " + strconv.Itoa(number)
	}
}

func ServeClickCSV(c *gin.Context, clicks []datatypes.Click, realParam string) {
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)

	shortDomain := os.Getenv("SHORT_DOMAIN")
	headers := []string{
		"Shortened URL ID", "Shortened URL", "Time", "Real URL", "City", "Country", "Browser", "OS", "Platform", "Mobile", "Bot", "From QR Code", "From Custom Handle", "Visitor Number",
	}
	writer.Write(headers)

	for _, click := range clicks {
		record := []string{
			realParam,
			shortDomain + "/" + click.Handle,
			click.Time.Format(time.RFC3339),
			click.RealURL,
			click.City,
			click.Country,
			click.Browser,
			click.OS,
			click.Platform,
			strconv.FormatBool(click.Mobile),
			strconv.FormatBool(click.Bot),
			strconv.FormatBool(click.FromQR),
			strconv.FormatBool(click.FromCustom),
			click.IPAddress,
		}

		writer.Write(record)
	}

	writer.Flush()

	filename := "click_data_" + realParam + ".csv"
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "text/csv")
	c.Writer.Write(buffer.Bytes())
}

func GetClickCSV(db *gorm.DB, auth *auth.Client, rdb *redis.Client) gin.HandlerFunc {
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

		id10, err := convert.FromSixFour(c.Param("id"))
		if err != nil {
			errorGet(c, err, "Failed to convert id param")
			return
		}

		allClicks, err, userIssue := GetJustClicksDB(db, id10, userid)
		if err != nil {
			if userIssue {
				errorGet(c, err, "Failed to get entry that matches ID and user for clicks from DB")
			} else {
				errorGet(c, err, "Failed to actually retrieve the clicks from DB")
			}
			return
		}

		AnonymizeIPAddresses(allClicks)
		ServeClickCSV(c, allClicks, c.Param("id"))

		elapsed := time.Since(startTimer)
		if elapsed < 12*time.Second {
			time.Sleep(12*time.Second - elapsed)
		}
	}
}
