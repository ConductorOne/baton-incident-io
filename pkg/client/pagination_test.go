package client

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/conductorone/baton-sdk/pkg/uhttp"
)

// clientServing returns an APIClient whose transport replies with body once. The response
// is built here rather than returned, so it never crosses a function boundary as a loose
// *http.Response.
func clientServing(body string) *APIClient {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
	resp.Header.Set("Content-Type", "application/json")

	return NewClient("token", uhttp.NewBaseHttpClient(&http.Client{
		Transport: stubTransport{response: resp},
	}))
}

// stubTransport is declared here rather than reused from pkg/test, which imports this
// package and would form a cycle.
type stubTransport struct{ response *http.Response }

func (s stubTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return s.response, nil
}

// incident.io keeps pagination_meta on the last page and simply omits after - the recorded
// usersMock.json fixture is exactly {"page_size": 25}. That is the end of the sync.
func TestListUsersLastPage(t *testing.T) {
	c := clientServing(`{"users":[{"id":"u1","name":"Ada","email":"ada@example.com"}],"pagination_meta":{"page_size":25}}`)

	users, after, _, err := c.ListUsers(context.Background(), PageOptions{})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("Expected 1 user, got %d", len(users))
	}
	if after != "" {
		t.Errorf("Expected an empty cursor on the last page, got %q", after)
	}
}

func TestListUsersNextCursor(t *testing.T) {
	c := clientServing(`{"users":[{"id":"u1"}],"pagination_meta":{"page_size":25,"after":"01JPWQ"}}`)

	_, after, _, err := c.ListUsers(context.Background(), PageOptions{})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if after != "01JPWQ" {
		t.Errorf("Expected cursor %q, got %q", "01JPWQ", after)
	}
}

// The case the opt-in exists for. pagination_meta is required on UsersListResultV2, so a
// 200 without it means the block was dropped rather than the list having ended.
func TestListUsersMissingPaginationMeta(t *testing.T) {
	c := clientServing(`{"users":[{"id":"u1"}]}`)

	_, _, _, err := c.ListUsers(context.Background(), PageOptions{})
	if err == nil {
		t.Fatal("Expected an error when pagination_meta is missing")
	}
	if !errors.Is(err, uhttp.ErrMissingPaginationData) {
		t.Fatalf("Expected ErrMissingPaginationData, got %v", err)
	}
}

// GetUser is a single-resource read with no pagination_meta, and it shares doRequest, so it
// must not start demanding one.
func TestGetUserDoesNotRequirePagination(t *testing.T) {
	c := clientServing(`{"user":{"id":"u1","name":"Ada","email":"ada@example.com"}}`)

	user, err := c.GetUser(context.Background(), "u1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if user.ID != "u1" {
		t.Errorf("Expected user id %q, got %q", "u1", user.ID)
	}
}
