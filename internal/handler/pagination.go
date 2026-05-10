package handler

import (
	"fmt"
	"net/http"
	"strconv"
)

type PaginationParams struct {
	Limit  int
	Offset int
}

func parsePaginationParams(r *http.Request) (PaginationParams, error) {
	limit, err := parseIntParam(r, "limit", 10)
	if err != nil {
		return PaginationParams{}, err
	}
	offset, err := parseIntParam(r, "offset", 0)
	if err != nil {
		return PaginationParams{}, err
	}
	return PaginationParams{Limit: limit, Offset: offset}, nil
}

func parseIntParam(r *http.Request, name string, defaultVal int) (int, error) {
	valStr := r.URL.Query().Get(name)
	if valStr == "" {
		return defaultVal, nil
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", name, err)
	}
	return val, nil
}
