package datatypes

import (
	"time"
)

type Click struct {
	ID        int `gorm:"primaryKey"`
	ParamKey  int `gorm:"index"`
	Time      time.Time
	RealURL   string
	City      string
	Country   string
	Browser   string
	OS        string
	Platform  string
	Mobile    bool
	Bot       bool
	FromQR    bool
	IPAddress string
}

type ClickDataPaid struct {
	Param          string      `json:"param"`
	User           string      `json:"user"`
	RealURL        string      `json:"realUrl"`
	Total          int         `json:"total"`
	FromQR         int         `json:"fromQr"`
	FromBot        int         `json:"fromBot"`
	FromMobile     int         `json:"fromMobile"`
	UniqueVisits   int         `json:"uniqueVisits"`
	DailyGraph     DateGraph   `json:"dailyGraph"`
	WeeklyGraph    DateGraph   `json:"weeklyGraph"`
	CountryGraph   StringGraph `json:"countryGraph"`
	CityGraph      StringGraph `json:"cityGraph"`
	BrowserGraph   StringGraph `json:"browserGraph"`
	OperatingGraph StringGraph `json:"operatingGraph"`
}

type ClickDataFree struct {
	Param        string      `json:"param"`
	User         string      `json:"user"`
	RealURL      string      `json:"realUrl"`
	Total        int         `json:"total"`
	FromQR       int         `json:"fromQr"`
	WeeklyGraph  DateGraph   `json:"weeklyGraph"`
	BrowserGraph StringGraph `json:"browserGraph"`
}

type DateGraph struct {
	Keys []time.Time `json:"keys"`
	Data []int       `json:"data"`
}

type StringGraph struct {
	Keys []string `json:"keys"`
	Data []int    `json:"data"`
}
