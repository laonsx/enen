//+build windows

package router

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	server   *http.Server
	listener net.Listener
)

func ListenAndServe(addr string, handler *gin.Engine) {

	err := handler.Run(addr)
	if err != nil {

		logrus.WithFields(logrus.Fields{
			"listen": addr,
			"error":  err,
		}).Error("Start server failed.")
	}
}
