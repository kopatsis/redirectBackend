package entries

import (
	"c361main/datatypes"
	"log"
	"time"

	"gorm.io/gorm"
)

func DeleteArchivedEntries(db *gorm.DB) {
	var entryIDs []int

	err := db.Model(&datatypes.Entry{}).
		Where("archived = ? AND archived_date < ?", true, time.Now().Add(-8*time.Hour)).
		Pluck("id", &entryIDs).
		Error

	if err != nil {
		log.Printf("Error finding archived entries to delete: %v", err)
		return
	}

	if len(entryIDs) > 0 {
		err = db.Exec("DELETE LOW_PRIORITY FROM datatypes_entries WHERE id IN ?", entryIDs).Error
		if err != nil {
			log.Printf("Error deleting archived entries with LOW_PRIORITY: %v", err)
		}

		err = db.Exec("DELETE LOW_PRIORITY FROM datatypes_clicks WHERE param_key IN ?", entryIDs).Error
		if err != nil {
			log.Printf("Error deleting clicks associated with archived entries with LOW_PRIORITY: %v", err)
		}
	}
}
