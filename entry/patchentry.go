package entry

import (
	"c361main/convert"
	"c361main/datatypes"
	"net/url"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func PatchEntryURLDB(db *gorm.DB, id int, url string) error {
	return db.Model(&datatypes.Entry{}).Where("id = ?", id).Updates(datatypes.Entry{
		RealURL: url,
	}).Error
}

func PatchEntryURL(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		id10, err := convert.FromSixFour(c.Param("id"))
		if err != nil {
			errorPatch(c, err, "Failed to convert id param", 400)
			return
		}

		var json struct {
			URL string `json:"url" binding:"required"`
		}

		if err := c.ShouldBindJSON(&json); err != nil {
			errorPatch(c, err, "Invalid or missing JSON input", 400)
			return
		}

		json.URL = datatypes.EnsureHttpPrefix(json.URL)
		if _, err := url.Parse(json.URL); err != nil {
			errorPatch(c, err, "Not a URL that can be parsed", 400)
			return
		}

		if err := PatchEntryURLDB(db, id10, json.URL); err != nil {
			errorPatch(c, err, "Failed to patch real url on the database", 400)
			return
		}

		c.Status(204)
	}
}
