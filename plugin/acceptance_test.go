package plugin

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/vault-plugin-auth-tencentcloud/sdk"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	envVarRunAccTests = "VAULT_ACC"

	// The access key and secret given must be for someone who is a trusted actor.
	envVarAccTestAccessKey   = "VAULT_ACC_TEST_ACCESS_KEY"
	envVarAccTestSecretKey   = "VAULT_ACC_TEST_SECRET_KEY"
	envVarAccTestSecretToken = "VAULT_ACC_TEST_SECRET_TOKEN"

	envVarAccTestSecretRegion = "VAULT_ACC_TEST_REGION"
)

func enableAcceptanceTest() bool {
	vaultAcc := strings.ToLower(os.Getenv(envVarRunAccTests))

	return vaultAcc == "TRUE" || vaultAcc == "1"
}

func prepareCallerIdentity(accessKey, secretKey, secretToken, region string) (arn, userId, accountId, principalId, userType string, err error) {
	client, err := sdk.NewClient(accessKey, secretKey, secretToken, region, &sdk.LogRoundTripper{Debug: true})
	if err != nil {
		return "", "", "", "", "", err
	}

	callerIdentity, err := client.GetCallerIdentity()
	if err != nil {
		return "", "", "", "", "", err
	}

	arn = *callerIdentity.Response.Arn
	userId = *callerIdentity.Response.UserId
	accountId = *callerIdentity.Response.AccountId
	principalId = *callerIdentity.Response.PrincipalId
	userType = *callerIdentity.Response.Type

	return
}

func TestBackendAcceptance(t *testing.T) {
	if !enableAcceptanceTest() {
		t.SkipNow()
	}

	accessKey, ok := os.LookupEnv(envVarAccTestAccessKey)
	if !ok {
		t.Fatalf("%s is not set", envVarAccTestAccessKey)
	}

	secretKey, ok := os.LookupEnv(envVarAccTestSecretKey)
	if !ok {
		t.Fatalf("%s is not set", envVarAccTestSecretKey)
	}

	secretToken := os.Getenv(envVarAccTestSecretToken)

	region, ok := os.LookupEnv(envVarAccTestSecretRegion)
	if !ok {
		t.Fatalf("%s is not set", envVarAccTestSecretRegion)
	}

	arn, userId, accountId, principalId, userType, err := prepareCallerIdentity(accessKey, secretKey, secretToken, region)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	e := testEnv{
		ctx:     ctx,
		storage: &logical.InmemStorage{},
		backend: func() logical.Backend {
			b := newBackend(&sdk.LogRoundTripper{Debug: true})

			conf := &logical.BackendConfig{
				System: &logical.StaticSystemView{
					DefaultLeaseTTLVal: time.Hour,
					MaxLeaseTTLVal:     time.Hour,
				},
			}

			if err := b.Setup(ctx, conf); err != nil {
				t.Fatal(err)
			}

			return b
		}(),
		arn:         arn,
		roleName:    "test",
		userId:      userId,
		accessKey:   accessKey,
		secretKey:   secretKey,
		secretToken: secretToken,
		accountId:   accountId,
		principalId: principalId,
		userType:    userType,
		region:      region,
	}

	t.Run("EmptyList", e.EmptyList)
	t.Run("CreateRole", e.CreateRole)
	t.Run("ReadRole", e.ReadRole)
	t.Run("ListOfOne", e.ListOfOne)
	t.Run("UpdateRole", e.UpdateRole)
	t.Run("ReadUpdatedRole", e.ReadUpdatedRole)
	t.Run("ListOfOne", e.ListOfOne)
	t.Run("DeleteRole", e.DeleteRole)
	t.Run("EmptyList", e.EmptyList)

	// Create the role again so we can test logging in.
	t.Run("CreateRole", e.CreateRole)
	t.Run("Login", e.Login)
	t.Run("Renew", e.Renew)
}
