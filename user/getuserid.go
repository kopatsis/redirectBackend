package user

import (
	"errors"
	"strings"

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

func GetUserID(c *gin.Context) (string, error) {
	firebaseID, err, empty := GetSubFromJWT(c)
	localid := c.GetHeader("X-User-ID")
	if firebaseID != "" {
		return firebaseID, nil
	}
	if localid != "" {
		return localid, nil
	}
	if empty {
		return "", errors.New("local id not supplied in X-User-ID header")
	}
	return "", err

}
