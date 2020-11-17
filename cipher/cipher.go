package cipher

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/sha3"
)

type AESGCM struct {
	cipher.AEAD
	nonce []byte
}

func New(key int) (*AESGCM, error) {
	keyBytes := keyGen(key)
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce, _ := hex.DecodeString("64a9433eae7ccceee2fc0eda")
	return &AESGCM{aesgcm, nonce}, nil
}

func (aesgcm *AESGCM) Encrypt(data []byte) ([]byte, error) {
	return aesgcm.Seal(nil, aesgcm.nonce, data, nil), nil
}

func (aesgcm *AESGCM) Decrypt(data []byte) ([]byte, error) {
	return aesgcm.Open(nil, aesgcm.nonce, data, nil)
}

func keyGen(key int) []byte {
	keyBytes := []byte(fmt.Sprintf("%d", key))
	read := make([]byte, 32)
	shake := sha3.NewShake128()
	shake.Write(keyBytes)
	shake.Read(read)
	return read
}
