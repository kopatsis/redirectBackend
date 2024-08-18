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

func (entry *Entry) InitalizeFormat() {
	entry.Date = time.Now()
	entry.Archived = false
	entry.RealURL = ensureHttpPrefix(entry.RealURL)
}

func ensureHttpPrefix(url string) string {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "https://" + url
	}
	return url
}
