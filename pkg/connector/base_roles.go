package connector

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-incident-io/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
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

func (b *baseRoleBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	bag, pageToken, err := getToken(pToken, baseRoleResourceType)
	if err != nil {
		return nil, "", nil, err
	}

	users, nextPageToken, _, err := b.client.ListUsers(ctx, client.PageOptions{
		After:    pageToken,
		PageSize: pToken.Size,
	})
	if err != nil {
		l.Error("Error fetching users for base roles", zap.Error(err))
		return nil, "", nil, fmt.Errorf("error fetching users for base roles: %w", err)
	}

	roleMap := make(map[string]client.BaseRole)

	for _, user := range users {
		if user.BaseRole.ID == "" {
			continue
		}

		if _, exists := roleMap[user.BaseRole.ID]; !exists {
			roleMap[user.BaseRole.ID] = user.BaseRole
		}
	}

	var resources []*v2.Resource
	for _, br := range roleMap {
		brCopy := br
		groupResource, err := resource.NewGroupResource(
			brCopy.Name,
			baseRoleResourceType,
			brCopy.ID,
			nil,
			resource.WithDescription(brCopy.Description),
			resource.WithParentResourceID(parentResourceID),
		)
		if err != nil {
			return nil, "", nil, fmt.Errorf("error creating base role group resource: %w", err)
		}

		resources = append(resources, groupResource)
	}
	err = bag.Next(nextPageToken)
	if err != nil {
		return nil, "", nil, err
	}

	nextToken, err := bag.Marshal()
	if err != nil {
		return nil, "", nil, err
	}

	return resources, nextToken, nil, nil
}

func (b *baseRoleBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var entitlements []*v2.Entitlement

	opts := []entitlement.EntitlementOption{
		entitlement.WithGrantableTo(userResourceType),
		entitlement.WithDescription(fmt.Sprintf("Base role: %s", resource.DisplayName)),
		entitlement.WithDisplayName(fmt.Sprintf("Role: %s", resource.DisplayName)),
	}

	entitlements = append(entitlements, entitlement.NewPermissionEntitlement(resource, "assigned", opts...))
	return entitlements, "", nil, nil
}

func (b *baseRoleBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func NewBaseRoleBuilder(c *client.APIClient) *baseRoleBuilder {
	return &baseRoleBuilder{
		resourceType: baseRoleResourceType,
		client:       c,
	}
}
