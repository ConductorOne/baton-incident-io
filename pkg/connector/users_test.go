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
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
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
			expectedUser.BaseRole = client.Role{
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

const singleUserMockJSON = `{
	"user": {
		"id": "01JPWQNM50YGKQYFJYW61BBPD7",
		"name": "test",
		"email": "test@example.com",
		"base_role": {
			"id": "01JPWQNJKADS4VZ8PEYV0PAQPA",
			"name": "Owner",
			"description": "A base role managed by incident.io for owners of your account.",
			"slug": "owner"
		},
		"custom_roles": [
			{
				"id": "01JPWQNJKAC407555HM47MP2V9",
				"name": "Custom Responder",
				"description": "A custom role.",
				"slug": "custom-responder"
			}
		]
	}
}`

func newSingleUserTestClient() *client.APIClient {
	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(singleUserMockJSON)),
	}
	mockResponse.Header.Set("Content-Type", "application/json")

	return test.NewTestClient(mockResponse, nil)
}

func hasGrantForResourceType(grants []*v2.Grant, resourceType string) bool {
	for _, g := range grants {
		if g.GetEntitlement().GetResource().GetId().GetResourceType() == resourceType {
			return true
		}
	}
	return false
}

func TestUserBuilder_Grants_RespectsSyncFilter(t *testing.T) {
	userResource := &v2.Resource{
		Id: &v2.ResourceId{
			ResourceType: "user",
			Resource:     "01JPWQNM50YGKQYFJYW61BBPD7",
		},
	}

	t.Run("both role types synced emits both grants", func(t *testing.T) {
		testClient := newSingleUserTestClient()
		builder := NewUserBuilder(testClient, true, true)

		grants, _, err := builder.Grants(context.Background(), userResource, resource.SyncOpAttrs{})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if !hasGrantForResourceType(grants, "base-role") {
			t.Errorf("Expected a base-role grant, got %+v", grants)
		}
		if !hasGrantForResourceType(grants, "custom-role") {
			t.Errorf("Expected a custom-role grant, got %+v", grants)
		}
	})

	t.Run("neither role type synced emits no grants", func(t *testing.T) {
		testClient := newSingleUserTestClient()
		builder := NewUserBuilder(testClient, false, false)

		grants, _, err := builder.Grants(context.Background(), userResource, resource.SyncOpAttrs{})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(grants) != 0 {
			t.Errorf("Expected zero grants, got %+v", grants)
		}
	})

	t.Run("only base role synced emits only base-role grant", func(t *testing.T) {
		testClient := newSingleUserTestClient()
		builder := NewUserBuilder(testClient, true, false)

		grants, _, err := builder.Grants(context.Background(), userResource, resource.SyncOpAttrs{})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if !hasGrantForResourceType(grants, "base-role") {
			t.Errorf("Expected a base-role grant, got %+v", grants)
		}
		if hasGrantForResourceType(grants, "custom-role") {
			t.Errorf("Expected no custom-role grant, got %+v", grants)
		}
	})
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
