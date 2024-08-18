package platform

import (
	"cloud.google.com/go/datastore"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func New(db *gorm.DB, client *datastore.Client) *gin.Engine {
	router := gin.Default()

	router.Use(CORSMiddleware())

	return router

}
