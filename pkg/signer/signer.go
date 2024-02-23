package signer

type Signer interface {
	Sign(data []byte, key string) string
}
