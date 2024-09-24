package routes

import (
	"c361main/specialty/cloudflare"
	"c361main/specialty/sendgridfn"
	"c361main/user"
	"fmt"
	"net/http"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"github.com/sendgrid/sendgrid-go"
)

type ContactForm struct {
	Email   string `form:"email"`
	Name    string `form:"name"`
	Subject string `form:"subject"`
	Body    string `form:"body"`
	Captcha string `form:"cf-turnstile-response"`
}

func InternalEmailHandler(client *http.Client, sendgridClient *sendgrid.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := actualEmailFunction(client, sendgridClient, c); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
		}
	}
}

func ExternalEmailHandler(client *http.Client, sendgridClient *sendgrid.Client, auth *auth.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err, empty := user.GetSubFromJWT(auth, c)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		} else if empty {
			c.JSON(400, gin.H{"error": "no token"})
			return
		} else if err := actualEmailFunction(client, sendgridClient, c); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
		}
	}
}

func actualEmailFunction(client *http.Client, sendgridClient *sendgrid.Client, c *gin.Context) error {
	var form ContactForm

	if err := c.ShouldBind(&form); err != nil {
		return err
	}

	success, err := cloudflare.VerifyTurnstile(client, form.Captcha)
	if err != nil {
		return err
	} else if !success {
		return fmt.Errorf("Unfortunately, your submission did not pass the Cloudflare verification. Close this window and try again.")
	}

	if err := sendgridfn.SendFormSubmissionEmail(sendgridClient, form.Email, form.Name, form.Subject, form.Body); err != nil {
		return err
	}

	response := map[string]any{"message": "Successfully sent the email! Expect a reply in 1-3 business days, but usually way sooner."}
	c.JSON(200, response)
	return nil
}
