package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func doTestRequest(t *testing.T, method, endpoint, body, contentType string) (*http.Request, *httptest.ResponseRecorder) {
	t.Helper()
	b := strings.NewReader(body)
	req := httptest.NewRequest(method, endpoint, b)
	req.Header.Set("Content-Type", contentType)

	w := httptest.NewRecorder()

	return req, w
}
