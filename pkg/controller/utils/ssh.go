package utils

import (
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"log"

	"golang.org/x/crypto/ssh"
)

const (
	bitSize = 4096
)

// Keypair takes a []byte private key, []byte public key
// and an error and returns the private key and publickey
// in string format if error is nil
func Keypair(privateKey, publicKey []byte, err error) (string, string) {
	if err != nil {
		log.Fatal(err)
	}
	return string(privateKey), string(publicKey)
}

// RSAKeyPair returns an SSH RSA public and private
// key of type byte slice and an error
func RSAKeyPair() ([]byte, []byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, nil, err
	}

	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	// Generate SSH public key
	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	publicKeyBytes := ssh.MarshalAuthorizedKey(publicKey)

	return pem.EncodeToMemory(block), publicKeyBytes, nil
}

// DSAKeyPair returns an SSH key pair of type DSA
func DSAKeyPair() ([]byte, []byte, error) {
	params := new(dsa.Parameters)

	if err := dsa.GenerateParameters(params, rand.Reader, dsa.L1024N160); err != nil {
		return nil, nil, err
	}

	privateKey := new(dsa.PrivateKey)
	privateKey.PublicKey.Parameters = *params

	if err := dsa.GenerateKey(privateKey, rand.Reader); err != nil {
		return nil, nil, err
	}

	asnBytes, err := asn1.Marshal(*privateKey)
	if err != nil {
		return nil, nil, err
	}

	block := &pem.Block{
		Type:  "OPENSSH PRIVATE KEY",
		Bytes: asnBytes,
	}

	// publicKey := privateKey.PublicKey
	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	return pem.EncodeToMemory(block), ssh.MarshalAuthorizedKey(publicKey), nil
}

// ECDSAKeyPair returns ECDSA SSH keypair and error
func ECDSAKeyPair() ([]byte, []byte, error) {
	privatekey, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	keyBytes, err := x509.MarshalECPrivateKey(privatekey)
	if err != nil {
		return nil, nil, err
	}

	block := &pem.Block{
		Type:  "ECDSA PRIVATE KEY",
		Bytes: keyBytes,
	}

	publicKey, err := ssh.NewPublicKey(&privatekey.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	return pem.EncodeToMemory(block), ssh.MarshalAuthorizedKey(publicKey), nil
}

// ED25519KeyPair return ED25519 SSH key pair and error
func ED25519KeyPair() ([]byte, []byte, error) {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	asnBytes, err := asn1.Marshal(private)
	if err != nil {
		return nil, nil, err
	}

	privatePEM := pem.Block{
		Type:  "OPENSSH PRIVATE KEY",
		Bytes: asnBytes,
	}

	publicKey, err := ssh.NewPublicKey(public)
	if err != nil {
		return nil, nil, err
	}

	return pem.EncodeToMemory(&privatePEM), ssh.MarshalAuthorizedKey(publicKey), nil
}
