package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/ratelimit"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
)

const (
	baseDomain           = "https://api.incident.io/v2"
	getUsersEndpoint     = "/users"
	getSchedulesEndpoint = "/schedules"
)

type APIClient struct {
	apiToken   string
	httpClient *http.Client
}

// NewClient creates a new API client with the provided API token.
func NewClient(apiToken string, httpClient ...*uhttp.BaseHttpClient) *APIClient {
	if httpClient == nil {
		httpClient = []*uhttp.BaseHttpClient{uhttp.NewBaseHttpClient(http.DefaultClient)}
	}

	return &APIClient{
		httpClient: httpClient[0].HttpClient,
		apiToken:   apiToken,
	}
}

// ListSchedules retrieves a list of schedules from the API.
func (c *APIClient) ListSchedules(ctx context.Context, options PageOptions) ([]Schedule, string, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)
	var res ScheduleResponse
	var annotation annotations.Annotations

	queryUrl, err := url.JoinPath(baseDomain, getSchedulesEndpoint)
	if err != nil {
		l.Error(fmt.Sprintf("Error creating UserResponse URL: %s", err))
		return nil, "", nil, err
	}

	annotation, err = c.getResourcesFromAPI(ctx, queryUrl, &res, WithPageAfter(options.After), WithPageLimit(options.PageSize))
	if err != nil {
		l.Error(fmt.Sprintf("Error getting resources: %s", err))
		return nil, "", nil, err
	}

	return res.Schedule, res.Meta.After, annotation, nil
}

// ListUsers retrieves a list of users from the API.
func (c *APIClient) ListUsers(ctx context.Context, options PageOptions) ([]User, string, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)
	var res UserResponse
	var annotation annotations.Annotations

	queryUrl, err := url.JoinPath(baseDomain, getUsersEndpoint)
	if err != nil {
		l.Error(fmt.Sprintf("Error creating UserResponse URL: %s", err))
		return nil, "", nil, err
	}

	annotation, err = c.getResourcesFromAPI(ctx, queryUrl, &res, WithPageAfter(options.After), WithPageLimit(options.PageSize))
	if err != nil {
		l.Error(fmt.Sprintf("Error getting resources: %s", err))
		return nil, "", nil, err
	}

	return res.Users, res.Meta.After, annotation, nil
}

// getResourcesFromAPI makes a GET request to the specified API endpoint.
func (c *APIClient) getResourcesFromAPI(ctx context.Context, urlAddress string, res any, reqOptions ...ReqOpt) (annotations.Annotations, error) {
	_, annotation, err := c.doRequest(ctx, http.MethodGet, urlAddress, &res, reqOptions...)
	if err != nil {
		return nil, err
	}

	return annotation, nil
}

// doRequest executes an HTTP request and processes the response
func (c *APIClient) doRequest(ctx context.Context, method, endpointUrl string, res any,
	reqOptions ...ReqOpt) (http.Header, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	urlAddress, err := url.Parse(endpointUrl)

	if err != nil {
		return nil, nil, err
	}

	for _, o := range reqOptions {
		o(urlAddress)
	}

	req, err := http.NewRequestWithContext(ctx, method, urlAddress.String(), nil)
	if err != nil {
		l.Error(fmt.Sprintf("Error creating request: %s", err))
		return nil, nil, err
	}

	// Manually set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("error in Do: %w", err)
	}

	defer resp.Body.Close()

	if res != nil {
		err = json.NewDecoder(resp.Body).Decode(&res)
		if err != nil {
			return nil, nil, err
		}
	}

	annotation := annotations.Annotations{}
	if resp != nil {
		if desc, err := ratelimit.ExtractRateLimitData(resp.StatusCode, &resp.Header); err == nil {
			annotation.WithRateLimiting(desc)
		} else {
			return nil, annotation, err
		}

		return resp.Header, annotation, nil
	}

	return nil, nil, err
}
