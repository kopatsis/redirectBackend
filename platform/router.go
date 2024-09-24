package platform

import (
	"c361main/clicks"
	"c361main/entries"
	"c361main/entry"
	"c361main/user"
	"net/http"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func New(db *gorm.DB, auth *auth.Client, rdb *redis.Client, httpClient *http.Client) *gin.Engine {
	router := gin.Default()

	router.LoadHTMLGlob("templates/*")

	router.Use(CORSMiddleware())
	router.Use(CookieMiddleware(rdb))

	// REWRITE routes incl on frontend to have combined
	router.POST("/user", user.PostUser())
	router.POST("/entry", entry.PostEntry(db, rdb, auth, httpClient))
	router.POST("/merge", user.MergeUser(db, auth))

	router.PATCH("/entry/:id/archive", entry.ArchivedEntry(db, auth, rdb))
	router.PATCH("/entry/:id/unarchive", entry.UnarchivedEntry(db, auth, rdb))
	router.PATCH("/entry/:id", entry.PatchEntryURL(db, auth, rdb))

	router.PATCH("/entry/:id/addcustom", entry.PatchCustomHandle(db, auth, rdb, httpClient))
	router.PATCH("/entry/:id/deletecustom", entry.DeleteCustomHandle(db, auth, rdb, httpClient))

	// router.GET("/entries", entries.GetEntries(firebase, db))

	router.GET("/search", entries.QueryEntries(auth, db, httpClient))
	router.GET("/search/:id", entries.QueryEntriesWithSingle(auth, db, httpClient))
	router.GET("/entriescsv", entries.GetEntriesCSV(db, auth, httpClient))

	router.GET("/clicks/:id", clicks.GetClicksByParam(db, auth, httpClient))
	router.GET("/clickcsv/:id", clicks.GetClickCSV(db, auth, httpClient))

	router.GET("/haspassword", user.HasPasswordHandler(auth, rdb))
	router.POST("/haspassword", user.HasPasswordPost(auth, rdb))

	router.POST("/emailexchange", user.AddExchange(rdb))
	router.GET("/emailexchange/:id", user.GetExchange(rdb))

	router.GET("/customcheck/:id", entry.CheckCustomHandle(db, auth, rdb))

	return router

}
