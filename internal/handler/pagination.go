package handler

import (
	"loyalty-service/internal/errs"
	"net/http"
	"strconv"
)

type PaginationParams struct {
	Limit  int
	Offset int
}

func parsePaginationParams(r *http.Request, maxResults int) (PaginationParams, error) {
	params := PaginationParams{
		Limit: -1,
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > maxResults {
			return PaginationParams{}, errs.ErrUnsupportedLimit
		}
		params.Limit = limit
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			return PaginationParams{}, errs.ErrUnsupportedOffset
		}
		params.Offset = offset
	}

	return params, nil
}
