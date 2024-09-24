package user

import (
	"context"
	"errors"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
)

func VerifyMultipassRetUID(c *gin.Context, auth *auth.Client) (string, error) {
	idToken := c.PostForm("idToken")
	if idToken == "" {
		return "", errors.New("missing idToken in form data")
	}

	token, err := auth.VerifyIDToken(context.Background(), idToken)
	if err != nil {
		return "", errors.New("failed to verify token")
	}

	return token.UID, nil
}

func GetSubFromJWT(auth *auth.Client, c *gin.Context) (string, error, bool) {

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", nil, true
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	token, err := auth.VerifyIDToken(context.Background(), tokenString)
	if err != nil {
		return "", err, false
	}

	return token.UID, nil, false
}

func GetUserID(auth *auth.Client, c *gin.Context) (string, bool, error) {
	firebaseID, fireFrr, empty := GetSubFromJWT(auth, c)
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

func GetBothIDs(auth *auth.Client, c *gin.Context) (string, string, error) {
	localid, err := getLocalIDTwo(c)
	if err != nil {
		return "", "", err
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", "", err
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	token, err := auth.VerifyIDToken(context.Background(), tokenString)
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
