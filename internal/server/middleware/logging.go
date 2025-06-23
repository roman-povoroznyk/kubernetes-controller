package middleware

import (
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
)

// RequestLogger is a middleware that logs HTTP requests
func RequestLogger(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		// Generate unique request ID
		requestID := uuid.New().String()
		ctx.Response.Header.Set("X-Request-ID", requestID)

		// Capture start time
		start := time.Now()

		// Extract request details
		method := string(ctx.Method())
		path := string(ctx.Path())
		ip := ctx.RemoteIP().String()
		userAgent := string(ctx.UserAgent())

		// Log incoming request
		log.Info().
			Str("request_id", requestID).
			Str("method", method).
			Str("path", path).
			Str("ip", ip).
			Str("user_agent", userAgent).
			Int("content_length", len(ctx.Request.Body())).
			Msg("Request received")

		// Process the request
		next(ctx)

		// Calculate duration
		duration := time.Since(start)

		// Select appropriate log level based on status code
		statusCode := ctx.Response.StatusCode()
		logger := log.Info()

		if statusCode >= 400 && statusCode < 500 {
			logger = log.Warn()
		} else if statusCode >= 500 {
			logger = log.Error()
		}

		// Log the response
		logger.
			Str("request_id", requestID).
			Int("status", statusCode).
			Dur("duration", duration).
			Str("method", method).
			Str("path", path).
			Int("response_size", len(ctx.Response.Body())).
			Msg("Request completed")
	}
}
