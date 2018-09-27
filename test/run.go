package test

import (
	"os"
	"os/signal"
	"syscall"

	"enen/common"
	"enen/test/robot"
	"github.com/laonsx/gamelib/gofunc"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var quit = make(chan os.Signal, 1)
var conf *common.Config

func Run() {

	serverConfs := make(common.ServiceConf)
	gofunc.LoadJsonConf(gofunc.CONFIGS, "server", &serverConfs)

	var ok bool

	conf, ok = serverConfs["test"]
	if !ok {

		panic("server name(test) not found")
	}

	robot.CenterAddr = conf.CenterAddr

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
