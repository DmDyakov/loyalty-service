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

func doTestRequestWithAuth(
	t *testing.T,
	h *Handler,
	method,
	endpoint,
	token,
	body,
	contentType string,
) *httptest.ResponseRecorder {
	t.Helper()
	router := h.RegisterRoutes()
	req := httptest.NewRequest(method, endpoint, strings.NewReader(body))
	req.Header.Set("Content-Type", contentType)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}
