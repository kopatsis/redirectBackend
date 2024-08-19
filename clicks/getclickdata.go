package clicks

import (
	"c361main/convert"
	"c361main/datatypes"
	"c361main/user"
	"fmt"
	"sort"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type kv struct {
	Key   string
	Value int
}

func errorGet(c *gin.Context, err error, reason string) {
	fmt.Printf("Failed to get: %v", err)
	c.JSON(400, gin.H{
		"Error Type":  reason,
		"Exact Error": err.Error(),
	})
}

func GetClicksDB(db *gorm.DB, paramkey int, user string) ([]datatypes.Click, string, error, bool) {
	var entry struct {
		ID      int
		RealURL string
	}

	err := db.Model(&datatypes.Entry{}).Where("id = ? AND user = ?", paramkey, user).Select("id, real_url").First(&entry).Error
	if err != nil {
		return nil, "", err, true
	}

	var clicks []datatypes.Click
	err = db.Where("param_key = ?", paramkey).Find(&clicks).Error
	if err != nil {
		return nil, "", err, false
	}

	return clicks, entry.RealURL, nil, false
}

func WeeklyDateFixer(click, start time.Time) (time.Time, bool) {
	if click.Before(start.AddDate(0, 0, -42)) {
		return click, false
	}
	elapsed := start.Sub(click)
	periods := (elapsed + (7 * 24 * time.Hour) - 1) / (7 * 24 * time.Hour)
	startOfPeriod := start.Add(-periods * 7 * 24 * time.Hour)
	return startOfPeriod, true
}

func DailyDateFixer(click, start time.Time) (time.Time, bool) {
	if click.Before(start.AddDate(0, 0, -7)) {
		return click, false
	}
	elapsed := start.Sub(click)
	periods := (elapsed + (7 * 24 * time.Hour) - 1) / (7 * 24 * time.Hour)
	startOfPeriod := start.Add(-periods * 7 * 24 * time.Hour)
	return startOfPeriod, true
}

func ProcessWeeklyGraph(dateMap map[time.Time]int, start time.Time) datatypes.DateGraph {
	var graph datatypes.DateGraph

	for i := 5; i >= 0; i-- {
		weekStart := start.AddDate(0, 0, -7*(i+1))

		graph.Keys = append(graph.Keys, weekStart)

		if value, exists := dateMap[weekStart]; exists {
			graph.Data = append(graph.Data, value)
		} else {
			graph.Data = append(graph.Data, 0)
		}
	}

	return graph
}

func ProcessDailyGraph(dateMap map[time.Time]int, start time.Time) datatypes.DateGraph {
	var graph datatypes.DateGraph

	for i := 6; i >= 0; i-- {
		dayStart := start.AddDate(0, 0, -i)

		graph.Keys = append(graph.Keys, dayStart)

		if value, exists := dateMap[dayStart]; exists {
			graph.Data = append(graph.Data, value)
		} else {
			graph.Data = append(graph.Data, 0)
		}
	}

	return graph
}

func ProcessMaxGraph(strMap map[string]int, max int) datatypes.StringGraph {
	var graph datatypes.StringGraph

	var pairs []kv
	for k, v := range strMap {
		pairs = append(pairs, kv{k, v})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Value > pairs[j].Value
	})

	if len(pairs) > max {
		otherTotal := 0
		for i, pair := range pairs {
			if i < max-1 {
				graph.Keys = append(graph.Keys, pair.Key)
				graph.Data = append(graph.Data, pair.Value)
			} else {
				otherTotal += pair.Value
			}
		}
		graph.Keys = append(graph.Keys, "Other")
		graph.Data = append(graph.Data, otherTotal)
	} else {
		for _, pair := range pairs {
			graph.Keys = append(graph.Keys, pair.Key)
			graph.Data = append(graph.Data, pair.Value)
		}
	}

	return graph
}

