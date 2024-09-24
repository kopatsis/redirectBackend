package routes

import (
	"c361main/payment/redisfn"
	"c361main/platform"
	"c361main/specialty/sendgridfn"
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sendgrid/sendgrid-go"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/customer"
	"github.com/stripe/stripe-go/v72/paymentmethod"
	"github.com/stripe/stripe-go/v72/sub"
)

type PaymentRequest struct {
	PaymentMethodID string `json:"paymentMethodID"`
}

func PostPayHandler(rdb *redis.Client, auth *auth.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		req := PaymentRequest{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userid, err := platform.GetCookieUserID(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no user in :" + err.Error()})
			return
		}

		firebaseUser, err := auth.GetUser(context.Background(), userid)
		if err != nil || !firebaseUser.EmailVerified || firebaseUser.Email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		email := firebaseUser.Email

		customerParams := &stripe.CustomerParams{
			Email: stripe.String(email),
		}

		userPayment, err := redisfn.GetUserPayment(rdb, userid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var stripeCustomer *stripe.Customer

		oldCust := userPayment != nil && userPayment.CustomerID != ""

		if oldCust {
			stripeCustomer, err = customer.Get(userPayment.CustomerID, nil)
			if err != nil {
				oldCust = false
			}
		}

		if !oldCust {
			stripeCustomer, err = customer.New(customerParams)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}

		attachParams := &stripe.PaymentMethodAttachParams{
			Customer: stripe.String(stripeCustomer.ID),
		}
		_, err = paymentmethod.Attach(req.PaymentMethodID, attachParams)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		customerUpdateParams := &stripe.CustomerParams{
			InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
				DefaultPaymentMethod: stripe.String(req.PaymentMethodID),
			},
		}
		_, err = customer.Update(stripeCustomer.ID, customerUpdateParams)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		priceID := os.Getenv("PRICE_ID")

		subscriptionParams := &stripe.SubscriptionParams{
			Customer: stripe.String(stripeCustomer.ID),
			Items: []*stripe.SubscriptionItemsParams{
				{
					Price: stripe.String(priceID),
				},
			},
		}

		newSub, err := sub.New(subscriptionParams)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := redisfn.CreateSetUserPayment(rdb, userid, stripeCustomer.ID, newSub.ID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		response := gin.H{"success": true}

		pubsub := rdb.Subscribe(context.Background(), "Subscription")
		defer pubsub.Close()

		timeoutDuration := 10 * time.Second

		for {
			select {
			case msg := <-pubsub.Channel():
				parts := strings.Split(msg.Payload, " --- ")
				if len(parts) == 2 {
					subscriptionID, status := parts[0], parts[1]
					if subscriptionID == newSub.ID && status == "Success" {
						c.JSON(http.StatusOK, response)
						return
					}
				}
			case <-time.After(timeoutDuration - time.Since(start)):
				c.JSON(http.StatusOK, gin.H{"error": "Timeout"})
				return
			}
		}
	}
}

func PostCancelHandler(rdb *redis.Client, auth *auth.Client, sendgridClient *sendgrid.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userid, err := platform.GetCookieUserID(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no user in :" + err.Error()})
			return
		}

		userPayment, err := redisfn.GetUserPayment(rdb, userid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		} else if userPayment == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no unarchived (active) subscriptions for user"})
			return
		}

		stripeSub, err := sub.Get(userPayment.SubscriptionID, nil)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		params := &stripe.SubscriptionParams{
			CancelAtPeriodEnd: stripe.Bool(true),
		}

		if _, err := sub.Update(stripeSub.ID, params); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		firebaseUser, err := auth.GetUser(context.Background(), userid)
		if err == nil && firebaseUser.Email != "" {
			sendgridfn.SendCancelEmail(sendgridClient, firebaseUser.Email, true)
		}

		response := gin.H{"success": true}
		c.JSON(http.StatusOK, response)
	}
}

func PostUncancelHandler(rdb *redis.Client, auth *auth.Client, sendgridClient *sendgrid.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userid, err := platform.GetCookieUserID(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no user in :" + err.Error()})
			return
		}

		userPayment, err := redisfn.GetUserPayment(rdb, userid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		} else if userPayment == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no unarchived (active) subscriptions for user"})
			return
		}

		stripeSub, err := sub.Get(userPayment.SubscriptionID, nil)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		params := &stripe.SubscriptionParams{
			CancelAtPeriodEnd: stripe.Bool(false),
		}

		if _, err := sub.Update(stripeSub.ID, params); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		firebaseUser, err := auth.GetUser(context.Background(), userid)
		if err == nil && firebaseUser.Email != "" {
			sendgridfn.SendCancelEmail(sendgridClient, firebaseUser.Email, false)
		}

		response := gin.H{"success": true}
		c.JSON(http.StatusOK, response)
	}
}

func PostUpdatePayment(rdb *redis.Client, auth *auth.Client, sendgridClient *sendgrid.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := PaymentRequest{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userid, err := platform.GetCookieUserID(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no user in :" + err.Error()})
			return
		}

		firebaseUser, err := auth.GetUser(context.Background(), userid)
		if err != nil || !firebaseUser.EmailVerified || firebaseUser.Email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userPayment, err := redisfn.GetUserPayment(rdb, userid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		} else if userPayment == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no unarchived (active) subscriptions for user"})
			return
		}

		s, err := sub.Get(userPayment.SubscriptionID, nil)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		params := &stripe.PaymentMethodAttachParams{
			Customer: stripe.String(s.Customer.ID),
		}
		if _, err = paymentmethod.Attach(req.PaymentMethodID, params); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		customerParams := &stripe.CustomerParams{
			InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
				DefaultPaymentMethod: stripe.String(req.PaymentMethodID),
			},
		}
		if _, err = customer.Update(s.Customer.ID, customerParams); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if _, err = sub.Update(s.ID, &stripe.SubscriptionParams{
			DefaultPaymentMethod: stripe.String(req.PaymentMethodID),
		}); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		sendgridfn.SendPaymentUpdateEmail(sendgridClient, firebaseUser.Email)

		response := gin.H{"success": true}
		c.JSON(http.StatusOK, response)
	}
}
