package signer

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
)

type Sha256Signer struct{}

func (s *Sha256Signer) Sign(body string, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(body))
	return fmt.Sprintf("%x", h.Sum(nil))
}
