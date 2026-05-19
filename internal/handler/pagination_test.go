package handler

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePaginationParams(t *testing.T) {
	const maxResults = 100

	tests := []struct {
		name        string
		queryString string
		wantLimit   int
		wantOffset  int
		wantErr     bool
		errContains string
	}{
		{
			name:        "default values - no query params",
			queryString: "",
			wantLimit:   -1,
			wantOffset:  0,
			wantErr:     false,
		},
		{
			name:        "valid limit only",
			queryString: "?limit=10",
			wantLimit:   10,
			wantOffset:  0,
			wantErr:     false,
		},
		{
			name:        "valid offset only",
			queryString: "?offset=20",
			wantLimit:   -1,
			wantOffset:  20,
			wantErr:     false,
		},
		{
			name:        "valid limit and offset",
			queryString: "?limit=10&offset=20",
			wantLimit:   10,
			wantOffset:  20,
			wantErr:     false,
		},
		{
			name:        "limit at max boundary",
			queryString: fmt.Sprintf("?limit=%d", maxResults),
			wantLimit:   maxResults,
			wantOffset:  0,
			wantErr:     false,
		},
		{
			name:        "limit = 1 (minimum)",
			queryString: "?limit=1",
			wantLimit:   1,
			wantOffset:  0,
			wantErr:     false,
		},
		{
			name:        "offset = 0",
			queryString: "?offset=0",
			wantLimit:   -1,
			wantOffset:  0,
			wantErr:     false,
		},
		{
			name:        "limit empty value ignored",
			queryString: "?limit=",
			wantLimit:   -1,
			wantOffset:  0,
			wantErr:     false,
		},
		{
			name:        "offset empty value ignored",
			queryString: "?offset=",
			wantLimit:   -1,
			wantOffset:  0,
			wantErr:     false,
		},
		{
			name:        "limit exceeds max",
			queryString: fmt.Sprintf("?limit=%d", maxResults+1),
			wantErr:     true,
			errContains: "unsupported limit",
		},
		{
			name:        "limit = 0 (invalid)",
			queryString: "?limit=0",
			wantErr:     true,
			errContains: "unsupported limit",
		},
		{
			name:        "limit negative",
			queryString: "?limit=-1",
			wantErr:     true,
			errContains: "unsupported limit",
		},
		{
			name:        "limit not a number",
			queryString: "?limit=abc",
			wantErr:     true,
			errContains: "unsupported limit",
		},
		{
			name:        "offset negative",
			queryString: "?offset=-5",
			wantErr:     true,
			errContains: "unsupported offset",
		},
		{
			name:        "offset not a number",
			queryString: "?offset=abc",
			wantErr:     true,
			errContains: "unsupported offset",
		},
		{
			name:        "valid limit with invalid offset",
			queryString: "?limit=10&offset=abc",
			wantErr:     true,
			errContains: "unsupported offset",
		},
		{
			name:        "invalid limit with valid offset",
			queryString: "?limit=abc&offset=10",
			wantErr:     true,
			errContains: "unsupported limit",
		},
		{
			name:        "extra params ignored",
			queryString: "?limit=10&offset=20&foo=bar",
			wantLimit:   10,
			wantOffset:  20,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test"+tt.queryString, nil)

			params, err := parsePaginationParams(req, maxResults)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantLimit, params.Limit)
				assert.Equal(t, tt.wantOffset, params.Offset)
			}
		})
	}
}

func TestParsePaginationParams_CustomMaxResults(t *testing.T) {
	t.Run("custom max results = 50", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test?limit=50", nil)
		params, err := parsePaginationParams(req, 50)

		require.NoError(t, err)
		assert.Equal(t, 50, params.Limit)
	})

	t.Run("exceeds custom max results", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test?limit=51", nil)
		_, err := parsePaginationParams(req, 50)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported limit")
	})
}
