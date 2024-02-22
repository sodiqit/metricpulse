package signer

type Signer interface {
	Sign(body string, key string) string
}