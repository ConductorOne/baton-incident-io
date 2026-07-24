package connector

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-incident-io/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

const roleGrantPermission = "assigned"

// userBuilder manages user-related resources.
type UserBuilder struct {
	resourceType    *v2.ResourceType
	client          *client.APIClient
	syncBaseRoles   bool
	syncCustomRoles bool
}

// ResourceType returns the type of resource managed by this builder.
func (o *UserBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return o.resourceType
}

// List retrieves users and converts them into Baton resources.
func (o *UserBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, opts resource.SyncOpAttrs) ([]*v2.Resource, *resource.SyncOpResults, error) {
	l := ctxzap.Extract(ctx)

	bag, pageToken, err := getToken(&opts.PageToken, userResourceType)
	if err != nil {
		return nil, nil, err
	}

	users, nextPageToken, _, err := o.client.ListUsers(ctx, client.PageOptions{
		After:    pageToken,
		PageSize: opts.PageToken.Size,
	})
	if err != nil {
		l.Error("Error fetching users", zap.Error(err))
		return nil, nil, fmt.Errorf("error fetching users: %w", err)
	}

	var resources []*v2.Resource
	for _, user := range users {
		profile := map[string]interface{}{
			"user_id": user.ID,
			"email":   user.Email,
		}

		userTraits := []resource.UserTraitOption{
			resource.WithEmail(user.Email, true),
		}

		// Create a Baton user resource
		userResource, err := resource.NewUserResource(
			user.Name,
			userResourceType,
			user.ID,
			userTraits,
			resource.WithResourceProfile(profile),
			resource.WithParentResourceID(parentResourceID),
		)
		if err != nil {
			return nil, nil, fmt.Errorf("error creating user resource: %w", err)
		}

		resources = append(resources, userResource)
	}

	err = bag.Next(nextPageToken)
	if err != nil {
		return nil, nil, err
	}

	nextPageToken, err = bag.Marshal()
	if err != nil {
		return nil, nil, err
	}

	return resources, &resource.SyncOpResults{NextPageToken: nextPageToken}, nil
}

// Entitlements always returns an empty slice for users.
func (o *UserBuilder) Entitlements(_ context.Context, res *v2.Resource, opts resource.SyncOpAttrs) ([]*v2.Entitlement, *resource.SyncOpResults, error) {
	return nil, nil, nil
}

// Role grants are implemented here for performance reasons.
func (o *UserBuilder) Grants(ctx context.Context, res *v2.Resource, opts resource.SyncOpAttrs) ([]*v2.Grant, *resource.SyncOpResults, error) {
	l := ctxzap.Extract(ctx)
	userID := res.Id.Resource

	user, err := o.client.GetUser(ctx, userID)
	if err != nil {
		l.Error("failed to fetch user for grant resolution", zap.String("user_id", userID), zap.Error(err))
		return nil, nil, err
	}

	var grants []*v2.Grant

	// BaseRole
	if o.syncBaseRoles && user.BaseRole.ID != "" {
		baseRoleResource := &v2.Resource{
			Id: &v2.ResourceId{
				ResourceType: baseRoleResourceType.Id,
				Resource:     user.BaseRole.ID,
			},
		}

		grant := grant.NewGrant(
			baseRoleResource,
			roleGrantPermission,
			res,
		)
		grants = append(grants, grant)
	}

	// CustomRoles
	if o.syncCustomRoles {
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
				roleGrantPermission,
				res,
			)
			grants = append(grants, grant)
		}
	}

	return grants, nil, nil
}

func NewUserBuilder(c *client.APIClient, syncBaseRoles, syncCustomRoles bool) *UserBuilder {
	resourceType := proto.Clone(userResourceType).(*v2.ResourceType)
	userAnnos := annotations.Annotations(resourceType.GetAnnotations())
	if !syncBaseRoles && !syncCustomRoles {
		// Neither cross-synced role type is enabled, so this builder will
		// never emit a grant -- skip both passes entirely.
		userAnnos.Update(&v2.SkipEntitlementsAndGrants{})
	} else {
		// Users have no entitlements of their own; only the grants pass is
		// conditional on which role types are being synced.
		userAnnos.Update(&v2.SkipEntitlements{})
	}
	resourceType.Annotations = userAnnos

	return &UserBuilder{
		resourceType:    resourceType,
		client:          c,
		syncBaseRoles:   syncBaseRoles,
		syncCustomRoles: syncCustomRoles,
	}
}
