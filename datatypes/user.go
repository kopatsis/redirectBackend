package datatypes

import "time"

type User struct {
	Date time.Time
}

type UserPreference struct {
	UID          string `gorm:"primaryKey"`
	HasPassword  bool
	AllowsEmails bool
}
