package yeettest

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func GenerateRSAKey(fname string) error {
	priv, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return fmt.Errorf("failed to generate RSA private key: %V", err)
	}

	pkcs8der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return fmt.Errorf("failed to marshal RSA private key: %V", err)
	}

	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8der,
	}

	out, err := os.Create(fname)
	if err != nil {
		return fmt.Errorf("failed to open RSA private key file: %V", err)
	}

	err = pem.Encode(out, block)
	if err != nil {
		return fmt.Errorf("failed to encode RSA private key as PEM: %V", err)
	}

	return nil
}
