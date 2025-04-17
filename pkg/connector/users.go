package connector

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-incident-io/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

// userBuilder manages user-related resources.
type UserBuilder struct {
	resourceType *v2.ResourceType
	client       *client.APIClient
}

// ResourceType returns the type of resource managed by this builder.
func (o *UserBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return userResourceType
}

// List retrieves users and converts them into Baton resources.
func (o *UserBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	bag, pageToken, err := getToken(pToken, userResourceType)
	if err != nil {
		return nil, "", nil, err
	}

	users, nextPageToken, _, err := o.client.ListUsers(ctx, client.PageOptions{
		After:    pageToken,
		PageSize: pToken.Size,
	})
	if err != nil {
		l.Error("Error fetching users", zap.Error(err))
		return nil, "", nil, fmt.Errorf("error fetching users: %w", err)
	}

	var resources []*v2.Resource
	for _, user := range users {
		profile := map[string]interface{}{
			"user_id": user.ID,
			"email":   user.Email,
		}

		userTraits := []resource.UserTraitOption{
			resource.WithUserProfile(profile),
			resource.WithEmail(user.Email, true),
		}

		// Create a Baton user resource
		userResource, err := resource.NewUserResource(
			user.Name,
			userResourceType,
			user.ID,
			userTraits,
			resource.WithParentResourceID(parentResourceID),
		)
		if err != nil {
			return nil, "", nil, fmt.Errorf("error creating user resource: %w", err)
		}

		resources = append(resources, userResource)
	}

	err = bag.Next(nextPageToken)
	if err != nil {
		return nil, "", nil, err
	}

	nextPageToken, err = bag.Marshal()
	if err != nil {
		return nil, "", nil, err
	}

	return resources, nextPageToken, nil, nil
}

// Entitlements always returns an empty slice for users.
func (o *UserBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *UserBuilder) Grants(ctx context.Context, res *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)
	userID := res.Id.Resource

	user, err := o.client.GetUser(ctx, userID)
	if err != nil {
		l.Error("failed to fetch user for grant resolution", zap.String("user_id", userID), zap.Error(err))
		return nil, "", nil, err
	}

	var grants []*v2.Grant

	// BaseRole
	if user.BaseRole.ID != "" {
		baseRoleResource := &v2.Resource{
			Id: &v2.ResourceId{
				ResourceType: baseRoleResourceType.Id,
				Resource:     user.BaseRole.ID,
			},
		}

		grant := grant.NewGrant(
			baseRoleResource,
			"assigned",
			res,
			grant.WithAnnotation(&v2.V1Identifier{
				Id: fmt.Sprintf("base-role-grant:%s:%s", user.BaseRole.ID, userID),
			}),
		)
		grants = append(grants, grant)
	}

	// CustomRoles
	for _, cr := range user.CustomRoles {
		if cr.ID == "" {
			continue
		}

		customRoleResource := &v2.Resource{
			Id: &v2.ResourceId{
				ResourceType: customRoleResourceType.Id,
				Resource:     cr.ID,
			},
		}

		grant := grant.NewGrant(
			customRoleResource,
			"assigned",
			res,
			grant.WithAnnotation(&v2.V1Identifier{
				Id: fmt.Sprintf("custom-role-grant:%s:%s", cr.ID, userID),
			}),
		)
		grants = append(grants, grant)
	}

	return grants, "", nil, nil
}

func NewUserBuilder(c *client.APIClient) *UserBuilder {
	return &UserBuilder{
		resourceType: userResourceType,
		client:       c,
	}
}
