package platform

import (
	"c361main/clicks"
	"c361main/entries"
	"c361main/entry"
	"c361main/payment/routes"
	"c361main/platform/middleware"
	"c361main/user"
	"net/http"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sendgrid/sendgrid-go"
	"gorm.io/gorm"
)

func New(db *gorm.DB, auth *auth.Client, rdb *redis.Client, httpClient *http.Client, sendgridClient *sendgrid.Client) *gin.Engine {
	router := gin.Default()

	router.LoadHTMLGlob("templates/*")

	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.CookieMiddleware(rdb))

	// ROUTES for pay portion
	router.GET("/", routes.GetHandler(rdb, auth))
	router.GET("/loginerror", routes.LoginErrorHandler)

	router.POST("/subscription", routes.PostPayHandler(rdb, auth))
	router.PATCH("/subscription/cancel", routes.PostCancelHandler(rdb, auth, sendgridClient))
	router.PATCH("/subscription/uncancel", routes.PostUncancelHandler(rdb, auth, sendgridClient))
	router.PATCH("/subscription", routes.PostUpdatePayment(rdb, auth, sendgridClient))

	router.POST("/multipass", routes.Multipass(rdb, auth))

	router.POST("/webhook", routes.HandleStripeWebhook(rdb, auth, sendgridClient))
	router.POST("/webhook/equivalent/:id", routes.WebhookEquiv(rdb, auth, sendgridClient))

	router.POST("/administrative/logout", routes.HandleLogAllOut(rdb, auth))
	router.POST("/administrative/delete", routes.HandleDeleteAccount(rdb, auth))

	router.GET("/check/:id", routes.ExternalGetHandler(rdb))
	router.POST("/verifyturn", routes.VerifyTurnstileHandler(httpClient))

	router.POST("/helpemail", routes.InternalEmailHandler(httpClient, sendgridClient))
	router.POST("/administrative/helpemail", routes.ExternalEmailHandler(httpClient, sendgridClient, auth))
	router.POST("/administrative/internalemail", routes.HandleInternalAlertEmail(sendgridClient))

	router.GET("/websocket/:id", routes.WebSocketHandler(rdb))

	router.POST("/logout", routes.HandleUserLogout())

	// REWRITE routes incl on frontend to have combined
	router.POST("/user", user.PostUser())
	router.POST("/entry", entry.PostEntry(db, rdb, auth, httpClient))
	router.POST("/merge", user.MergeUser(db, auth))

	router.PATCH("/entry/:id/archive", entry.ArchivedEntry(db, auth, rdb))
	router.PATCH("/entry/:id/unarchive", entry.UnarchivedEntry(db, auth, rdb))
	router.PATCH("/entry/:id", entry.PatchEntryURL(db, auth, rdb))

	router.PATCH("/entry/:id/addcustom", entry.PatchCustomHandle(db, auth, rdb, httpClient))
	router.PATCH("/entry/:id/deletecustom", entry.DeleteCustomHandle(db, auth, rdb, httpClient))

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
