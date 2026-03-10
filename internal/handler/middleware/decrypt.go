package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/TheLuckymadman/metawatch/internal/crypto"
)

func DecryptWrapper(privKey *rsa.PrivateKey) Middleware {
	return func(h http.HandlerFunc) http.HandlerFunc {
		f := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			aesKeyHeader := r.Header.Get("AES-Secret")
			if privKey != nil && aesKeyHeader != "" {
				log.Println("content decryption is starting")

				decodedKey, err := base64.StdEncoding.DecodeString(aesKeyHeader)
				if err != nil {
					http.Error(w, "invalid AES-Secret header", http.StatusBadRequest)
					return
				}

				aesKey, err := rsa.DecryptPKCS1v15(rand.Reader, privKey, []byte(decodedKey))
				if err != nil {
					log.Printf("DecryptWrapper RSA decryptiion of the AES-Secret header error: %v", err)
					http.Error(w, "faled to decrypt the request", http.StatusBadRequest)
					return
				}

				encryptedBody, err := io.ReadAll(r.Body)
				if err != nil {
					log.Printf("DecryptWrapper: %v", err)
					http.Error(w, fmt.Sprintf("faled to read body: %v", err), http.StatusBadRequest)
					return
				}
				defer r.Body.Close()

				body, err := crypto.Decrypt(encryptedBody, aesKey)
				if err != nil {
					log.Printf("DecryptWrapper AES decryption body error: %v", err)
					http.Error(w, "faled to decrypt the request", http.StatusBadRequest)
					return
				}
				r.Body = io.NopCloser(bytes.NewReader(body))
				log.Println("content decryption finished")
			}
			h(w, r)
		})
		return f
	}
}
