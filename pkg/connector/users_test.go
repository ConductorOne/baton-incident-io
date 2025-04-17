package connector

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/conductorone/baton-incident-io/pkg/client"
	"github.com/conductorone/baton-incident-io/pkg/test"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
)

var pageOptions = client.PageOptions{
	After:    "",
	PageSize: 10,
}

// Tests that the client can fetch users based on the documented API.
func TestIncidentClient_GetUsers(t *testing.T) {
	body, err := test.ReadFile("usersMock.json")
	if err != nil {
		t.Fatalf("Error reading body: %s", err)
	}
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
	mockResponse.Header.Set("Content-Type", "application/json")

	testClient := test.NewTestClient(mockResponse, nil)

	ctx := context.Background()

	result, _, nextOptions, err := testClient.ListUsers(ctx, pageOptions)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	expectedCount := len(test.Users)
	if len(result) != expectedCount {
		t.Errorf("Expected count to be %d, got %d", expectedCount, len(result))
	}

	for index, user := range result {
		expectedUser := client.User{
			ID:    test.Users[index]["id"].(string),
			Name:  test.Users[index]["name"].(string),
			Email: test.Users[index]["email"].(string),
		}

		if baseRoleData, ok := test.Users[index]["base_role"].(map[string]interface{}); ok {
			expectedUser.BaseRole = client.BaseRole{
				ID:          baseRoleData["id"].(string),
				Name:        baseRoleData["name"].(string),
				Description: baseRoleData["description"].(string),
				Slug:        baseRoleData["slug"].(string),
			}
		}

		if user.ID != expectedUser.ID ||
			user.Name != expectedUser.Name ||
			user.Email != expectedUser.Email ||
			!reflect.DeepEqual(user.BaseRole, expectedUser.BaseRole) ||
			len(user.CustomRoles) != len(expectedUser.CustomRoles) {
			t.Errorf("Unexpected user: got %+v, want %+v", user, expectedUser)
		}
	}

	if nextOptions == nil {
		t.Fatal("Expected non-nil nextOptions")
	}
}

func TestIncidentClient_GetUsers_RequestDetails(t *testing.T) {
	var capturedRequest *http.Request

	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Body: io.NopCloser(strings.NewReader(`{
	"users": [],
	"pagination_meta": {
		"page_size": 10,
		"after": ""
	}
}`)),
		Header: make(http.Header),
	}
	mockResponse.Header.Set("Content-Type", "application/json")

	mockTransport := &test.MockRoundTripper{
		Response: mockResponse,
		Err:      nil,
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			capturedRequest = req
			return mockResponse, nil
		},
	}

	httpClient := &http.Client{Transport: mockTransport}
	baseHttpClient := uhttp.NewBaseHttpClient(httpClient)
	testClient := client.NewClient("test", baseHttpClient)

	ctx := context.Background()

	_, _, nextOptions, err := testClient.ListUsers(ctx, pageOptions)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if capturedRequest == nil {
		t.Fatal("capturedRequest is nil — the HTTP request was not captured")
	}

	expectedURL := "https://api.incident.io/v2/users"
	actualURL := capturedRequest.URL.String()
	if !strings.HasPrefix(actualURL, expectedURL) {
		t.Errorf("Expected URL to start with %s, got %s", expectedURL, actualURL)
	}

	expectedHeaders := map[string]string{
		"Accept":        "application/json",
		"Content-Type":  "application/json",
		"Authorization": "Bearer test",
	}

	for key, expectedValue := range expectedHeaders {
		if value := capturedRequest.Header.Get(key); value != expectedValue {
			t.Errorf("Expected header %s to be %s, got %s", key, expectedValue, value)
		}
	}

	if nextOptions == nil {
		t.Fatal("Expected non-nil nextOptions")
	}
}
