package platform

import (
	"c361main/clicks"
	"c361main/entries"
	"c361main/entry"
	"c361main/user"
	"net/http"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func New(db *gorm.DB, firebase *firebase.App, rdb *redis.Client, httpClient *http.Client) *gin.Engine {
	router := gin.Default()

	router.Use(CORSMiddleware())

	router.POST("/user", user.PostUser())
	router.POST("/entry", entry.PostEntry(db, rdb))
	router.POST("/merge", user.MergeUser(db, firebase))

	router.PATCH("/entry/:id/archive", entry.ArchivedEntry(db, firebase, rdb))
	router.PATCH("/entry/:id/unarchive", entry.UnarchivedEntry(db, firebase, rdb))
	router.PATCH("/entry/:id", entry.PatchEntryURL(db, firebase, rdb))

	router.PATCH("/entry/:id/addcustom", entry.PatchCustomHandle(db, firebase, rdb, httpClient))
	router.PATCH("/entry/:id/deletecustom", entry.DeleteCustomHandle(db, firebase, rdb, httpClient))

	// router.GET("/entries", entries.GetEntries(firebase, db))

	router.GET("/search", entries.QueryEntries(firebase, db))
	router.GET("/search/:id", entries.QueryEntriesWithSingle(firebase, db))
	router.GET("/entriescsv", entries.GetEntriesCSV(db, firebase, httpClient))

	router.GET("/clicks/:id", clicks.GetClicksByParam(db, firebase, httpClient))
	router.GET("/clickcsv/:id", clicks.GetClickCSV(db, firebase, httpClient))

	router.GET("/haspassword", user.HasPasswordHandler(firebase))

	router.POST("/emailexchange", user.AddExchange(rdb))
	router.GET("/emailexchange/:id", user.GetExchange(rdb))

	router.GET("/customcheck/:id", entry.CheckCustomHandle(db, rdb))

	return router

}
