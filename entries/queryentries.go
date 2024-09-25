package entries

import (
	"c361main/convert"
	"c361main/datatypes"
	"c361main/payment/redisfn"
	"c361main/user"
	"errors"
	"strconv"
	"strings"
	"sync"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"gorm.io/gorm"
)

func GetSingleEntryDB(db *gorm.DB, user string, id int64) (datatypes.ShortenedEntry, error, bool) {
	var entry datatypes.Entry
	err := db.First(&entry, id).Error
	if err != nil {
		return datatypes.ShortenedEntry{}, err, false
	} else if entry.User != user {
		return datatypes.ShortenedEntry{}, errors.New("entry did not belong to user"), false
	} else if entry.Archived {
		return datatypes.ShortenedEntry{}, errors.New("entry archived"), true
	}

	param, err := convert.ToSixFour(entry.ID)
	if err != nil {
		return datatypes.ShortenedEntry{}, err, false
	}

	ret := datatypes.ShortenedEntry{
		Param:        param,
		User:         entry.User,
		RealURL:      entry.RealURL,
		Date:         entry.Date,
		Count:        entry.Count,
		CustomHandle: entry.CustomHandle,
	}

	return ret, nil, false
}

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
			Count:   entry.Count,
		}
		shortenedEntries = append(shortenedEntries, shortenedEntry)
	}

	return shortenedEntries, page, nil
}

func SearchFilterEntries(db *gorm.DB, user, search, sort string, page int, paying bool) ([]datatypes.ShortenedEntry, int, error) {
	entries, _, err := GetQueriedDB(db, user, sort, 0)
	if err != nil {
		return []datatypes.ShortenedEntry{}, 0, err
	}

	var filteredEntries []datatypes.ShortenedEntry

	for _, entry := range entries {
		if paying {
			if !strings.Contains(strings.ToLower(entry.Param), strings.ToLower(search)) && !fuzzy.MatchFold(search, entry.RealURL) && !fuzzy.MatchFold(search, entry.CustomHandle) {
				continue
			}
		} else {
			if !strings.Contains(strings.ToLower(entry.Param), strings.ToLower(search)) && !fuzzy.MatchFold(search, entry.RealURL) {
				continue
			}
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

func QueryEntriesShared(db *gorm.DB, c *gin.Context, userid string, paying bool) (datatypes.EntryList, string, error) {

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
		filteredEntries, page, err = SearchFilterEntries(db, userid, search, sort, page, paying)
		if err != nil {
			return datatypes.EntryList{}, "failed to query actual entries", err
		}

	} else {
		filteredEntries, page, err = GetQueriedDB(db, userid, sort, page)
		if err != nil {
			return datatypes.EntryList{}, "failed to query actual entries", err
		}
	}

	more := len(filteredEntries) == 26

	if more {
		filteredEntries = filteredEntries[:len(filteredEntries)-1]
	}

	ret := datatypes.EntryList{
		FilteredEntries: filteredEntries,
		More:            more,
		Less:            page > 1,
		Page:            page,
		Search:          search,
		Sort:            sort,
	}

	return ret, "", nil
}

func QueryEntries(auth *auth.Client, db *gorm.DB, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

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

		response, reason, err := QueryEntriesShared(db, c, userid, paying)
		if err != nil {
			errorGet(c, err, reason)
		}

		if !paying {
			for i, e := range response.FilteredEntries {
				e.CustomHandle = ""
				response.FilteredEntries[i] = e
			}
		}

		c.JSON(200, gin.H{
			"response": response,
		})

	}
}

func QueryEntriesWithSingle(auth *auth.Client, db *gorm.DB, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		id10, err := convert.FromSixFour(c.Param("id"))
		if err != nil {
			errorGet(c, err, "Failed to convert id param")
			return
		}

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

		var wg sync.WaitGroup

		var entriesResult datatypes.EntryList
		var entriesErr error
		var entriesMsg string

		var entryResult datatypes.ShortenedEntry
		var entryErr error
		var entryArchived bool

		wg.Add(1)
		go func() {
			defer wg.Done()
			entriesResult, entriesMsg, entriesErr = QueryEntriesShared(db, c, userid, paying)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			entryResult, entryErr, entryArchived = GetSingleEntryDB(db, userid, id10)
		}()

		wg.Wait()

		if entriesErr != nil {
			errorGet(c, entriesErr, entriesMsg)
			return
		}

		if !paying {
			entryResult.CustomHandle = ""
			for i, e := range entriesResult.FilteredEntries {
				e.CustomHandle = ""
				entriesResult.FilteredEntries[i] = e
			}
		}

		response := gin.H{
			"error":    "",
			"archived": entryArchived,
			"entry":    entryResult,
			"response": entriesResult,
		}

		if entryErr != nil {
			response["error"] = entryErr.Error()
		}

		c.JSON(200, response)

	}
}
