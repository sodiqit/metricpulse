package signer

type Signer interface {
	Sign(data []byte) string
	Verify(data []byte, signature string) bool
}
