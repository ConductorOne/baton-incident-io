package connector

import (
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

var userResourceType = &v2.ResourceType{
	Id:          "user",
	DisplayName: "User",
	Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_USER},
}

var baseRoleResourceType = &v2.ResourceType{
	Id:          "baseRole",
	DisplayName: "baseRole",
}

var customRoleResourceType = &v2.ResourceType{
	Id:          "customRole",
	DisplayName: "customRole",
}

var scheduleResourceType = &v2.ResourceType{
	Id:          "schedule",
	DisplayName: "schedule",
	Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_GROUP},
}
