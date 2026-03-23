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

type baseRoleBuilder struct {
	resourceType *v2.ResourceType
	client       *client.APIClient
}

func (b *baseRoleBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return baseRoleResourceType
}

func (b *baseRoleBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, opts resource.SyncOpAttrs) ([]*v2.Resource, *resource.SyncOpResults, error) {
	l := ctxzap.Extract(ctx)

	bag, pageToken, err := getToken(&opts.PageToken, baseRoleResourceType)
	if err != nil {
		return nil, nil, err
	}

	users, nextPageToken, _, err := b.client.ListUsers(ctx, client.PageOptions{
		After:    pageToken,
		PageSize: opts.PageToken.Size,
	})
	if err != nil {
		l.Error("Error fetching users for base roles", zap.Error(err))
		return nil, nil, fmt.Errorf("error fetching users for base roles: %w", err)
	}

	roleMap := make(map[string]client.Role)

	for _, user := range users {
		if user.BaseRole.ID == "" {
			continue
		}

		if _, exists := roleMap[user.BaseRole.ID]; !exists {
			roleMap[user.BaseRole.ID] = user.BaseRole
		}
	}

	var resources []*v2.Resource
	for _, baseRole := range roleMap {
		baseRoleCopy := baseRole
		roleResource, err := resource.NewRoleResource(
			baseRoleCopy.Name,
			baseRoleResourceType,
			baseRoleCopy.ID,
			nil,
			resource.WithDescription(baseRoleCopy.Description),
		)
		if err != nil {
			return nil, nil, fmt.Errorf("error creating base role resource: %w", err)
		}

		resources = append(resources, roleResource)
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

func (b *baseRoleBuilder) Entitlements(_ context.Context, res *v2.Resource, opts resource.SyncOpAttrs) ([]*v2.Entitlement, *resource.SyncOpResults, error) {
	var entitlements []*v2.Entitlement

	entitlementOpts := []entitlement.EntitlementOption{
		entitlement.WithGrantableTo(userResourceType),
		entitlement.WithDescription(fmt.Sprintf("Base role: %s", res.DisplayName)),
		entitlement.WithDisplayName(fmt.Sprintf("Role: %s", res.DisplayName)),
	}

	entitlements = append(entitlements, entitlement.NewPermissionEntitlement(res, "assigned", entitlementOpts...))
	return entitlements, nil, nil
}

// The logic for role grants is implemented in users.go for performance reasons.
func (b *baseRoleBuilder) Grants(ctx context.Context, res *v2.Resource, opts resource.SyncOpAttrs) ([]*v2.Grant, *resource.SyncOpResults, error) {
	return nil, nil, nil
}

func NewBaseRoleBuilder(c *client.APIClient) *baseRoleBuilder {
	return &baseRoleBuilder{
		resourceType: baseRoleResourceType,
		client:       c,
	}
}
