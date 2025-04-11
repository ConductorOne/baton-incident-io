package test

import (
	"net/http"
	"os"

	"github.com/conductorone/baton-incident-io/pkg/client"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
)

var (
	Users = []map[string]interface{}{
		{
			"id":    "01JPWQNM50YGKQYFJYW61BBPD7",
			"name":  "test",
			"email": "test@example.com",
		},
		{
			"id":    "01JPWQP39ZE3X1NRHC3PJAWZVQ",
			"name":  "Alejandro",
			"email": "alejandro@example.com",
		},
	}
)

// Custom RoundTripper for testing.
type TestRoundTripper struct {
	response *http.Response
	err      error
}

type MockRoundTripper struct {
	Response      *http.Response
	Err           error
	RoundTripFunc func(*http.Request) (*http.Response, error)
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.RoundTripFunc != nil {
		return m.RoundTripFunc(req)
	}
	return m.Response, m.Err
}

func (t *TestRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return t.response, t.err
}

// Helper function to create a test client with custom transport.
func NewTestClient(response *http.Response, err error) *client.APIClient {
	transport := &TestRoundTripper{response: response, err: err}
	httpClient := &http.Client{Transport: transport}
	baseHttpClient := uhttp.NewBaseHttpClient(httpClient)

	bearerToken := "test"

	newClientT := client.NewClient(bearerToken, baseHttpClient)

	return newClientT
}

func ReadFile(fileName string) (string, error) {
	data, err := os.ReadFile("../test/mockResponses/" + fileName)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
