package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func LoginErrorHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "loginerror.html", nil)
}
