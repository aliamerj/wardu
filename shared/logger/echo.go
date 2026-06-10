package logger

import (
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/rs/zerolog"
)

// RequestLogger returns Echo middleware that emits one structured log line per request.
func RequestLogger(log zerolog.Logger) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		HandleError:      true,
		LogLatency:       true,
		LogMethod:        true,
		LogStatus:        true,
		LogURI:           true,
		LogURIPath:       true,
		LogRoutePath:     true,
		LogRemoteIP:      true,
		LogRequestID:     true,
		LogUserAgent:     true,
		LogReferer:       true,
		LogContentLength: true,
		LogResponseSize:  true,
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			event := log.Info()
			if v.Error != nil {
				event = log.Error().Err(v.Error)
			}

			event.
				Str("method", v.Method).
				Str("uri", v.URI).
				Str("path", v.URIPath).
				Str("route", v.RoutePath).
				Str("remote_ip", v.RemoteIP).
				Str("request_id", v.RequestID).
				Str("user_agent", v.UserAgent).
				Str("referer", v.Referer).
				Int("status", v.Status).
				Str("content_length", v.ContentLength).
				Int64("response_size", v.ResponseSize).
				Dur("latency", v.Latency).
				Msg("http request")

			return nil
		},
	})
}
