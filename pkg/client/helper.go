package client

import (
	"net/url"
	"strconv"
)

// Default number of items per page if not specified.
const ItemsPerPage = 100

// PageOptions holds pagination parameters.
type PageOptions struct {
	PageSize int
	After    string
}

// WithPageLimit sets the "page_size" query parameter for pagination
func WithPageLimit(pageSize int) ReqOpt {
	if pageSize != 0 {
		return WithQueryParam("page_size", strconv.Itoa(pageSize))
	}
	pageSize = ItemsPerPage
	return WithQueryParam("page_size", strconv.Itoa(pageSize))
}

// WithPageAfter sets the "after" query parameter to continue pagination.
func WithPageAfter(nextPageToken string) ReqOpt {
	return WithQueryParam("after", nextPageToken)
}

// WithQueryParam sets a query parameter on the URL.
func WithQueryParam(key string, value string) ReqOpt {
	return func(reqURL *url.URL) {
		if value != "" {
			q := reqURL.Query()
			q.Set(key, value)
			reqURL.RawQuery = q.Encode()
		}
	}
}

// ReqOpt defines a function that modifies a request URL.
type ReqOpt func(reqURL *url.URL)
