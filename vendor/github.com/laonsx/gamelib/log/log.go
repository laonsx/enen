package log

import (
	"os"

	"github.com/sirupsen/logrus"
)

func InitLogrus(path string, debug bool) {

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {

		logrus.SetOutput(file)
	} else {

		panic("InitLogrus path: " + path + " err: " + err.Error())
	}

	logrus.SetFormatter(&logrus.TextFormatter{})

	if debug {

		logrus.SetLevel(logrus.DebugLevel)
	} else {

		logrus.SetLevel(logrus.InfoLevel)
	}
}
