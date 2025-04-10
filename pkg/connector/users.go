package connector

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-incident-io/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"go.uber.org/zap"
)

// userBuilder manages user-related resources.
type UserBuilder struct {
	resourceType *v2.ResourceType
	client       *client.APIClient
	logger       *zap.Logger
}

// ResourceType returns the type of resource managed by this builder.
func (o *UserBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return userResourceType
}

// List retrieves users and converts them into Baton resources.
func (o *UserBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {

	bag, pageToken, err := getToken(pToken, userResourceType)
	if err != nil {
		return nil, "", nil, err
	}

	users, nextPageToken, _, err := o.client.ListUsers(ctx, client.PageOptions{
		After:    pageToken,
		PageSize: pToken.Size,
	})
	if err != nil {
		o.logger.Error("Error fetching users", zap.Error(err))
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
func (o *UserBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}
func NewUserBuilder(c *client.APIClient) *UserBuilder {
	return &UserBuilder{
		resourceType: userResourceType,
		client:       c,
		logger:       zap.NewExample(),
	}
}
