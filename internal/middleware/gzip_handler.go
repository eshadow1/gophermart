package middleware

import (
	"net/http"
	"strings"

	"github.com/eshadow1/gophermart/internal/encoding"
)

// GzipMiddleware создает middleware для прозрачного сжатия HTTP-ответов
// и распаковки HTTP-запросов с использованием алгоритма gzip.
func GzipMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ow := w

			acceptEncoding := r.Header.Get("Accept-Encoding")
			supportsGzip := strings.Contains(acceptEncoding, "gzip")
			if supportsGzip {
				cw := encoding.NewCompressWriter(w)
				ow = cw
				defer cw.Close()
			}

			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				cr, err := encoding.NewCompressReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				r.Body = cr
				r.Header.Del("Content-Encoding")
				defer cr.Close()
			}

			next.ServeHTTP(ow, r)
		})
	}
}
