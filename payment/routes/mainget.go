package routes

import (
	"c361main/payment/redisfn"
	"c361main/platform/middleware"
	stripefunc "c361main/specialty/stripe"
	"context"
	"net/http"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/setupintent"
	"github.com/stripe/stripe-go/v72/sub"
)

func GetHandler(rdb *redis.Client, auth *auth.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := middleware.GetCookieUserID(c)
		if err != nil {
			c.HTML(http.StatusOK, "error.html", gin.H{"Error": "No user id in cookie available: " + err.Error()})
			return
		}

		firebaseUser, err := auth.GetUser(context.Background(), userID)
		if err != nil || !firebaseUser.EmailVerified || firebaseUser.Email == "" {
			if err != nil {
				c.HTML(http.StatusOK, "error.html", gin.H{"Error": err.Error()})
			} else {
				c.HTML(http.StatusOK, "error.html", gin.H{"Error": "Email not verified or no email at all."})
			}
			return
		}

		email := firebaseUser.Email

		userPayment, err := redisfn.GetUserPayment(rdb, userID)
		if err != nil {
			c.HTML(http.StatusOK, "error.html", gin.H{"Error": err.Error(), "Email": email})
			return
		}

		if userPayment == nil {
			params := &stripe.SetupIntentParams{
				PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
			}
			si, err := setupintent.New(params)
			if err != nil {
				c.HTML(http.StatusOK, "error.html", gin.H{"Error": err.Error(), "Email": email})
				return
			}
			c.HTML(http.StatusOK, "pay.html", gin.H{"Secret": si.ClientSecret, "Email": email})
			return
		}

		s, err := sub.Get(userPayment.SubscriptionID, nil)
		if err != nil {
			c.HTML(http.StatusOK, "error.html", gin.H{"Error": err.Error(), "Email": email})
			return
		}

		switch string(s.Status) {
		case "incomplete":
			c.HTML(http.StatusOK, "processing.html", gin.H{"ID": s.ID, "Email": email})
			return
		case "past_due":
			c.HTML(http.StatusOK, "updatepay.html", gin.H{"ID": s.ID, "Email": email})
			return
		}

		if s.CancelAtPeriodEnd {
			c.HTML(http.StatusOK, "ending.html", gin.H{"EndDate": time.Unix(s.CurrentPeriodEnd, 0), "Email": email})
			return
		}

		params := &stripe.SetupIntentParams{
			PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		}
		si, err := setupintent.New(params)
		if err != nil {
			c.HTML(http.StatusOK, "error.html", gin.H{"Error": err.Error(), "Email": email})
			return
		}

		paymentType, cardBrand, lastFour, expMonth, expYear, err := stripefunc.GetPaymentMethodDetails(s.ID)
		if err != nil {
			c.HTML(http.StatusOK, "error.html", gin.H{"Error": err.Error(), "Email": email})
			return
		}

		c.HTML(http.StatusOK, "admin.html", gin.H{
			"PaymentType": paymentType,
			"CardBrand":   cardBrand,
			"LastFour":    lastFour,
			"ExpMonth":    expMonth,
			"ExpYear":     expYear,
			"Secret":      si.ClientSecret,
			"EndDate":     time.Unix(s.CurrentPeriodEnd, 0),
			"Email":       email,
		})
	}
}
