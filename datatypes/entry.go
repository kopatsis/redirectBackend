package datatypes

import (
	"strings"
	"time"
)

const BATCH = 2

type Entry struct {
	ID           int64     `gorm:"primaryKey;autoIncrement:false" json:"-"`
	User         string    `gorm:"index" json:"user"`
	RealURL      string    `json:"url"`
	CustomHandle string    `gorm:"index" json:"-"`
	Count        int       `json:"-"`
	Archived     bool      `json:"-"`
	Date         time.Time `json:"-"`
	ArchivedDate time.Time `json:"-"`
}

type ShortenedEntry struct {
	Param        string    `json:"param"`
	User         string    `json:"user"`
	RealURL      string    `json:"url"`
	Date         time.Time `json:"date"`
	Count        int       `json:"count"`
	CustomHandle string    `json:"custom"`
}

type EntryList struct {
	FilteredEntries []ShortenedEntry `json:"entries"`
	More            bool             `json:"more"`
	Less            bool             `json:"less"`
	Page            int              `json:"page"`
	Search          string           `json:"search"`
	Sort            string           `json:"sort"`
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
