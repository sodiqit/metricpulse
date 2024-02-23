package signer

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
)

type Sha256Signer struct{}

func (s *Sha256Signer) Sign(data []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func NewSHA256Signer() *Sha256Signer {
	return &Sha256Signer{}
}
