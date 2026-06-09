package encoding

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCompressWriter(t *testing.T) {
	tests := []struct {
		name                    string
		body                    string
		contentType             string
		expectedStatus          int
		expectedContentEncoding string
		expectedBody            string
	}{
		{
			name:                    "without compression",
			body:                    "test",
			contentType:             "app",
			expectedStatus:          http.StatusOK,
			expectedContentEncoding: "",
			expectedBody:            "test",
		},
		{
			name:                    "with compression json",
			body:                    "test",
			contentType:             "application/json",
			expectedStatus:          http.StatusOK,
			expectedContentEncoding: "gzip",
			expectedBody:            "\x1f\x8b\b\x00\x00\x00\x00\x00\x00\xff",
		},
		{
			name:                    "with compression html",
			body:                    "test",
			contentType:             "text/html",
			expectedStatus:          http.StatusOK,
			expectedContentEncoding: "gzip",
			expectedBody:            "\x1f\x8b\b\x00\x00\x00\x00\x00\x00\xff",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			w.Header().Set("Content-Type", tc.contentType)

			cw := NewCompressWriter(w)
			defer cw.Close()

			cw.WriteHeader(http.StatusOK)

			_, errWrite := cw.Write([]byte(tc.body))
			require.NoError(t, errWrite)

			body, err := io.ReadAll(w.Body)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.Equal(t, tc.expectedContentEncoding, cw.Header().Get("Content-Encoding"))
			assert.Equal(t, tc.expectedContentEncoding, w.Header().Get("Content-Encoding"))
			assert.Equal(t, tc.expectedBody, string(body))
		})
	}
}
