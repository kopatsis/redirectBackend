package routes

import (
	"c361main/payment/redisfn"
	"c361main/specialty/sendgridfn"
	"context"
	"net/http"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sendgrid/sendgrid-go"
	"github.com/stripe/stripe-go/v72/sub"
)

func WebhookEquiv(rdb *redis.Client, auth *auth.Client, sendgridClient *sendgrid.Client) gin.HandlerFunc {
	return func(c *gin.Context) {

		id := c.Param("id")

		subscription, err := sub.Get(id, nil)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userid, err := redisfn.GetUserBySubID(rdb, subscription.ID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		firebaseUser, err := auth.GetUser(context.Background(), userid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userPayment, err := redisfn.GetUserPayment(rdb, userid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		} else if userPayment == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no user payment"})
			return
		}

		newPayment := !userPayment.LastDate.IsZero()

		if err := redisfn.SetUserPaymentActive(rdb, userid, subscription.ID, time.Unix(subscription.CurrentPeriodEnd, 0)); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := rdb.Publish(context.Background(), "Subscriptions", subscription.ID+" --- "+"Success").Err(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := sendgridfn.SendSuccessEmail(sendgridClient, firebaseUser.Email, newPayment); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "didn't send email but everything else worked: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}
