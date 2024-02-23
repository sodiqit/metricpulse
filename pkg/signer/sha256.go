package signer

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
)

type Sha256Signer struct {
	key string
}

func (s *Sha256Signer) Sign(data []byte) string {
	h := hmac.New(sha256.New, []byte(s.key))
	h.Write(data)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (s *Sha256Signer) Verify(data []byte, expectedSignature string) bool {
	return s.Sign(data) == expectedSignature
}

func NewSHA256Signer(key string) *Sha256Signer {
	return &Sha256Signer{key}
}
