package user

import (
	"c361main/datatypes"
	"context"
	"errors"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func OLDhasEmailPasswordAccount(client *auth.Client, uid string) (bool, error) {
	userRecord, err := client.GetUser(context.Background(), uid)
	if err != nil {
		return false, err
	}

	for _, provider := range userRecord.ProviderUserInfo {
		if provider.ProviderID == "password" {
			return true, nil
		}
	}

	return false, nil
}

func isEmailSubbed(uid string, db *gorm.DB) (bool, error) {
	var pref datatypes.UserPreference
	if err := db.Where("uid = ?", uid).First(&pref).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return true, nil
		}
		return false, err
	}
	return pref.AllowsEmails, nil
}

func changeIsEmailSubbed(uid string, subbed bool, db *gorm.DB) error {
	pref := datatypes.UserPreference{UID: uid}
	result := db.Where(&datatypes.UserPreference{UID: uid}).First(&pref)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			pref.HasPassword = false
			pref.AllowsEmails = subbed
			return db.Create(&pref).Error
		}
		return result.Error
	}

	return db.Model(&pref).Updates(datatypes.UserPreference{AllowsEmails: subbed}).Error
}

func hasPasswordAccount(uid string, db *gorm.DB) (bool, error) {
	var pref datatypes.UserPreference
	if err := db.Where("uid = ?", uid).First(&pref).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return pref.HasPassword, nil
}

func addHasPasswordAccount(uid string, db *gorm.DB) error {
	pref := datatypes.UserPreference{UID: uid}
	result := db.Where(&datatypes.UserPreference{UID: uid}).First(&pref)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			pref.HasPassword = true
			pref.AllowsEmails = true
			return db.Create(&pref).Error
		}
		return result.Error
	}

	return db.Model(&pref).Updates(datatypes.UserPreference{HasPassword: true}).Error
}

func HasPasswordPost(auth *auth.Client, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err, empty := GetSubFromJWT(auth, c)
		if empty {
			c.JSON(401, gin.H{
				"Error Type":  "Incorrect auth",
				"Exact Error": errors.New("no token"),
			})
			return
		} else if err != nil {
			c.JSON(401, gin.H{
				"Error Type":  "Incorrect auth",
				"Exact Error": err.Error(),
			})
			return
		}

		if err := addHasPasswordAccount(id, db); err != nil {
			c.JSON(400, gin.H{
				"Error Type":  "Unable to create entry for has password",
				"Exact Error": err.Error(),
			})
			return
		}

		c.JSON(201, gin.H{"uid": id})
	}
}

func HasPasswordHandler(auth *auth.Client, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err, empty := GetSubFromJWT(auth, c)
		if empty {
			c.JSON(401, gin.H{
				"Error Type":  "Incorrect auth",
				"Exact Error": errors.New("no token"),
			})
			return
		} else if err != nil {
			c.JSON(401, gin.H{
				"Error Type":  "Incorrect auth",
				"Exact Error": err.Error(),
			})
			return
		}

		has, err := hasPasswordAccount(id, db)
		if err != nil {
			c.JSON(400, gin.H{
				"Error Type":  "Firebase error",
				"Exact Error": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"HasPassword": has,
		})
	}
}

func IsEmailSubbedPost(auth *auth.Client, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err, empty := GetSubFromJWT(auth, c)
		if empty {
			c.JSON(401, gin.H{
				"Error Type":  "Incorrect auth",
				"Exact Error": errors.New("no token"),
			})
			return
		} else if err != nil {
			c.JSON(401, gin.H{
				"Error Type":  "Incorrect auth",
				"Exact Error": err.Error(),
			})
			return
		}

		if err := changeIsEmailSubbed(id, true, db); err != nil {
			c.JSON(400, gin.H{
				"Error Type":  "Unable to create entry for has password",
				"Exact Error": err.Error(),
			})
			return
		}

		c.JSON(201, gin.H{"uid": id})
	}
}

func IsEmailSubbedDelete(auth *auth.Client, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err, empty := GetSubFromJWT(auth, c)
		if empty {
			c.JSON(401, gin.H{
				"Error Type":  "Incorrect auth",
				"Exact Error": errors.New("no token"),
			})
			return
		} else if err != nil {
			c.JSON(401, gin.H{
				"Error Type":  "Incorrect auth",
				"Exact Error": err.Error(),
			})
			return
		}

		if err := changeIsEmailSubbed(id, false, db); err != nil {
			c.JSON(400, gin.H{
				"Error Type":  "Unable to create entry for has password",
				"Exact Error": err.Error(),
			})
			return
		}

		c.JSON(201, gin.H{"uid": id})
	}
}

func IsEmailSubbedGet(auth *auth.Client, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err, empty := GetSubFromJWT(auth, c)
		if empty {
			c.JSON(401, gin.H{
				"Error Type":  "Incorrect auth",
				"Exact Error": errors.New("no token"),
			})
			return
		} else if err != nil {
			c.JSON(401, gin.H{
				"Error Type":  "Incorrect auth",
				"Exact Error": err.Error(),
			})
			return
		}

		is, err := isEmailSubbed(id, db)
		if err != nil {
			c.JSON(400, gin.H{
				"Error Type":  "Firebase error",
				"Exact Error": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"AllowsEmails": is,
		})
	}
}

func UnsubscribeEmailsViaGet(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.HTML(200, "unsuberr.html", nil)
			return
		}

		if err := changeIsEmailSubbed(id, false, db); err != nil {
			c.HTML(200, "unsuberr.html", nil)
			return
		}

		c.HTML(200, "unsub.html", nil)
	}
}
