// +build !windows

package router

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	server   *http.Server
	listener net.Listener
)

func contains(s []string, e string) bool {

	for _, a := range s {

		if a == e {

			return true
		}
	}

	return false
}

func reload(listener net.Listener) error {

	tl, ok := listener.(*net.TCPListener)
	if !ok {

		return errors.New("listener is not tcp listener")
	}

	f, err := tl.File()
	if err != nil {

		return err
	}

	args := os.Args[1:]
	if !contains(args, "-g=true") {

		args = append(args, "-g=true")
	}

	cmd := exec.Command(os.Args[0], args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// put socket FD at the first entry
	cmd.ExtraFiles = []*os.File{f}

	return cmd.Start()
}

func signalHandler(server *http.Server, listener net.Listener) {

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)
	for {

		sig := <-ch

		logrus.WithFields(logrus.Fields{
			"pid":    os.Getpid(),
			"signal": sig,
		}).Warn("Signal received.")

		// timeout context for shutdown
		ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
		switch sig {

		case syscall.SIGINT, syscall.SIGTERM:

			// stop
			logrus.WithFields(logrus.Fields{
				"pid":    os.Getpid(),
				"signal": sig,
			}).Warn("SignalHandler stop.")

			signal.Stop(ch)
			server.Shutdown(ctx)

			logrus.WithFields(logrus.Fields{
				"pid":    os.Getpid(),
				"signal": sig,
			}).Warn("SignalHandler graceful shutdown.")

			return

		case syscall.SIGUSR2:

			// reload
			logrus.WithFields(logrus.Fields{
				"pid":    os.Getpid(),
				"signal": sig,
			}).Warn("SignalHandler reload.")

			err := reload(listener)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"pid":    os.Getpid(),
					"signal": sig,
					"error":  err,
				}).Error("SignalHandler graceful failed.")
			}

			server.Shutdown(ctx)

			logrus.WithFields(logrus.Fields{
				"pid":    os.Getpid(),
				"signal": sig,
			}).Warn("SignalHandler graceful reload.")

			return
		}
	}
}

// graceful restart server support.
// reference http://kuangchanglang.com/golang/2017/04/27/golang-graceful-restart
func ListenAndServe(addr string, handler http.Handler) {

	server = &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	var err error
	if viper.GetBool("gmt.graceful") {

		logrus.WithFields(logrus.Fields{
			"pid":    os.Getpid(),
			"listen": addr,
		}).Info("Listening to existing file descriptor 3.")

		// cmd.ExtraFiles: If non-nil, entry i becomes file descriptor 3+i.
		// when we put socket FD at the first entry, it will always be 3(0+3)
		f := os.NewFile(3, "")
		listener, err = net.FileListener(f)
	} else {

		logrus.WithFields(logrus.Fields{
			"pid":    os.Getpid(),
			"listen": addr,
		}).Info("Listening on a new file descriptor.")

		listener, err = net.Listen("tcp4", server.Addr)
	}

	if err != nil {

		logrus.WithFields(logrus.Fields{
			"pid":    os.Getpid(),
			"listen": addr,
			"error":  err,
		}).Error("listener error.")

		return
	}

	go func() {
		// server.Shutdown() stops Serve() immediately, thus server.Serve() should not be in main goroutine
		err = server.Serve(listener)

		if err != nil {

			logrus.WithFields(logrus.Fields{
				"pid":    os.Getpid(),
				"listen": addr,
				"error":  err,
			}).Warn("server.Serve status.")
		}
	}()

	signalHandler(server, listener)
}
