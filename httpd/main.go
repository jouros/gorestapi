package main

import (
	"restapi/httpd/handler" // handlers to handle different requests
	"restapi/platform/newsfeed"
	"restapi/platform/initialize" // db setup
	"restapi/platform/data" // sql

	"github.com/gin-gonic/gin"
)

func main() {
	feed := newsfeed.New() // feed == struct Items []item
	r := gin.Default()     // create a router to handle http traffic to handlers

	r.GET("/ping", handler.PingGet()) // call PingGet func from package handler
	r.GET("/newsfeed", handler.NewsFeedGet(feed))
	r.POST("/newsfeed", handler.NewsFeedPost(feed))

	r.Run() // listen and serve on 0.0.0.0:8080

}
