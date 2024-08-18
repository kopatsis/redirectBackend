package datatypes

import (
	"strings"
	"time"
)

type Entry struct {
	ID           int       `gorm:"primaryKey" json:"-"`
	User         string    `json:"user"`
	RealURL      string    `json:"url"`
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
