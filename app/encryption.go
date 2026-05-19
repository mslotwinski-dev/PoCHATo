package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"errors"
)

const (
	RSAKeySize = 2048
)

// GenerateKeyPair generates a new RSA public/private key pair
func GenerateKeyPair() (publicKeyStr, privateKeyStr string, err error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, RSAKeySize)
	if err != nil {
		return "", "", err
	}

	// Encode private key
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyStr = base64.StdEncoding.EncodeToString(privateKeyBytes)

	// Encode public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}
	publicKeyStr = base64.StdEncoding.EncodeToString(publicKeyBytes)

	return publicKeyStr, privateKeyStr, nil
}

// DecodePublicKey decodes a base64-encoded public key string
func DecodePublicKey(publicKeyStr string) (*rsa.PublicKey, error) {
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyStr)
	if err != nil {
		return nil, err
	}

	publicKey, err := x509.ParsePKIXPublicKey(publicKeyBytes)
	if err != nil {
		return nil, err
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return rsaPublicKey, nil
}

// DecodePrivateKey decodes a base64-encoded private key string
func DecodePrivateKey(privateKeyStr string) (*rsa.PrivateKey, error) {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyStr)
	if err != nil {
		return nil, err
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// EncryptMessage encrypts a message using RSA public key encryption
func EncryptMessage(publicKeyStr string, message string) (string, error) {
	publicKey, err := DecodePublicKey(publicKeyStr)
	if err != nil {
		return "", err
	}

	encryptedBytes, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		publicKey,
		[]byte(message),
		nil,
	)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encryptedBytes), nil
}

// DecryptMessage decrypts a message using RSA private key decryption
func DecryptMessage(privateKeyStr string, encryptedMessage string) (string, error) {
	privateKey, err := DecodePrivateKey(privateKeyStr)
	if err != nil {
		return "", err
	}

	encryptedBytes, err := base64.StdEncoding.DecodeString(encryptedMessage)
	if err != nil {
		return "", err
	}

	decryptedBytes, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		privateKey,
		encryptedBytes,
		nil,
	)
	if err != nil {
		return "", err
	}

	return string(decryptedBytes), nil
}
