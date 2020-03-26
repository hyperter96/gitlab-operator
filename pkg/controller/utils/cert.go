package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"time"
)

var (
	keyBitSize = 4096
)

func getCACertificate() x509.Certificate {
	return x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"GitLab, Inc. CA"},
			Country:      []string{"US"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24 * 366),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
}

func getClientCertificate() x509.Certificate {
	return x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"GitLab, Inc."},
			Country:      []string{"US"},
		},
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24 * 366),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}
}

// PrivateKeyRSA returns a RSA private key
func PrivateKeyRSA() (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, keyBitSize)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// RSAPrivateKeyPEM returns an RSA private key in PEM format
func RSAPrivateKeyPEM() (string, error) {
	key, err := PrivateKeyRSA()
	if err != nil {
		return "", err
	}

	return string(pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)), nil
}

// KeyCertificate returns certificate, privateKey and error
func KeyCertificate() ([]byte, []byte, error) {
	key, err := PrivateKeyRSA()
	if err != nil {
		return nil, nil, err
	}

	privateKeyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)

	cert := getCACertificate()
	ca := getClientCertificate()

	certificateDER, err := x509.CreateCertificate(rand.Reader, &cert, &ca, &key.PublicKey, key)
	if err != nil {
		return nil, nil, err
	}

	return privateKeyPEM,
		pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: certificateDER,
		}), nil
}
