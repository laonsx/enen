package router

import (
	"bytes"
	"io/ioutil"
	"os"
	"time"

	"enen/gmt/service"
	"github.com/gin-gonic/gin"
	"github.com/newrelic/go-agent"
	"github.com/sirupsen/logrus"
)

func Run(httpAddr string) {

	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	newrelicApp := initNewRelic()
	if newrelicApp != nil {

		router.Use(newRelic(newrelicApp))
	}

	// Add a ginrus middleware, which:
	//   - Logs all requests, like a combined access and error log.
	//   - Logs to stdout.
	//   - RFC3339 with UTC time format.
	router.Use(ginrus(logrus.StandardLogger(), time.RFC3339, true))

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	router.Use(gin.Recovery())

	var basePath = "/api/v1"

	v1 := router.Group(basePath)
	{
		game := v1.Group("/game")
		service.Start(game)
	}

	// support cors https://github.com/gin-gonic/gin/issues/29
	router.Use(func(c *gin.Context) {
		// Run this on all requests
		// Should be moved to a proper middleware
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, DELETE, PUT, PATCH, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "API-Key, accept, Content-Type, Token")
		c.Writer.Header().Set("Access-Control-Max-Age", "0")
		c.Writer.Header().Set("Content-Length", "0")
		c.Next()
	})

	router.OPTIONS("/*cors", func(c *gin.Context) {
		// Empty 200 response
	})

	logrus.WithFields(logrus.Fields{
		"pid":    os.Getpid(),
		"listen": httpAddr,
		"path":   basePath,
	}).Info("Start server.")

	ListenAndServe(httpAddr, router)
}

var (
	newRelicEnable  = false
	newRelicAppName = ""
	newRelicLicense = ""
)

func initNewRelic() newrelic.Application {

	if newRelicEnable || newRelicAppName == "" || newRelicLicense == "" {

		return nil
	}

	logrus.WithFields(logrus.Fields{
		"app_name": newRelicAppName,
		"license":  newRelicLicense,
	}).Info("New Relic start.")

	config := newrelic.NewConfig(newRelicAppName, newRelicLicense)
	app, err := newrelic.NewApplication(config)

	if err != nil {

		logrus.WithFields(logrus.Fields{
			"app_name": newRelicAppName,
			"license":  newRelicLicense,
			"error":    err,
		}).Error("New Relic NewApplication failed.")

		return nil
	}

	return app
}

func newRelic(app newrelic.Application) gin.HandlerFunc {
	return func(c *gin.Context) {

		txn := app.StartTransaction(c.Request.RequestURI, c.Writer, c.Request)
		defer txn.End()

		c.Next()
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

type loggerEntryWithFields interface {
	WithFields(fields logrus.Fields) *logrus.Entry
}

// Ginrus returns a gin.HandlerFunc (middleware) that logs requests using logrus.
//
// Requests with errors are logged using logrus.Error().
// Requests without errors are logged using logrus.Info().
//
// It receives:
//   1. A time package format string (e.g. time.RFC3339).
//   2. A boolean stating whether to use UTC time zone or local.
func ginrus(logger loggerEntryWithFields, timeFormat string, utc bool) gin.HandlerFunc {
	return func(c *gin.Context) {

		if c.Request.Method != "POST" && c.Request.Method != "GET" {
			c.Next()
			return
		}

		// some evil middleware modify this values
		path := c.Request.URL.Path

		// Read the Body content
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = ioutil.ReadAll(c.Request.Body)

			// Restore the io.ReadCloser to its original state
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		start := time.Now()

		c.Next()

		end := time.Now()
		latency := end.Sub(start)
		if utc {
			end = end.UTC()
		}

		fields := logrus.Fields{
			"status":     c.Writer.Status(),
			"method":     c.Request.Method,
			"path":       path,
			"ip":         c.ClientIP(),
			"latency":    latency,
			"user-agent": c.Request.UserAgent(),
			"time":       end.Format(timeFormat),
			"response":   blw.body.String(),
		}

		if bodyBytes != nil {
			fields["request"] = string(bodyBytes)
		}

		entry := logger.WithFields(fields)

		if len(c.Errors) > 0 {
			// Append error field if this is an erroneous request.
			entry.Error(c.Errors.String())
		} else {
			entry.Info("Gin Processed")
		}
	}
}
