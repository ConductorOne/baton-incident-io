package connector

import (
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/pagination"
)

func getToken(pToken *pagination.Token, resourceType *v2.ResourceType) (*pagination.Bag, string, error) {
	bag, pageToken, err := unmarshalSkipToken(pToken)
	if err != nil {
		return nil, "", err
	}

	if bag.Current() == nil {
		bag.Push(pagination.PageState{
			ResourceTypeID: resourceType.Id,
		})
	}

	return bag, pageToken, nil
}

func unmarshalSkipToken(token *pagination.Token) (*pagination.Bag, string, error) {
	b := &pagination.Bag{}
	err := b.Unmarshal(token.Token)
	if err != nil {
		return nil, "", err
	}

	current := b.Current()
	if current == nil || current.Token == "" {
		return b, "", nil
	}

	return b, current.Token, nil
}
