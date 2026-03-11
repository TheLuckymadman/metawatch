package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/TheLuckymadman/metawatch/internal/crypto"
	"github.com/TheLuckymadman/metawatch/internal/model"
)

type jsonSender struct {
	client    *http.Client
	address   string
	compress  bool
	key       string
	cryptoKey string
}

func NewJSONSender(client *http.Client, address string, compress bool, key string, cryptoKey string) *jsonSender {
	return &jsonSender{client, address, compress, key, cryptoKey}
}

func (j *jsonSender) SendMetric(ctx context.Context, metric model.Metrics, localIP string) error {
	return j.sendData(ctx, []model.Metrics{metric}, localIP)
}

func (j *jsonSender) SendMetrics(ctx context.Context, metrics []model.Metrics, localIP string) error {
	return j.sendData(ctx, metrics, localIP)
}

func (j *jsonSender) sendData(ctx context.Context, metrics []model.Metrics, localIP string) error {
	var url = fmt.Sprintf("%s/update/", j.address)
	var aesSecret string
	metricsJSON, err := json.Marshal(&metrics)
	if err != nil {
		log.Printf("marshalling metrics data failed with error: %v", err)
		return fmt.Errorf("marshalling metrics data failed with error: %w", err)
	}

	var body bytes.Buffer
	if j.compress {
		gzipBody := gzip.NewWriter(&body)
		_, err = gzipBody.Write(metricsJSON)

		if err != nil {
			return fmt.Errorf("gzip write failed: %w", err)
		}

		if err = gzipBody.Close(); err != nil {
			return fmt.Errorf("gzip close failed: %w", err)
		}
	} else {
		body.Write(metricsJSON)
	}

	if j.cryptoKey != "" {
		certificate, err := crypto.ReadCert(j.cryptoKey)
		if err != nil {
			return fmt.Errorf("certificate reading failure: %w", err)
		}
		aesKey, err := crypto.GenerateAESKey()
		if err != nil {
			return fmt.Errorf("generate AES key error: %w", err)
		}
		encryptedAESKey, err := rsa.EncryptPKCS1v15(rand.Reader, certificate.PublicKey.(*rsa.PublicKey), aesKey)
		if err != nil {
			return fmt.Errorf("RSA encrypt AES key error: %w", err)
		}
		encryptedBody, err := crypto.Encrypt(body.Bytes(), aesKey)
		if err != nil {
			return fmt.Errorf("create new AES cipher error: %w", err)
		}
		body.Reset()
		body.Write(encryptedBody)
		aesSecret = base64.StdEncoding.EncodeToString(encryptedAESKey)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &body)
	if err != nil {
		log.Printf("creating request failed: %v", err)
		return fmt.Errorf("creating request failed: %w", err)
	}

	if aesSecret != "" {
		request.Header.Set("AES-Secret", aesSecret)
	}
	if j.compress {
		request.Header.Set("Content-Encoding", "gzip")
	}
	request.Header.Set("Content-Type", "application/json")
	if j.key != "" {
		request.Header.Set("HashSHA256", generateHMAC(j.key, metricsJSON))

	}
	if localIP != "" {
		request.Header.Set("X-Real-IP", localIP)
	}

	response, err := j.client.Do(request)
	if err != nil {
		log.Printf("sending the metrics batch failed with: %v", err)
		return fmt.Errorf("sending the metrics batch failed with: %w", err)
	}
	defer response.Body.Close()
	resBody, _ := io.ReadAll(response.Body)
	log.Printf("sending metrics batch on %v with status: %v, body: %v", url, response.Status, string(resBody))

	return nil
}

func generateHMAC(key string, body []byte) string {
	h := hmac.New(sha256.New, []byte(key))
	_, err := h.Write(body)
	if err != nil {
		log.Printf("HMAC generation failed: %v", err)
	}
	result := h.Sum(nil)
	return hex.EncodeToString(result)
}

func (j *jsonSender) Close() error {
	return nil
}
