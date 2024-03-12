package routes

import (
	"fmt"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/gin-gonic/gin"
)

func countClicks(clicks []Click, start time.Time, end time.Time) int {
	count := 0
	for _, click := range clicks {
		if click.Date.After(start) && click.Date.Before(end) {
			count++
		}
	}
	return count
}

func GetClicksHourly(client *datastore.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		param := c.Param("id")

		var clicks []Click

		startOfThisHour := time.Now().Truncate(time.Hour)
		startTime := startOfThisHour.Add(-23 * time.Hour)

		query := datastore.NewQuery("Click").FilterField("Param", "=", param).FilterField("Date", ">=", startTime)
		_, err := client.GetAll(c, query, &clicks)
		if err != nil {
			fmt.Printf("Failed to get all: %v", err)
			c.JSON(400, gin.H{
				"Error Type":  "Failed to get hourly clicks for provided param",
				"Exact Error": err.Error(),
			})
			return
		}

		timeKeyMap := make(map[string]int)

		for i := 0; i < 24; i++ {
			currentTime := startTime.Add(time.Duration(i) * time.Hour)
			endTime := currentTime.Add(time.Hour)
			count := countClicks(clicks, currentTime, endTime)
			timeKeyMap[currentTime.Format(time.RFC3339)] = count
		}

		c.JSON(200, timeKeyMap)
	}
}

func GetClicksDaily(client *datastore.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		param := c.Param("id")

		var clicks []Click

		startOfThisDay := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local)
		startTime := startOfThisDay.AddDate(0, 0, -6)

		query := datastore.NewQuery("Click").FilterField("Param", "=", param).FilterField("Date", ">=", startTime)
		_, err := client.GetAll(c, query, &clicks)
		if err != nil {
			fmt.Printf("Failed to get all: %v", err)
			c.JSON(400, gin.H{
				"Error Type":  "Failed to get daily clicks for provided param",
				"Exact Error": err.Error(),
			})
			return
		}

		timeKeyMap := make(map[string]int)

		for i := 0; i < 7; i++ {
			currentTime := startTime.AddDate(0, 0, i)
			endTime := currentTime.AddDate(0, 0, 1)
			count := countClicks(clicks, currentTime, endTime)
			timeKeyMap[currentTime.Format(time.RFC3339)] = count
		}

		c.JSON(200, timeKeyMap)
	}
}

func GetClicksWeekly(client *datastore.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		param := c.Param("id")

		var clicks []Click

		startOfThisDay := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local)
		startOfThisWeek := startOfThisDay.AddDate(0, 0, -6)
		startTime := startOfThisWeek.AddDate(0, 0, -28)

		query := datastore.NewQuery("Click").FilterField("Param", "=", param).FilterField("Date", ">=", startTime)
		_, err := client.GetAll(c, query, &clicks)
		if err != nil {
			fmt.Printf("Failed to get all: %v", err)
			c.JSON(400, gin.H{
				"Error Type":  "Failed to get monthly clicks for provided param",
				"Exact Error": err.Error(),
			})
			return
		}

		timeKeyMap := make(map[string]int)

		for i := 0; i < 5; i++ {
			currentTime := startTime.AddDate(0, 0, i*7)
			endTime := currentTime.AddDate(0, 0, 7)
			count := countClicks(clicks, currentTime, endTime)
			timeKeyMap[currentTime.Format(time.RFC3339)] = count
		}

		c.JSON(200, timeKeyMap)
	}
}
