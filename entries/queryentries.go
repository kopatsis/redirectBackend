package entries

import (
	"c361main/convert"
	"c361main/datatypes"
	"c361main/user"
	"strconv"
	"strings"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"gorm.io/gorm"
)

func GetQueriedDB(db *gorm.DB, user, sort string, page int) ([]datatypes.ShortenedEntry, int, error) {
	var entries []datatypes.Entry
	var shortenedEntries []datatypes.ShortenedEntry

	query := db.Where("user = ? AND archived = ?", user, false)

	switch sort {
	case "aa":
		query = query.Order("real_url ASC")
	case "ad":
		query = query.Order("real_url DESC")
	case "da":
		query = query.Order("date ASC")
	case "ca":
		query = query.Order("count ASC")
	case "cd":
		query = query.Order("count DESC")
	default:
		query = query.Order("date DESC")
	}

	if page > 0 {
		query = query.Offset((page - 1) * datatypes.BATCH)
		query = query.Limit(datatypes.BATCH + 1)
	}

	if err := query.Find(&entries).Error; err != nil {
		return nil, 0, err
	}

	if len(entries) == 0 && page > 1 {
		page = 1
		query = query.Offset(-1)
		if err := query.Find(&entries).Error; err != nil {
			return nil, 0, err
		}
	}

	for _, entry := range entries {

		param, err := convert.ToSixFour(entry.ID)
		if err != nil {
			return nil, 0, err
		}

		shortenedEntry := datatypes.ShortenedEntry{
			Param:   param,
			User:    entry.User,
			RealURL: entry.RealURL,
			Date:    entry.Date,
		}
		shortenedEntries = append(shortenedEntries, shortenedEntry)
	}

	return shortenedEntries, page, nil
}

func SearchFilterEntries(db *gorm.DB, user, search, sort string, page int) ([]datatypes.ShortenedEntry, int, error) {
	entries, _, err := GetQueriedDB(db, user, sort, 0)
	if err != nil {
		return []datatypes.ShortenedEntry{}, 0, err
	}

	var filteredEntries []datatypes.ShortenedEntry

	for _, entry := range entries {
		if !strings.Contains(strings.ToLower(entry.Param), strings.ToLower(search)) && !fuzzy.MatchFold(search, entry.RealURL) {
			continue
		}

		filteredEntries = append(filteredEntries, entry)
	}

	skips := (page - 1) * datatypes.BATCH

	returnEntries := filteredEntries[min(skips, len(filteredEntries)):min(skips+datatypes.BATCH+1, len(filteredEntries))]

	if page > 1 && len(returnEntries) == 0 {
		returnEntries = filteredEntries[0:min(datatypes.BATCH+1, len(filteredEntries))]
		page = 1
	}

	return returnEntries, page, nil
}

func QueryEntries(app *firebase.App, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		userid, _, err := user.GetUserID(app, c)
		if err != nil {
			errorGet(c, err, "failed to get jwt or header user id")
			return
		}

		page, err := strconv.Atoi(c.DefaultQuery("p", "1"))
		if err != nil || page <= 0 {
			page = 1
		}

		search := c.DefaultQuery("q", "")
		if len(search) > 128 {
			search = search[0:127]
		}

		sort := c.DefaultQuery("s", "dd")
		if sort != "aa" && sort != "ad" && sort != "da" && sort != "dd" && sort != "ca" && sort != "cd" {
			sort = "dd"
		}

		var filteredEntries []datatypes.ShortenedEntry

		if search != "" {
			filteredEntries, page, err = SearchFilterEntries(db, userid, search, sort, page)
			if err != nil {
				errorGet(c, err, "failed to query actual entries")
				return
			}

		} else {
			filteredEntries, page, err = GetQueriedDB(db, userid, sort, page)
			if err != nil {
				errorGet(c, err, "failed to query actual entries")
				return
			}
		}

		more := len(filteredEntries) == 26

		if len(filteredEntries) > 0 {
			filteredEntries = filteredEntries[:len(filteredEntries)-1]
		}

		c.JSON(200, gin.H{
			"entries": &filteredEntries,
			"more":    more,
			"less":    page > 1,
			"page":    page,
			"search":  search,
			"sort":    sort,
		})

	}
}