func ProcessClicksPaid(clicks []datatypes.Click, param, realURL, userID string) datatypes.ClickDataPaid {
	ret := datatypes.ClickDataPaid{
		Param:   param,
		RealURL: realURL,
		User:    userID,
	}

	total, fromqr, frombot, frommobile := 0, 0, 0, 0
	dailyMap, weeklyMap := map[time.Time]int{}, map[time.Time]int{}
	browserMap, operatingMap, cityMap, countryMap := map[string]int{}, map[string]int{}, map[string]int{}, map[string]int{}
	ipSet := map[string]bool{}

	startTime := time.Now()
	for _, click := range clicks {
		total++

		if click.FromQR {
			fromqr++
		}
		if click.Bot {
			frombot++
		}
		if click.Mobile {
			frommobile++
		}

		date, inMap := WeeklyDateFixer(click.Time, startTime)
		if inMap {
			if count, exists := weeklyMap[date]; exists {
				weeklyMap[date] = count + 1
			} else {
				weeklyMap[date] = 1
			}
		}

		date, inMap = DailyDateFixer(click.Time, startTime)
		if inMap {
			if count, exists := dailyMap[date]; exists {
				dailyMap[date] = count + 1
			} else {
				dailyMap[date] = 1
			}
		}

		if count, exists := browserMap[click.Browser]; exists {
			browserMap[click.Browser] = count + 1
		} else {
			browserMap[click.Browser] = 1
		}

		if count, exists := operatingMap[click.OS]; exists {
			operatingMap[click.OS] = count + 1
		} else {
			operatingMap[click.OS] = 1
		}

		if count, exists := countryMap[click.Country]; exists {
			countryMap[click.Country] = count + 1
		} else {
			countryMap[click.Country] = 1
		}

		city := click.City + ", " + click.Country
		if count, exists := cityMap[city]; exists {
			cityMap[city] = count + 1
		} else {
			cityMap[city] = 1
		}

		ipSet[click.IPAddress] = true
	}

	ret.DailyGraph = ProcessDailyGraph(dailyMap, startTime)
	ret.WeeklyGraph = ProcessWeeklyGraph(weeklyMap, startTime)

	ret.BrowserGraph = ProcessMaxGraph(browserMap, 5)
	ret.OperatingGraph = ProcessMaxGraph(operatingMap, 5)
	ret.CountryGraph = ProcessMaxGraph(countryMap, 8)
	ret.CityGraph = ProcessMaxGraph(cityMap, 10)

	ret.Total = total
	ret.FromQR = fromqr
	ret.FromBot = frombot
	ret.FromMobile = frommobile

	ret.UniqueVisits = len(ipSet)

	return ret
}

func ProcessClicksFree(clicks []datatypes.Click, param, realURL, userID string) datatypes.ClickDataFree {

	ret := datatypes.ClickDataFree{
		Param:   param,
		RealURL: realURL,
		User:    userID,
	}

	total, fromqr := 0, 0
	dateMap, browserMap := map[time.Time]int{}, map[string]int{}
	startTime := time.Now()
	for _, click := range clicks {
		total++
		if click.FromQR {
			fromqr++
		}

		date, inMap := WeeklyDateFixer(click.Time, startTime)
		if inMap {
			if count, exists := dateMap[date]; exists {
				dateMap[date] = count + 1
			} else {
				dateMap[date] = 1
			}
		}

		if count, exists := browserMap[click.Browser]; exists {
			browserMap[click.Browser] = count + 1
		} else {
			browserMap[click.Browser] = 1
		}
	}

	ret.WeeklyGraph = ProcessWeeklyGraph(dateMap, startTime)
	ret.BrowserGraph = ProcessMaxGraph(browserMap, 5)

	ret.Total = total
	ret.FromQR = fromqr

	return ret
}

func GetClicksByParam(db *gorm.DB, firebase *firebase.App) gin.HandlerFunc {
	return func(c *gin.Context) {

		startTimer := time.Now()

		userid, inFirebase, err := user.GetUserID(c)
		if err != nil {
			errorGet(c, err, "failed to get jwt or header user id")
			return
		}

		paying := false
		if inFirebase {
			isPaying, isValid := user.CheckTokenAndPaymentStatus(firebase, c)
			if !isValid {
				errorGet(c, err, "not a valid firebase token")
				return
			}
			paying = isPaying
		}

		id10, err := convert.FromSixFour(c.Param("id"))
		if err != nil {
			errorGet(c, err, "Failed to convert id param")
			return
		}

		allClicks, realURL, err, userIssue := GetClicksDB(db, id10, userid)
		if err != nil {
			if userIssue {
				errorGet(c, err, "Failed to get entry that matches ID and user for clicks from DB")
			} else {
				errorGet(c, err, "Failed to actually retrieve the clicks from DB")
			}
			return
		}

		if paying {
			data := ProcessClicksPaid(allClicks, c.Param("id"), realURL, userid)
			c.JSON(200, gin.H{
				"data": data,
			})
		} else {
			data := ProcessClicksFree(allClicks, c.Param("id"), realURL, userid)
			c.JSON(200, gin.H{
				"data": data,
			})
		}

		elapsed := time.Since(startTimer)
		if elapsed < 5*time.Second {
			time.Sleep(5*time.Second - elapsed)
		}
	}
}
