package crypto

import (
	"github.com/google/uuid"
)

type Crypto struct{}

func NewCrypto() *Crypto {
	return &Crypto{}
}

func (c *Crypto) GenSecretKey() string {
	return uuid.NewString()
}

func (c *Crypto) Encrypt(buffer []byte, password string) ([]byte, error) {
	return []byte{}, nil
}
