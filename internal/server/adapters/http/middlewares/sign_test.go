package middlewares_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/sodiqit/metricpulse.git/internal/constants"
	"github.com/sodiqit/metricpulse.git/internal/server/adapters/http/middlewares"
	"github.com/sodiqit/metricpulse.git/internal/server/config"
	"github.com/sodiqit/metricpulse.git/pkg/signer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSuite(config *config.Config) *httptest.Server {
	r := chi.NewRouter()

	r.Use(middlewares.WithSignValidator(config, signer.NewSHA256Signer()))

	r.Post("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		b, err := io.ReadAll(r.Body)

		if err != nil {
			http.Error(w, "error", http.StatusInternalServerError)
			return
		}

		w.Write(b)
	})

	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test"))
	})

	return httptest.NewServer(r)

}

func TestSignValidatorMiddleware(t *testing.T) {
	client := resty.New()
	sha256Signer := signer.NewSHA256Signer()

	tests := []struct {
		name             string
		url              string
		method           string
		body             string
		createSignature  func(signer.Signer, string) string
		needAssignHeader bool
		config           *config.Config
		expectedResult   string
		expectedStatus   int
	}{
		{
			name:   "should return result if provided valid signature",
			method: http.MethodPost,
			body:   `{"test": true}`,
			createSignature: func(signer signer.Signer, key string) string {
				body := `{"test": true}`
				return signer.Sign([]byte(body), key)
			},
			url:              "/test",
			needAssignHeader: true,
			config:           &config.Config{SecretKey: "test"},
			expectedResult:   `{"test": true}`,
			expectedStatus:   http.StatusOK,
		},
		{
			name:   "should return error if provided invalid signature",
			method: http.MethodPost,
			body:   `{"test": false}`,
			createSignature: func(signer signer.Signer, key string) string {
				body := `{"test": true}`
				return signer.Sign([]byte(body), key)
			},
			url:              "/test",
			config:           &config.Config{SecretKey: "test"},
			needAssignHeader: true,
			expectedStatus:   http.StatusBadRequest,
		},
		{
			name:   "should return error if not provide header",
			method: http.MethodPost,
			body:   `{"test": true}`,
			createSignature: func(signer signer.Signer, key string) string {
				body := `{"test": true}`
				return signer.Sign([]byte(body), key)
			},
			url:              "/test",
			config:           &config.Config{SecretKey: "test"},
			needAssignHeader: false,
			expectedStatus:   http.StatusBadRequest,
		},
		{
			name:   "should not validate sign if not POST method",
			method: http.MethodGet,
			createSignature: func(signer signer.Signer, key string) string {
				return ""
			},
			url:              "/test",
			needAssignHeader: false,
			config:           &config.Config{},
			expectedStatus:   http.StatusOK,
			expectedResult:   "test",
		},
		{
			name:   "should not validate sign if not provide key in config",
			method: http.MethodPost,
			body:   `{"test": false}`,
			createSignature: func(signer signer.Signer, key string) string {
				body := `{"test": true}`
				return signer.Sign([]byte(body), key)
			},
			url:              "/test",
			needAssignHeader: true,
			config:           &config.Config{SecretKey: ""},
			expectedStatus:   http.StatusOK,
			expectedResult:   `{"test": false}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := setupSuite(tc.config)

			signature := tc.createSignature(sha256Signer, tc.config.SecretKey)

			defer ts.Close()

			req := client.R()

			req.Method = tc.method
			req.URL = ts.URL + tc.url
			if tc.method == http.MethodPost {
				req.Body = tc.body
			}

			if tc.needAssignHeader {
				req.Header.Add(constants.HashHeader, signature)
			}

			resp, err := req.Send()

			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, resp.StatusCode())

			if tc.expectedStatus == http.StatusOK {
				assert.Equal(t, tc.expectedResult, resp.String())
			}
		})
	}
}
