package entries

import (
	"c361main/datatypes"
	"log"
	"time"

	"gorm.io/gorm"
)

func DeleteArchivedEntries(db *gorm.DB) {
	err := db.Where("archived = ? AND archived_date < ?", true, time.Now().Add(-8*time.Hour)).Delete(&datatypes.Entry{}).Error
	if err != nil {
		log.Printf("Error deleting archived entries: %v", err)
	}
}
