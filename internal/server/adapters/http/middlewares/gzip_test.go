package middlewares_test

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/sodiqit/metricpulse.git/internal/server/adapters/http/middlewares"
	"github.com/stretchr/testify/require"
)

func TestGzipMiddleware(t *testing.T) {
	r := chi.NewRouter()

	r.Use(middlewares.Gzip)

	successResult := fmt.Sprintf(`{"test": "%s"}`, strings.Repeat("test", 1000))

	r.Post("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		b, err := io.ReadAll(r.Body)

		if err != nil {
			http.Error(w, "error", http.StatusInternalServerError)
			return
		}

		w.Write(b)
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	requestBody := successResult

	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		r := httptest.NewRequest("POST", ts.URL+"/test", buf)
		r.RequestURI = ""
		r.Header.Add("Content-Encoding", "gzip")
		r.Header.Set("Accept-Encoding", "")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		require.JSONEq(t, successResult, string(b))
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		buf := bytes.NewBufferString(requestBody)

		r := httptest.NewRequest("POST", ts.URL+"/test", buf)
		r.RequestURI = ""
		r.Header.Add("Content-Encoding", "")
		r.Header.Add("Content-Type", "application/json")
		r.Header.Set("Accept-Encoding", "gzip")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		defer resp.Body.Close()

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		require.JSONEq(t, successResult, string(b))
	})
}
