package entry

import (
	"c361main/convert"
	"c361main/datatypes"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func errorPatch(c *gin.Context, err error, reason string, errorCode int) {
	fmt.Printf("Failed to patch: %v", err)
	c.JSON(errorCode, gin.H{
		"Error Type":  reason,
		"Exact Error": err.Error(),
	})
}

func UnarchivedEntryDB(db *gorm.DB, id int) error {
	return db.Model(&datatypes.Entry{}).Where("id = ?", id).Updates(datatypes.Entry{
		Archived: false,
	}).Error
}

func UnarchivedEntry(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		id10, err := convert.FromSixFour(c.Param("id"))
		if err != nil {
			errorPatch(c, err, "Failed to convert id param", 400)
			return
		}

		if err := UnarchivedEntryDB(db, id10); err != nil {
			errorPatch(c, err, "Failed to archive on the database", 400)
			return
		}

		c.Status(204)
	}
}
