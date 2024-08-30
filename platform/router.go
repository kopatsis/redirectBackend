package platform

import (
	"c361main/clicks"
	"c361main/entries"
	"c361main/entry"
	"c361main/user"
	"net/http"

	firebase "firebase.google.com/go/v4"
	"github.com/dgraph-io/badger/v3"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func New(db *gorm.DB, firebase *firebase.App, rdb *redis.Client, httpClient *http.Client, bdb *badger.DB) *gin.Engine {
	router := gin.Default()

	router.Use(CORSMiddleware())

	router.POST("/user", user.PostUser())
	router.POST("/entry", entry.PostEntry(db, rdb))
	router.POST("/merge", user.MergeUser(db, firebase))

	router.PATCH("/entry/:id/archive", entry.ArchivedEntry(db, rdb))
	router.PATCH("/entry/:id/unarchive", entry.UnarchivedEntry(db, rdb))
	router.PATCH("/entry/:id", entry.PatchEntryURL(db, rdb))

	router.GET("/entries", entries.GetEntries(firebase, db))
	router.GET("/clicks/:id", clicks.GetClicksByParam(db, firebase, httpClient))
	router.GET("/clickcsv/:id", clicks.GetClickCSV(db, firebase, httpClient))

	router.GET("/haspassword", user.HasPasswordHandler(firebase))

	router.POST("/emailexchange", nil)
	router.GET("/emailexchange", nil)

	return router

}
