package middleware

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
)

type ResponseWriterCompressor struct {
	http.ResponseWriter
	gzip         *gzip.Writer
	needCompress bool
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

var gzipPool = sync.Pool{
	New: func() any {
		w, _ := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed)
		return w
	},
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

			// gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
			// if err != nil {
			// 	http.Error(w, fmt.Sprintf("failed to create gzip writer: %v", err), http.StatusInternalServerError)
			// 	return
			// }
			// defer gz.Close()
			gz := gzipPool.Get().(*gzip.Writer)
			gz.Reset(w)
			defer func() {
				gz.Close()
				gzipPool.Put(gz)
			}()

			rwc := ResponseWriterCompressor{w, gz, false}
			h(&rwc, r)
			return
		}
		h(w, r)
	})
	return f
}
