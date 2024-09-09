package datatypes

import (
	"strings"
	"time"
)

const BATCH = 25

type Entry struct {
	ID           int       `gorm:"primaryKey" json:"-"`
	User         string    `gorm:"index" json:"user"`
	RealURL      string    `json:"url"`
	CustomHandle string    `gorm:"unique;index" json:"-"`
	Count        int       `json:"-"`
	Archived     bool      `json:"-"`
	Date         time.Time `json:"-"`
	ArchivedDate time.Time `json:"-"`
}

type ShortenedEntry struct {
	Param   string    `json:"param"`
	User    string    `json:"user"`
	RealURL string    `json:"url"`
	Date    time.Time `json:"date"`
}

type LengthenedEntry struct {
	Param   string    `json:"param"`
	User    string    `json:"user"`
	RealURL string    `json:"url"`
	Date    time.Time `json:"date"`
	Count   int64     `json:"count"`
}

func (entry *Entry) InitalizeFormat() {
	entry.Date = time.Now()
	entry.Archived = false
	entry.RealURL = EnsureHttpPrefix(entry.RealURL)
}

func EnsureHttpPrefix(url string) string {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "https://" + url
	}
	return url
}
