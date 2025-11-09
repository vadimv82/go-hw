package middleware

import (
	"encoding/json"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type LogEntry struct {
	Timestamp                 string `json:"timestamp"`
	HTTPServerRequestDuration int64  `json:"http.server.request.duration"`
	HTTPLogLevel              string `json:"http.log.level"`
	HTTPRequestMethod         string `json:"http.request.method"`
	HTTPResponseStatusCode    int    `json:"http.response.status_code"`
	HTTPRoute                 string `json:"http.route"`
	HTTPRequestMessage        string `json:"http.request.message"`
	ServerAddress             string `json:"server.address"`
	HTTPRequestHost           string `json:"http.request.host"`
	UserID                    string `json:"user_id,omitempty"`
}

func JSONLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start).Milliseconds()

		statusCode := c.Writer.Status()

		logLevel := "info"
		if statusCode >= 400 && statusCode < 500 {
			logLevel = "warn"
		} else if statusCode >= 500 {
			logLevel = "error"
		}

		userID := ""
		if uid, exists := c.Get("user_id"); exists {
			if uidStr, ok := uid.(string); ok {
				userID = uidStr
			}
		}

		route := c.FullPath()
		if route == "" {
			route = path
		}

		logEntry := LogEntry{
			Timestamp:                 time.Now().Format(time.RFC3339Nano),
			HTTPServerRequestDuration: duration,
			HTTPLogLevel:              logLevel,
			HTTPRequestMethod:         method,
			HTTPResponseStatusCode:    statusCode,
			HTTPRoute:                 route,
			HTTPRequestMessage:        "Incoming request:",
			ServerAddress:             path,
			HTTPRequestHost:           c.Request.Host,
		}

		if userID != "" {
			logEntry.UserID = userID
		}

		logJSON, err := json.Marshal(logEntry)
		if err != nil {
			os.Stderr.WriteString("Incoming request: " + err.Error() + "\n")
			return
		}

		os.Stderr.WriteString(string(logJSON) + "\n")
	}
}
