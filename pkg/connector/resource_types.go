package connector

import (
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

const (
	BaseRoleResourceTypeID   = "base-role"
	CustomRoleResourceTypeID = "custom-role"
)

var userResourceType = &v2.ResourceType{
	Id:          "user",
	DisplayName: "User",
	Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_USER},
}

var baseRoleResourceType = &v2.ResourceType{
	Id:          BaseRoleResourceTypeID,
	DisplayName: "base Role",
}

var customRoleResourceType = &v2.ResourceType{
	Id:          CustomRoleResourceTypeID,
	DisplayName: "custom Role",
}

var scheduleResourceType = &v2.ResourceType{
	Id:          "schedule",
	DisplayName: "schedule",
	Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_GROUP},
}
