package datatypes

import (
	"strings"
	"time"
)

const BATCH = 20

type Entry struct {
	ID           int        `gorm:"primaryKey;autoIncrement" json:"-"`
	Param        string     `gorm:"index;unique" json:"handle"`
	User         string     `gorm:"index" json:"user"`
	RealURL      string     `json:"url"`
	Custom       bool       `json:"custom"`
	Count        int        `json:"-"`
	Archived     bool       `json:"-"`
	Date         time.Time  `json:"-"`
	ArchivedDate *time.Time `json:"-"`
}

type ShortenedEntry struct {
	Param   string    `json:"param"`
	User    string    `json:"user"`
	RealURL string    `json:"url"`
	Date    time.Time `json:"date"`
	Count   int       `json:"count"`
	Custom  bool      `json:"custom"`
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
