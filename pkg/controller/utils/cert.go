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

const (
	// clientCertType represents client certificates
	clientCertType = "client"
	// caCertType represents a CA certificate
	caCertType = "ca"
)

func templateCertificate(certificateType string) x509.Certificate {
	cert := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"GitLab, Inc. CA"},
			Country:      []string{"US"},
		},
		NotBefore:             time.Now().UTC(),
		NotAfter:              time.Now().Add(time.Hour * 24 * 366).UTC(),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	if certificateType == caCertType {
		cert.KeyUsage |= x509.KeyUsageCertSign
		cert.IsCA = true
	}

	if certificateType == clientCertType {
		cert.IPAddresses = []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback}
		cert.IsCA = false
	}

	return cert
}

// PrivateKeyRSA returns a RSA private key
func PrivateKeyRSA(size int) (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, size)
}

// EncodePrivateKeyToPEM returns privateKey in PEM format
func EncodePrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		},
	)
}

// EncodeCertificateToPEM converts a certificate to PEM format
func EncodeCertificateToPEM(cert *x509.Certificate) []byte {
	return pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		},
	)
}

// ClientCertificate returns certificate, privateKey and error
// Takes an RSA private key, CA cert, CA key and slice of hosts
// Returns: certificate, error
func ClientCertificate(key *rsa.PrivateKey, caKey *rsa.PrivateKey, caCert *x509.Certificate, hosts []string) (*x509.Certificate, error) {
	cert := templateCertificate(clientCertType)

	if len(hosts) > 0 {
		for _, hostname := range hosts {
			cert.DNSNames = append(cert.DNSNames, hostname)
		}
	}

	certificateDER, err := x509.CreateCertificate(rand.Reader, &cert, caCert, &key.PublicKey, caKey)
	if err != nil {
		return nil, err
	}

	return x509.ParseCertificate(certificateDER)
}

// CACertificate will be used to sing other certificates
// Returns: certificate, key, error
func CACertificate(key *rsa.PrivateKey) (*x509.Certificate, error) {
	certCA := templateCertificate(clientCertType)
	certDER, err := x509.CreateCertificate(rand.Reader, &certCA, &certCA, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}

	return x509.ParseCertificate(certDER)
}

func decodePEM(data []byte) []byte {
	decoded, _ := pem.Decode(data)
	if decoded == nil {
		return nil
	}

	return decoded.Bytes
}

// ParsePEMPrivateKey takes private key in PEM
//  returns *rsa.PrivateKey
func ParsePEMPrivateKey(key []byte) (*rsa.PrivateKey, error) {
	return x509.ParsePKCS1PrivateKey(decodePEM(key))
}

// ParsePEMCertificate takes certificate in PEM
//  returns *x509.Certificate
func ParsePEMCertificate(certificate []byte) (*x509.Certificate, error) {
	return x509.ParseCertificate(decodePEM(certificate))
}
