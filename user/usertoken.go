package user

import (
	"context"
	"strings"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
)

func CheckTokenAndPaymentStatus(app *firebase.App, c *gin.Context) (bool, bool) {
	authClient, err := app.Auth(context.Background())
	if err != nil {
		return false, false
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return false, false
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	token, err := authClient.VerifyIDToken(context.Background(), tokenString)
	if err != nil {
		return false, false
	}

	isPaying, ok := token.Claims["isPaying"].(bool)
	if !ok {
		isPaying = false
	}

	return isPaying, true
}
