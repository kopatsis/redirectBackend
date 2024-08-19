package user

import (
	"context"
	"errors"
	"strings"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func GetSubFromJWT(c *gin.Context) (string, error, bool) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", jwt.ErrInvalidType, true
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", jwt.ErrInvalidType, true
	}

	tokenString := parts[1]

	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return "", err, false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", jwt.ErrInvalidKey, false
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return "", jwt.ErrInvalidKey, false
	}

	return sub, nil, false
}

func GetUserID(c *gin.Context) (string, bool, error) {
	firebaseID, err, empty := GetSubFromJWT(c)
	localid := c.GetHeader("X-User-ID")

	if firebaseID != "" {
		return firebaseID, true, nil
	}
	if localid != "" {
		return localid, false, nil
	}
	if empty {
		return "", false, errors.New("local id not supplied in X-User-ID header")
	}
	return "", false, err

}

func GetBothIDs(app *firebase.App, c *gin.Context) (string, string, error) {
	localid := c.GetHeader("X-User-ID")
	if localid == "" {
		return "", "", errors.New("no local id")
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
