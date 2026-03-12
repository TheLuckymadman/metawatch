package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
)

type ResponseWriteHash struct {
	realRespWrite http.ResponseWriter
	status        int
	header        http.Header
	buf           bytes.Buffer
}

func NewResponseWriteHash(w http.ResponseWriter) ResponseWriteHash {
	return ResponseWriteHash{realRespWrite: w, header: make(http.Header), status: http.StatusOK}
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
