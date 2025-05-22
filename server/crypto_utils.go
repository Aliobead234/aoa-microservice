package main

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

var privKey *rsa.PrivateKey

func SetPrivateKey(pk *rsa.PrivateKey) {
	privKey = pk
}

func Sign(data []byte) (string, error) {
	return SignRSA(privKey, data)
}

func HashSHA256(data string) string {
	sum := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", sum[:])
}

func EncryptAES256(key []byte, plaintext []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	ct := make([]byte, aes.BlockSize+len(plaintext))
	copy(ct, iv)
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ct[aes.BlockSize:], plaintext)
	return base64.StdEncoding.EncodeToString(ct), nil
}

func DecryptAES256(key []byte, b64cipher string) ([]byte, error) {
	ct, err := base64.StdEncoding.DecodeString(b64cipher)
	if err != nil {
		return nil, err
	}
	if len(ct) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := ct[:aes.BlockSize]
	ctData := ct[aes.BlockSize:]
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	plaintext := make([]byte, len(ctData))
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(plaintext, ctData)
	return plaintext, nil
}

func SignRSA(priv *rsa.PrivateKey, data []byte) (string, error) {
	// Compute SHA-256 hash of data
	hashed := sha256.Sum256(data)
	// Sign using the crypto.SHA256 constant
	sig, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

func VerifyRSA(pub *rsa.PublicKey, data []byte, sigB64 string) error {
	sig, err := base64.StdEncoding.DecodeString(sigB64)
	if err != nil {
		return err
	}
	hashed := sha256.Sum256(data)
	return rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashed[:], sig)
}
