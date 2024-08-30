package user

import (
	"net/http"
	"os"

	"github.com/dgraph-io/badger/v3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AddExchange(bdb *badger.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		if c.GetHeader("X-Passcode-ID") != os.Getenv("CHECK_PASSCODE") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		id := uuid.New().String()

		err := bdb.Update(func(txn *badger.Txn) error {
			return txn.Set([]byte(id), []byte(req.Email))
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store data"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"id": id})
	}
}

func GetExchange(bdb *badger.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-Passcode-ID") != os.Getenv("CHECK_PASSCODE") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		id := c.Param("id")

		var email string
		err := bdb.Update(func(txn *badger.Txn) error {
			item, err := txn.Get([]byte(id))
			if err != nil {
				return err
			}
			err = item.Value(func(val []byte) error {
				email = string(val)
				return nil
			})
			if err != nil {
				return err
			}
			return txn.Delete([]byte(id))
		})

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Key not found or failed to retrieve"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"email": email})
	}
}
