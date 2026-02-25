package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io/fs"
	"log"
	"math/big"
	"os"
	"time"
)

func GenerateCert() {
	// create a certificate template
	sn, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	if err != nil {
		log.Fatal(err)
	}
	cert := &x509.Certificate{
		// unique certificate number
		SerialNumber: sn,
		Subject: pkix.Name{
			Organization: []string{"LKS"},
			Country:      []string{"RU"},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatal(err)
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		log.Fatal(err)
	}

	var certPEM bytes.Buffer
	err = pem.Encode(&certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		log.Fatal(err)
	}

	var privateKyePEM bytes.Buffer
	err = pem.Encode(&privateKyePEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		log.Fatal(err)
	}

	if !fileExists("cert.pem") {
		if err = os.WriteFile("cert.pem", certPEM.Bytes(), 0644); err != nil {
			log.Fatal(err)
		}
		if err = os.WriteFile("private.pem", privateKyePEM.Bytes(), 0400); err != nil {
			log.Fatal(err)
		}
		if err = os.WriteFile("cert.der", certBytes, 0644); err != nil {
			log.Fatal(err)
		}
		if err = os.WriteFile("private.der", x509.MarshalPKCS1PrivateKey(privateKey), 0400); err != nil {
			log.Fatal(err)
		}
	}
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		log.Printf("file %s already exists", filename)
		return true
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false
	}
	log.Fatal(err)
	return false
}

func main() {
	GenerateCert()
}
