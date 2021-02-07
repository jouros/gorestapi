package handler

import (
	"net/http"
	"restapi/platform/newsfeed"

	"github.com/gin-gonic/gin"
)

// takes GetAll for struct feed from interface, returns http.statusOK with all values in JSON
func NewsFeedGet(feed newsfeed.Getter) gin.HandlerFunc {
	return func(c *gin.Context) {
		results := feed.GetAll()
		c.JSON(http.StatusOK, results)

	}
}
