package platform

import (
	"c361main/clicks"
	"c361main/entries"
	"c361main/entry"
	"c361main/user"

	"cloud.google.com/go/datastore"
	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func New(db *gorm.DB, client *datastore.Client, firebase *firebase.App) *gin.Engine {
	router := gin.Default()

	router.Use(CORSMiddleware())

	router.POST("/user", user.PostUser(client))
	router.POST("/entry", entry.PostEntry(db))
	router.POST("/merge", user.MergeUser(db, firebase))

	router.PATCH("/entry/:id/archive", entry.ArchivedEntry(db))
	router.PATCH("/entry/:id/unarchive", entry.UnarchivedEntry(db))
	router.PATCH("/entry/:id", entry.PatchEntryURL(db))

	router.GET("/entries", entries.GetEntries(db))
	router.GET("/clicks/:id", clicks.GetClicksByParam(db, firebase))
	router.GET("/clickcsv/:id", clicks.GetClickCSV(db, firebase))

	return router

}
