package routes

import (
	"c361main/payment/redisfn"
	"c361main/specialty/sendgridfn"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sendgrid/sendgrid-go"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/charge"
	"github.com/stripe/stripe-go/v72/invoice"
	"github.com/stripe/stripe-go/v72/sub"
	"github.com/stripe/stripe-go/v72/webhook"
)

func HandleStripeWebhook(rdb *redis.Client, auth *auth.Client, sendgridClient *sendgrid.Client) gin.HandlerFunc {
	const MaxBodyBytes = int64(65536)

	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)

		payload, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		sigHeader := c.Request.Header.Get("Stripe-Signature")
		endpointSecret := os.Getenv("END_SECR")

		event, err := webhook.ConstructEvent(payload, sigHeader, endpointSecret)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		switch event.Type {
		case "invoice.payment_succeeded":
			var invoice stripe.Invoice
			if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			subscription, err := sub.Get(invoice.Subscription.ID, nil)
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

		case "invoice.payment_failed":
			var invoice stripe.Invoice
			if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			subscription, err := sub.Get(invoice.Subscription.ID, nil)
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

			if err := redisfn.SetUserPaymentInactive(rdb, userid, subscription.ID); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			if err := rdb.Publish(context.Background(), "Subscriptions", subscription.ID+" --- "+"Fail").Err(); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			if err := sendgridfn.SendFailureEmail(sendgridClient, firebaseUser.Email); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "didn't send email but everything else worked: " + err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"success": true})

		case "charge.dispute.created":
			chargeID, ok := event.Data.Object["charge"].(string)
			if !ok {
				c.JSON(http.StatusBadRequest, gin.H{"error": "unable to extract charge ID from event"})
				return
			}

			chargeObj, err := charge.Get(chargeID, nil)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			invoiceID := chargeObj.Invoice.ID
			if invoiceID == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "no invoice associated with this charge"})
				return
			}

			invoiceObj, err := invoice.Get(invoiceID, nil)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			subscriptionID := invoiceObj.Subscription.ID

			subscription, err := sub.Get(subscriptionID, nil)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			userid, err := redisfn.GetUserBySubID(rdb, subscription.ID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			email := "NO EMAIL FOR THIS ONE"

			firebaseUser, err := auth.GetUser(context.Background(), userid)
			if err == nil && firebaseUser.Email != "" {
				email = firebaseUser.Email
			}

			if err := sendgridfn.SendChargeBackAlert(sendgridClient, subscriptionID, userid, email, string(subscription.Status)); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"success": true})

		default:
			c.JSON(http.StatusOK, gin.H{"status": "event received"})
		}
	}
}
