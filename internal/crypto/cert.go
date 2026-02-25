package crypto

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
)

func ReadCert(certPath string) (*x509.Certificate, error) {
	certBytes, err := os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}

	certPemBlock, _ := pem.Decode(certBytes)
	if certPemBlock == nil {
		return nil, err
	}

	cert, err := x509.ParseCertificate(certPemBlock.Bytes)
	if err != nil {
		return nil, err
	}
	return cert, nil
}

func ReadPrivKey(keyPath string) (*rsa.PrivateKey, error) {
	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	keyPemBlock, _ := pem.Decode(keyBytes)
	if keyPemBlock == nil {
		return nil, err
	}

	privKey, err := x509.ParsePKCS1PrivateKey(keyPemBlock.Bytes)
	if err != nil {
		return nil, err
	}
	return privKey, nil
}
