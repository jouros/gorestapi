package handler

import (
	"net/http"
	"restapi/platform/newsfeed"

	"github.com/gin-gonic/gin"
)

// JSON struct formatting
type newsfeedPostRequest struct {
	Title string `json:"title"`
	Post  string `json:"post"`
}

// takes method Add() from interface Adder for struct feed,
// return: 1) add post data to struct Item and return http 204 nocontent to REST API call
func NewsFeedPost(feed newsfeed.Adder) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestBody := newsfeedPostRequest{} // newsfeedPostRequest struct
		c.Bind(&requestBody)                 // query data from address of struct

		item := newsfeed.Item{ // fill in data struct Item
			Title: requestBody.Title,
			Post:  requestBody.Post,
		}

		feed.Add(item) // call Add() interface

		c.Status(http.StatusNoContent) // return http 204 nocontent to REST API call
	}
}
