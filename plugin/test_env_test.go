package plugin

import (
	"context"

	"github.com/hashicorp/vault/sdk/logical"
)

type testEnv struct {
	ctx            context.Context
	storage        logical.Storage
	backend        logical.Backend
	mostRecentAuth *logical.Auth

	arn         string
	roleName    string
	userId      string
	accessKey   string
	secretKey   string
	secretToken string
	accountId   string
	principalId string
	userType    string
	region      string
}
