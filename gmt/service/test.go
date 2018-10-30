package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func init() {
	getHandlers["/test"] = account
}

func account(c *gin.Context) {

	c.JSON(http.StatusOK, map[string]string{
		"api":  "api/v1/game/test",
		"text": "测试接口",
	})
}
