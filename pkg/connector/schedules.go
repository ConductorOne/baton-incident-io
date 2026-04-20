package connector

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-incident-io/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

// scheduleBuilder handles resource type and client interactions
// for managing schedule resources.
type scheduleBuilder struct {
	resourceType *v2.ResourceType
	client       *client.APIClient
}

// ResourceType returns the resource type associated with schedules.
func (o *scheduleBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return scheduleResourceType
}

// List retrieves a list of schedule resources.
func (o *scheduleBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, opts resource.SyncOpAttrs) ([]*v2.Resource, *resource.SyncOpResults, error) {
	l := ctxzap.Extract(ctx)

	bag, pageToken, err := getToken(&opts.PageToken, scheduleResourceType)
	if err != nil {
		return nil, nil, err
	}

	// Fetch schedules from the API
	resp, nextPageToken, _, err := o.client.ListSchedules(ctx, client.PageOptions{
		After:    pageToken,
		PageSize: opts.PageToken.Size,
	})
	if err != nil {
		l.Error("Error fetching schedules", zap.Error(err))
		return nil, nil, fmt.Errorf("error fetching schedules: %w", err)
	}

	var resources []*v2.Resource

	for _, schedule := range resp {
		scheduleCopy := schedule

		scheduleResource, err := resource.NewGroupResource(
			scheduleCopy.Name,
			scheduleResourceType,
			scheduleCopy.ID,
			nil,
			resource.WithParentResourceID(parentResourceID),
		)
		if err != nil {
			return nil, nil, fmt.Errorf("error creating schedule resource: %w", err)
		}

		resources = append(resources, scheduleResource)
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

// Entitlements returns predefined roles associated with schedules.
func (o *scheduleBuilder) Entitlements(ctx context.Context, teamResource *v2.Resource, opts resource.SyncOpAttrs) ([]*v2.Entitlement, *resource.SyncOpResults, error) {
	entitlementRoles := []string{
		"On_Call",
		"Member",
	}

	var entitlements []*v2.Entitlement
	for _, role := range entitlementRoles {
		entitlements = append(entitlements, entitlement.NewPermissionEntitlement(
			teamResource,
			role,
			entitlement.WithGrantableTo(userResourceType),
			entitlement.WithDisplayName(fmt.Sprintf("Role: %s", role)),
		))
	}

	return entitlements, nil, nil
}

// Grants assigns {'on call','member'} to users based on their schedule participation.
func (o *scheduleBuilder) Grants(ctx context.Context, scheduleResource *v2.Resource, opts resource.SyncOpAttrs) ([]*v2.Grant, *resource.SyncOpResults, error) {
	l := ctxzap.Extract(ctx)

	bag, pageToken, err := getToken(&opts.PageToken, scheduleResourceType)
	if err != nil {
		return nil, nil, err
	}

	// Fetch schedules from the API
	schedulesResp, nextPageToken, _, err := o.client.ListSchedules(ctx, client.PageOptions{
		After:    pageToken,
		PageSize: opts.PageToken.Size,
	})
	if err != nil {
		l.Error("Error fetching schedules", zap.Error(err))
		return nil, nil, fmt.Errorf("error fetching schedules: %w", err)
	}

	var grants []*v2.Grant

	for _, schedule := range schedulesResp {
		scheduleRes := &v2.Resource{
			Id: &v2.ResourceId{
				ResourceType: "schedule",
				Resource:     schedule.ID,
			},
			DisplayName: schedule.Name,
		}

		// users "On Call"
		for _, shift := range schedule.CurrentShifts {
			if shift.User.ID == "" || shift.User.Email == "" || shift.User.ID == "NOBODY" {
				continue
			}

			grant, err := createGrant(scheduleRes, client.User{
				ID:    shift.User.ID,
				Email: shift.User.Email,
			}, "On_Call")
			if err != nil {
				l.Error("Error creating grant", zap.Error(err))
				continue
			}

			if grant != nil {
				grants = append(grants, grant)
			}
		}

		// users "Member"
		seenUsers := make(map[string]bool)
		for _, rotation := range schedule.Config.Rotation {
			for _, user := range rotation.Users {
				if user.ID == "NOBODY" || user.ID == "" || user.Email == "" {
					continue
				}

				if seenUsers[user.ID] {
					l.Debug("Duplicate user detected", zap.String("user_id", user.ID))
					continue
				}

				seenUsers[user.ID] = true

				grant, err := createGrant(scheduleRes, client.User{
					ID:    user.ID,
					Email: user.Email,
				}, "Member")
				if err != nil {
					l.Error("Error creating grant", zap.Error(err))
					continue
				}

				if grant != nil {
					grants = append(grants, grant)
				}
			}
		}
	}

	err = bag.Next(nextPageToken)
	if err != nil {
		return nil, nil, err
	}

	nextPageToken, err = bag.Marshal()
	if err != nil {
		return nil, nil, err
	}

	return grants, &resource.SyncOpResults{NextPageToken: nextPageToken}, nil
}

// createGrant generates a grant for a user with the specified role.
func createGrant(scheduleResource *v2.Resource, user client.User, role string) (*v2.Grant, error) {
	roleResource := &v2.Resource{
		Id: &v2.ResourceId{
			ResourceType: scheduleResourceType.Id,
			Resource:     scheduleResource.Id.Resource,
		},
	}

	principalID, err := resource.NewResourceID(userResourceType, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource ID for user: %s", user.ID)
	}

	return grant.NewGrant(
		roleResource,
		role,
		principalID,
	), nil
}

// newScheduleBuilder initializes a new schedule builder.
func NewScheduleBuilder(c *client.APIClient) *scheduleBuilder {
	return &scheduleBuilder{
		resourceType: scheduleResourceType,
		client:       c,
	}
}
