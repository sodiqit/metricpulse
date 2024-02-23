package middlewares

import (
	"bytes"
	"io"
	"net/http"

	"github.com/sodiqit/metricpulse.git/internal/constants"
	"github.com/sodiqit/metricpulse.git/pkg/signer"
)

type hashResponseWriter struct {
	http.ResponseWriter
	signer signer.Signer
}

func (r *hashResponseWriter) Write(b []byte) (int, error) {
	r.ResponseWriter.Header().Add(constants.HashHeader, r.signer.Sign(b))

	size, err := r.ResponseWriter.Write(b)

	if err != nil {
		return size, err
	}

	return size, err
}

func WithSignValidator(signer signer.Signer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Invalid request", http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(body))

			signature := r.Header.Get(constants.HashHeader)

			if !signer.Verify(body, signature) {
				http.Error(w, "Invalid signature", http.StatusBadRequest)
				return
			}

			hw := &hashResponseWriter{ResponseWriter: w, signer: signer}

			next.ServeHTTP(hw, r)
		})
	}
}
