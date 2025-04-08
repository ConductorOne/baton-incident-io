package connector

import (
	"context"
	"os"
	"testing"

	"github.com/conductorone/baton-incident-io/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

var (
	ctx              = context.Background()
	parentResourceID = &v2.ResourceId{}
	pToken           = &pagination.Token{Size: 10}
)

func initClient(t *testing.T) *client.APIClient {
	_ = godotenv.Load()

	apiToken := os.Getenv("API_TOKEN")
	if apiToken == "" {
		t.Skipf("Missing required API token")
	}

	return client.NewClient(apiToken)
}

func TestUserBuilderList(t *testing.T) {
	c := initClient(t)

	u := NewUserBuilder(c)

	res, _, _, err := u.List(ctx, parentResourceID, pToken)
	assert.Nil(t, err)
	assert.NotNil(t, res)

	t.Logf("Amount of users obtained: %d", len(res))
}

func TestScheduleBuilderList(t *testing.T) {
	c := initClient(t)

	s := NewScheduleBuilder(c)

	res, _, _, err := s.List(ctx, parentResourceID, pToken)
	assert.Nil(t, err)
	assert.NotNil(t, res)

	t.Logf("Amount of schedules obtained: %d", len(res))
}
