package platform

import (
	"c361main/entry"
	"c361main/user"

	"cloud.google.com/go/datastore"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func New(db *gorm.DB, client *datastore.Client) *gin.Engine {
	router := gin.Default()

	router.POST("/user", user.PostUser(client))
	router.POST("/entry", entry.PostEntry(db))

	router.PATCH("/entry/:id/archive", entry.ArchivedEntry(db))
	router.PATCH("/entry/:id/unarchive", entry.UnarchivedEntry(db))
	router.PATCH("/entry/:id", entry.PatchEntryURL(db))

	router.Use(CORSMiddleware())

	return router

}
