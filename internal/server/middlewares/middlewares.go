package middlewares

import (
	"net/http"
	"time"

	"github.com/sodiqit/metricpulse.git/internal/logger"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func WithLogger(logger logger.ILogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			responseData := &responseData{
				status: http.StatusOK,
				size:   0,
			}
			lw := &loggingResponseWriter{
				ResponseWriter: w,
				responseData:   responseData,
			}

			next.ServeHTTP(lw, r)

			duration := time.Since(start)

			logger.Infow(
				"New request",
				"uri", r.RequestURI,
				"method", r.Method,
				"status", lw.responseData.status,
				"duration", duration.String(),
				"size", lw.responseData.size,
			)
		})
	}
}
