package middlewares

import (
	"bytes"
	"io"
	"net/http"

	"github.com/sodiqit/metricpulse.git/internal/constants"
	"github.com/sodiqit/metricpulse.git/pkg/signer"
)

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

			next.ServeHTTP(w, r)
		})
	}
}
