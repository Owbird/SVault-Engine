package crypto

type Crypto struct{}

func NewCrypto() *Crypto {
	return &Crypto{}
}

func (c *Crypto) Encrypt(buffer []byte, password string) ([]byte, error) {
	return []byte{}, nil
}
