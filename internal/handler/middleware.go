package handler

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

type (
	Middleware   func(http.HandlerFunc) http.HandlerFunc
	responseData struct {
		status int
		size   int
	}
	ResponseWriterLogger struct {
		http.ResponseWriter
		responseData responseData
	}
	ResponseWriterCompressor struct {
		http.ResponseWriter
		gzip         *gzip.Writer
		needCompress bool
	}
	ResponseWriteHash struct {
		realRespWrite http.ResponseWriter
		status        int
		header        http.Header
		buf           bytes.Buffer
	}
)

func NewResponseWriteHash(w http.ResponseWriter) ResponseWriteHash {
	return ResponseWriteHash{realRespWrite: w, header: make(http.Header), status: http.StatusOK}
}

func MiddlewareConveyor(h http.HandlerFunc, m ...Middleware) http.HandlerFunc {
	for i := len(m) - 1; i >= 0; i-- {
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

func (r *ResponseWriteHash) Header() http.Header {
	return r.header
}

func (r *ResponseWriteHash) Write(b []byte) (int, error) {
	return r.buf.Write(b)
}

func (r *ResponseWriteHash) WriteHeader(statusCode int) {
	r.status = statusCode
}

func (r *ResponseWriteHash) Body() []byte {
	return r.buf.Bytes()
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

func HashWrapper(key string) Middleware {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if key != "" {
				hashStr := r.Header.Get("HashSHA256") 
				if hashStr != "" {
					log.Printf("hash check")
				
					hashData, err := hex.DecodeString(hashStr)
					if err != nil {
						log.Printf("cannot decode hash to string %v", r.RemoteAddr)
						http.Error(w, "cannot decode hash to string", http.StatusBadRequest)
						return
					}
					myHmac := hmac.New(sha256.New, []byte(key))
					body, err := io.ReadAll(r.Body)
					if err != nil {
						log.Printf("read body error in request from %v", r.RemoteAddr)
						http.Error(w, "read body error", http.StatusBadRequest)
						return
					}
					myHmac.Write(body)
					myHash := myHmac.Sum(nil)
					if !hmac.Equal(myHash, hashData) {
						log.Printf("invalid hash in request from %v", r.RemoteAddr)
						http.Error(w, "invalid hash", http.StatusBadRequest)
						return
					} else {
						log.Printf("hash is valid in request from %v", r.RemoteAddr)
					}
					r.Body = io.NopCloser(bytes.NewReader(body))
				}
				rwh := NewResponseWriteHash(w)
				h(&rwh, r)

				log.Printf("calc hash in reply to %v", r.RemoteAddr)
				myHmac := hmac.New(sha256.New, []byte(key))
				myHmac.Write(rwh.buf.Bytes())
				myHash := myHmac.Sum(nil)

				rwh.header.Set("HashSHA256", hex.EncodeToString(myHash))
				for k, v := range rwh.header {
					rwh.realRespWrite.Header()[k] = v
				}
				rwh.realRespWrite.WriteHeader(rwh.status)
				_, err := rwh.realRespWrite.Write(rwh.buf.Bytes())
				if err != nil {
					log.Printf("failed to write response body: %v", err)
				}
				return
			}

			h(w, r)
		}
	}
}
