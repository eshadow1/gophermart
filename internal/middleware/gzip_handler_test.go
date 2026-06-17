package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipMiddleware(t *testing.T) {
	expectRespEnc := "gzip"

	tests := []struct {
		name            string
		acceptEnc       string
		contentEnc      string
		reqBodyRaw      string
		gzipReqBody     bool
		respContentType string
		respBody        string
		expectStatus    int
		expectGzipResp  bool
		expectReqBody   string
	}{
		{
			name:            "json_gzip_with_compress",
			acceptEnc:       "gzip, deflate, br",
			respContentType: "application/json",
			respBody:        `{"url":"http://localhost:8080/"}`,
			expectStatus:    http.StatusOK,
			expectGzipResp:  true,
			expectReqBody:   "",
		},
		{
			name:            "html_gzip_with_compress",
			acceptEnc:       "gzip",
			respContentType: "text/html",
			respBody:        "<h1>http://localhost:8080/</h1>",
			expectStatus:    http.StatusOK,
			expectGzipResp:  true,
			expectReqBody:   "",
		},
		{
			name:            "plain_gzip_without_compress",
			acceptEnc:       "gzip",
			respContentType: "text/plain",
			respBody:        "http://localhost:8080/",
			expectStatus:    http.StatusOK,
			expectGzipResp:  false,
			expectReqBody:   "",
		},
		{
			name:            "json_without_gzip",
			respContentType: "application/json",
			respBody:        `{"url":"http://localhost:8080/"}`,
			expectStatus:    http.StatusOK,
			expectGzipResp:  false,
			expectReqBody:   "",
		},
		{
			name:            "compress_request",
			contentEnc:      "gzip",
			reqBodyRaw:      `{"url":"https://practicum.yandex.ru/"}`,
			gzipReqBody:     true,
			respContentType: "application/json",
			respBody:        `{"url":"http://localhost:8080/"}`,
			expectStatus:    http.StatusOK,
			expectGzipResp:  false,
			expectReqBody:   `{"url":"https://practicum.yandex.ru/"}`,
		},
		{
			name:            "compress_request_and_compress_response",
			acceptEnc:       "gzip",
			contentEnc:      "gzip",
			reqBodyRaw:      `{"url":"https://practicum.yandex.ru/"}`,
			gzipReqBody:     true,
			respContentType: "application/json",
			respBody:        `{"url":"http://localhost:8080/"}`,
			expectStatus:    http.StatusOK,
			expectGzipResp:  true,
			expectReqBody:   `{"url":"https://practicum.yandex.ru/"}`,
		},
		{
			name:            "compress_request_with_invalid_data",
			contentEnc:      "gzip",
			reqBodyRaw:      "invalid data",
			gzipReqBody:     false,
			respContentType: "application/json",
			respBody:        `{"url":"http://localhost:8080/"}`,
			expectStatus:    http.StatusBadRequest,
			expectGzipResp:  false,
			expectReqBody:   "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var receivedBody []byte

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				receivedBody = body
				w.Header().Set("Content-Type", tc.respContentType)
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(tc.respBody))
				assert.NoError(t, err)
			})

			h := GzipMiddleware()(next)
			var bodyReader io.Reader
			if tc.gzipReqBody {
				var buf bytes.Buffer
				gw := gzip.NewWriter(&buf)
				_, err := gw.Write([]byte(tc.reqBodyRaw))
				require.NoError(t, err)
				gw.Close()
				bodyReader = &buf
			} else {
				bodyReader = bytes.NewBufferString(tc.reqBodyRaw)
			}

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", bodyReader)
			req.Header.Set("Accept-Encoding", tc.acceptEnc)
			req.Header.Set("Content-Encoding", tc.contentEnc)
			req.Header.Set("Content-Type", tc.respContentType)

			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectStatus, rec.Code)
			assert.Equal(t, tc.expectReqBody, string(receivedBody))
			if tc.expectStatus != http.StatusOK {
				return
			}

			respEnc := rec.Header().Get("Content-Encoding")
			respBytes := rec.Body.Bytes()

			if tc.expectGzipResp {
				assert.Equal(t, expectRespEnc, respEnc)
				gr, err := gzip.NewReader(bytes.NewReader(respBytes))
				require.NoError(t, err)
				defer gr.Close()

				respBytes, err = io.ReadAll(gr)
				require.NoError(t, err)
			} else {
				assert.NotEqual(t, expectRespEnc, respEnc)
			}

			assert.Equal(t, tc.respBody, string(respBytes))
		})
	}
}
