package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type ResponseWriterLogger struct {
	http.ResponseWriter
	responseData responseData
}

func (r *ResponseWriterLogger) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size = size
	return size, err
}

func (r *ResponseWriterLogger) WriteHeader(statusCode int) {
	r.responseData.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func LoggerWrapper(sugar zap.SugaredLogger) func(h http.HandlerFunc) http.HandlerFunc {
	f := func(h http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			uri := r.RequestURI
			method := r.Method

			rwl := ResponseWriterLogger{
				w, responseData{},
			}

			h.ServeHTTP(&rwl, r)

			duration := time.Since(start)

			sugar.Infoln(
				"uri", uri,
				"method", method,
				"duration", duration,
				"r_status", rwl.responseData.status,
				"r_size", rwl.responseData.size,
			)

		})
	}

	return f
}
