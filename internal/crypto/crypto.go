package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

type Crypto struct{}

func NewCrypto() *Crypto {
	return &Crypto{}
}

func (c *Crypto) GenSecretKey() []byte {
	KEY_SIZE := 32

	key := make([]byte, KEY_SIZE)

	_, err := rand.Read(key)
	if err != nil {
		return nil
	}
	return key
}

func (c *Crypto) Encrypt(buffer, key []byte) ([]byte, error) {
	ciph, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}

	gcm, err := cipher.NewGCM(ciph)
	if err != nil {
		return []byte{}, err
	}

	nonce := make([]byte, gcm.NonceSize())

	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return []byte{}, err
	}

	return gcm.Seal(nonce, nonce, buffer, nil), nil
}

func (c *Crypto) Decrypt(encrypted, key []byte) ([]byte, error) {
	ciph, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(ciph)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(encrypted) < nonceSize {
		return nil, fmt.Errorf("invalid encrypted data size")
	}

	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]

	originalData, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return originalData, nil
}
