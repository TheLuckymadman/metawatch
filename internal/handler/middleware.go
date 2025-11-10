package handler

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

type (
	Middleware func(http.HandlerFunc) http.HandlerFunc
	responseData struct {
		status int
		size int
	}
	ResponseWriterLogger struct {
		http.ResponseWriter
		responseData responseData
	}
	ResponseWriterCompressor struct {
		http.ResponseWriter
		gzip *gzip.Writer
		needCompress bool
	}
)

func MiddlewareConveyor(h http.HandlerFunc, m ...Middleware) http.HandlerFunc {
	for i := len(m)-1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}

func (r *ResponseWriterCompressor) Write(b []byte) (int, error) {
	if r.needCompress {
		return r.gzip.Write(b)	
	}
	return r.ResponseWriter.Write(b)
}

func (r *ResponseWriterCompressor) WriteHeader(statusCode int) {
	if statusCode == http.StatusOK && (strings.Contains(r.ResponseWriter.Header().Get("Content-Type"), "text/html") || strings.Contains(r.Header().Get("Content-Type"), "application/json")) {
		r.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		r.needCompress = true
	}
	
	r.ResponseWriter.WriteHeader(statusCode)
}

func CompressWrapper(h http.HandlerFunc) http.HandlerFunc {
	f := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			log.Println("content decompressing is starting")
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to create gzip reader: %v", err), http.StatusInternalServerError)
				return
			}
			defer gz.Close()

			r.Body = io.NopCloser(gz)
			r.Header.Del("Content-Encoding")
		}
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			log.Printf("compressing is requested")
			
			gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to create gzip writer: %v", err), http.StatusInternalServerError)
				return
			}
			defer gz.Close()

			rwc := ResponseWriterCompressor{w, gz, false}
			h(&rwc, r)	
			return 
		}
		h(w, r)
	})
	return f
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
