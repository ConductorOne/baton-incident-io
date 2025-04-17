package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

const (
	baseDomain           = "https://api.incident.io/v2"
	getUsersEndpoint     = "/users"
	getSchedulesEndpoint = "/schedules"
)

type APIClient struct {
	apiToken string
	wrapper  *uhttp.BaseHttpClient
}

// NewClient creates a new API client with the provided API token.
func NewClient(apiToken string, httpClient *uhttp.BaseHttpClient) *APIClient {
	if httpClient == nil {
		httpClient = uhttp.NewBaseHttpClient(http.DefaultClient)
	}

	return &APIClient{
		wrapper:  httpClient,
		apiToken: apiToken,
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

// doRequest executes an HTTP request and processes the response.
func (c *APIClient) doRequest(ctx context.Context, method, endpointUrl string, res any,
	reqOptions ...ReqOpt) (http.Header, annotations.Annotations, error) {
	logger := ctxzap.Extract(ctx)

	urlAddress, err := url.Parse(endpointUrl)
	if err != nil {
		return nil, nil, err
	}

	for _, o := range reqOptions {
		o(urlAddress)
	}

	options := []uhttp.RequestOption{
		uhttp.WithContentTypeJSONHeader(),
		uhttp.WithAcceptJSONHeader(),
		uhttp.WithBearerToken(c.apiToken),
	}

	request, err := c.wrapper.NewRequest(ctx, method, urlAddress, options...)
	if err != nil {
		logger.Error("failed to create request", zap.Error(err))
		return nil, nil, err
	}

	annotation := annotations.Annotations{}
	doOptions := []uhttp.DoOption{}

	if res != nil {
		doOptions = append(doOptions, uhttp.WithJSONResponse(res))
	}

	response, err := c.wrapper.Do(request, doOptions...)
	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}

	if err != nil {
		return nil, annotation, fmt.Errorf("error in Do: %w", err)
	}

	return response.Header, annotation, nil
}

func (c *APIClient) GetUser(ctx context.Context, userID string) (*User, error) {
	l := ctxzap.Extract(ctx)

	var res SingleUserResponse

	endpointURL, err := url.JoinPath(baseDomain, getUsersEndpoint, userID)
	if err != nil {
		l.Error("failed to build user URL", zap.Error(err))
		return nil, err
	}

	_, _, err = c.doRequest(ctx, http.MethodGet, endpointURL, &res)
	if err != nil {
		l.Error("failed to get user", zap.Error(err))
		return nil, err
	}

	return &res.User, nil
}
