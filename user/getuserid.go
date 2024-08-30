package user

import (
	"context"
	"errors"
	"strings"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
)

func GetSubFromJWT(app *firebase.App, c *gin.Context) (string, error, bool) {
	authClient, err := app.Auth(context.Background())
	if err != nil {
		return "", nil, true
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", nil, true
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	token, err := authClient.VerifyIDToken(context.Background(), tokenString)
	if err != nil {
		return "", err, false
	}

	return token.UID, nil, false
}

func GetUserID(app *firebase.App, c *gin.Context) (string, bool, error) {
	firebaseID, fireFrr, empty := GetSubFromJWT(app, c)
	localid, localErr := getLocalIDTwo(c)

	if firebaseID != "" && fireFrr == nil {
		return firebaseID, true, nil
	}

	if localErr == nil {
		return localid, false, nil
	}

	if empty {
		return "", false, localErr
	}

	return "", false, fireFrr

}

func GetBothIDs(app *firebase.App, c *gin.Context) (string, string, error) {
	localid, err := getLocalIDTwo(c)
	if err != nil {
		return "", "", err
	}

	authClient, err := app.Auth(context.Background())
	if err != nil {
		return "", "", err
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", "", err
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	token, err := authClient.VerifyIDToken(context.Background(), tokenString)
	if err != nil {
		return "", "", err
	}

	return token.UID, localid, nil
}

func getLocalIDTwo(c *gin.Context) (string, error) {
	localid, err := c.Cookie("useruuid")
	if localid == "" || err != nil {
		headerid := c.GetHeader("X-User-ID")
		if headerid == "" {
			return "", errors.New("no user id in either cookie or header")
		}
		return headerid, nil
	}
	return localid, nil
}
