package aes

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
)

type Aes struct{}

func New() *Aes {
	return &Aes{}
}

var aesKeySize = map[string]int{
	"256": 32,
	"192": 24,
	"128": 16,
}

func generateRandomKey(length int) (string, error) {
	key := make([]byte, length)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	keyHex := hex.EncodeToString(key)
	return keyHex, nil
}

func (aes *Aes) Query(q string) ([]string, error) {
	length, ok := aesKeySize[q]
	if !ok {
		return nil, errors.New("Invalid AES encryption variant. Available variants are 128,192,256")
	}
	key, err := generateRandomKey(length)

	if err != nil {
		return nil, errors.New("Unable to generate AES encryption key")
	}

	r := fmt.Sprintf("%s 1 TXT \"key = %s\"", q, key)

	return []string{r}, nil
}

func (d *Aes) Dump() ([]byte, error) {
	return nil, nil
}
