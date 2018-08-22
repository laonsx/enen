package test

import (
	"os"
	"os/signal"
	"syscall"

	"enen/test/robot"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var quit = make(chan os.Signal, 1)

func Run() {

	if viper.GetBool("test.debug") {

		logrus.SetLevel(logrus.DebugLevel)
	} else {

		logrus.SetLevel(logrus.InfoLevel)
	}

	switch viper.GetString("test.func") {

	case "robot":

		robot.Test()

	default:

	}

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
