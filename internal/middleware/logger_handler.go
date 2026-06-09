package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/eshadow1/gophermart/internal/loggers"
)

// LoggerMiddleware создает middleware для логирования всех HTTP-запросов.
func LoggerMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			lrw := newLoggingResponseWriter(w)

			next.ServeHTTP(lrw, r)

			duration := time.Since(start)

			loggers.Log.Infow("http_request",
				"method", r.Method,
				"uri", r.RequestURI,
				"duration", duration.String(),
				"status", lrw.ResponseData.Status,
				"size", fmt.Sprintf("%d bytes", lrw.ResponseData.Size),
			)
		})
	}
}

// responseData хранит метаданные HTTP-ответа: статус код и размер тела ответа.
type (
	responseData struct {
		Status int
		Size   int
	}

	// loggingResponseWriter реализует интерфейс http.ResponseWriter и перехватывает
	// вызовы Write и WriteHeader для сбора статистики об ответе.
	loggingResponseWriter struct {
		http.ResponseWriter
		ResponseData *responseData
	}
)

// newLoggingResponseWriter создает и возвращает новый экземпляр loggingResponseWriter,
// оборачивающий оригинальный http.ResponseWriter для сбора метаданных ответа.
func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{
		ResponseWriter: w,
		ResponseData:   &responseData{},
	}
}

// Write перехватывает запись тела ответа, делегирует её оригинальному ResponseWriter
// и накапливает размер записанных данных в ResponseData.Size.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.ResponseData.Size += size
	return size, err
}

// WriteHeader перехватывает установку статуса ответа, делегирует её оригинальному
// ResponseWriter и сохраняет статус код в ResponseData.Status.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.ResponseData.Status = statusCode
}
