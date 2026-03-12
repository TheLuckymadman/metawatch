package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func testHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("test handler response"))
}

func TestCompressWrapper_Response(t *testing.T) {
	router := chi.NewRouter()
	router.Post("/", CompressWrapper(testHandler))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	router.ServeHTTP(rec, req)

	respHeader := rec.Header().Get("Content-Encoding")
	assert.Equal(t, "gzip", respHeader)
}

func TestHashWrapper_HashValid(t *testing.T) {
	key := []byte("123")
	messageToSigned := []byte("singed message")

	hmac := hmac.New(sha256.New, key)
	hmac.Write(messageToSigned)
	hash := hmac.Sum(nil)
	hashStr := hex.EncodeToString(hash)

	router := chi.NewRouter()
	router.Post("/", HashWrapper(string(key))(testHandler))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(messageToSigned))
	req.Header.Set("HashSHA256", hashStr)

	router.ServeHTTP(rec, req)
	res := rec.Result()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)
	rec.Result().Body.Close()
}
