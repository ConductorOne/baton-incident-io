package connector

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-incident-io/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type customRoleBuilder struct {
	resourceType *v2.ResourceType
	client       *client.APIClient
}

// ResourceType returns the type of resource managed by this builder.
func (o *customRoleBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return customRoleResourceType
}

func (o *customRoleBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, opts resource.SyncOpAttrs) ([]*v2.Resource, *resource.SyncOpResults, error) {
	l := ctxzap.Extract(ctx)

	bag, pageToken, err := getToken(&opts.PageToken, customRoleResourceType)
	if err != nil {
		return nil, nil, err
	}

	users, nextPageToken, _, err := o.client.ListUsers(ctx, client.PageOptions{
		After:    pageToken,
		PageSize: opts.PageToken.Size,
	})
	if err != nil {
		l.Error("Error fetching users for custom roles", zap.Error(err))
		return nil, nil, err
	}

	roleMap := make(map[string]client.Role)

	for _, user := range users {
		for _, cr := range user.CustomRoles {
			if cr.ID == "" {
				continue
			}
			if _, exists := roleMap[cr.ID]; !exists {
				roleMap[cr.ID] = cr
			}
		}
	}

	var resources []*v2.Resource
	for _, customRole := range roleMap {
		customRoleCopy := customRole
		groupResource, err := resource.NewRoleResource(
			customRoleCopy.Name,
			customRoleResourceType,
			customRoleCopy.ID,
			nil,
			resource.WithDescription(customRoleCopy.Description),
		)
		if err != nil {
			return nil, nil, err
		}
		resources = append(resources, groupResource)
	}

	err = bag.Next(nextPageToken)
	if err != nil {
		return nil, nil, err
	}

	nextToken, err := bag.Marshal()
	if err != nil {
		return nil, nil, err
	}

	return resources, &resource.SyncOpResults{NextPageToken: nextToken}, nil
}

func (o *customRoleBuilder) Entitlements(_ context.Context, res *v2.Resource, opts resource.SyncOpAttrs) ([]*v2.Entitlement, *resource.SyncOpResults, error) {
	var entitlements []*v2.Entitlement

	entitlementOpts := []entitlement.EntitlementOption{
		entitlement.WithGrantableTo(userResourceType),
		entitlement.WithDescription(fmt.Sprintf("Custom role: %s", res.DisplayName)),
		entitlement.WithDisplayName(fmt.Sprintf("Role: %s", res.DisplayName)),
	}

	entitlements = append(entitlements, entitlement.NewPermissionEntitlement(res, "assigned", entitlementOpts...))
	return entitlements, nil, nil
}

// The logic for role grants is implemented in users.go for performance reasons.
func (o *customRoleBuilder) Grants(ctx context.Context, res *v2.Resource, opts resource.SyncOpAttrs) ([]*v2.Grant, *resource.SyncOpResults, error) {
	return nil, nil, nil
}

func NewCustomRoleBuilder(c *client.APIClient) *customRoleBuilder {
	return &customRoleBuilder{
		resourceType: customRoleResourceType,
		client:       c,
	}
}
