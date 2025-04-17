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

type customRoleBuilder struct {
	resourceType *v2.ResourceType
	client       *client.APIClient
}

// ResourceType returns the type of resource managed by this builder.
func (o *customRoleBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return customRoleResourceType
}

func (o *customRoleBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	bag, pageToken, err := getToken(pToken, customRoleResourceType)
	if err != nil {
		return nil, "", nil, err
	}

	users, nextPageToken, _, err := o.client.ListUsers(ctx, client.PageOptions{
		After:    pageToken,
		PageSize: pToken.Size,
	})
	if err != nil {
		l.Error("Error fetching users for custom roles", zap.Error(err))
		return nil, "", nil, err
	}

	roleMap := make(map[string]client.BaseRole)

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
	for _, cr := range roleMap {
		crCopy := cr
		groupResource, err := resource.NewGroupResource(
			crCopy.Name,
			customRoleResourceType,
			crCopy.ID,
			nil,
			resource.WithDescription(crCopy.Description),
			resource.WithParentResourceID(parentResourceID),
		)
		if err != nil {
			return nil, "", nil, err
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

func (o *customRoleBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var entitlements []*v2.Entitlement

	opts := []entitlement.EntitlementOption{
		entitlement.WithGrantableTo(userResourceType),
		entitlement.WithDescription(fmt.Sprintf("Custom role: %s", resource.DisplayName)),
		entitlement.WithDisplayName(fmt.Sprintf("Role: %s", resource.DisplayName)),
	}

	entitlements = append(entitlements, entitlement.NewPermissionEntitlement(resource, "assigned", opts...))
	return entitlements, "", nil, nil
}

func (o *customRoleBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}
func NewCustomRoleBuilder(c *client.APIClient) *customRoleBuilder {
	return &customRoleBuilder{
		resourceType: customRoleResourceType,
		client:       c,
	}
}
