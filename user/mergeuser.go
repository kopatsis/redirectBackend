package user

import (
	"c361main/datatypes"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func MergeUser(db *gorm.DB, auth *auth.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		firebaseID, localID, err := GetBothIDs(auth, c)
		if err != nil {
			errorPost(c, err, "couldn't get both IDs")
		}

		err = db.Model(&datatypes.Entry{}).Where("user = ?", localID).Update("user", firebaseID).Error
		if err != nil {
			errorPost(c, err, "couldn't update entries to firebaseID")
			return
		}

		c.JSON(200, gin.H{"status": "success"})

	}
}
