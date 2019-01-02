package service

import (
	"github.com/gin-gonic/gin"
)

var posthandlers = make(map[string]gin.HandlerFunc)
var getHandlers = make(map[string]gin.HandlerFunc)

func Start(group *gin.RouterGroup) {

	for key, value := range posthandlers {

		group.POST(key, value)
	}

	for key, value := range getHandlers {

		group.GET(key, value)
	}
}
