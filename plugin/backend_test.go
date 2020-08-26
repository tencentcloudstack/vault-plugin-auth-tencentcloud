package plugin

import (
	"context"
	"encoding/base64"
	"testing"
	"time"

	"github.com/hashicorp/vault-plugin-auth-tencentcloud/cli"
	"github.com/hashicorp/vault/sdk/logical"
)

func TestBackendIntegration(t *testing.T) {
	ctx := context.Background()

	arn := "qcs::sts:1000262233:federated-user/10002616666"
	userId := "10002616666:federatedUserName"
	accountId := "1000262233"
	principalId := "10002616666"
	userType := "CAMUser"

	transport := &fakeTransport{
		arn:         Arn(arn),
		accountId:   accountId,
		userId:      userId,
		principalId: principalId,
		userType:    userType,
	}

	e := testEnv{
		ctx:     ctx,
		storage: &logical.InmemStorage{},
		backend: func() logical.Backend {
			b := newBackend(transport)

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
		accessKey:   "ak",
		secretKey:   "sk",
		secretToken: "token",
		accountId:   accountId,
		principalId: principalId,
		userType:    userType,
		region:      "ap-guangzhou",
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

func (e *testEnv) EmptyList(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ListOperation,
		Path:      "roles/",
		Storage:   e.storage,
	}

	resp, err := e.backend.HandleRequest(e.ctx, req)
	if err != nil {
		t.Fatal(err)
	}

	if resp == nil {
		t.Fatal("expected an response")
	}

	if resp.Data["keys"] != nil {
		t.Fatal("no keys should have been returned")
	}
}

func (e *testEnv) CreateRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.CreateOperation,
		Path:      "roles/test",
		Storage:   e.storage,
		Data: map[string]interface{}{
			"arn":         e.arn,
			"policies":    "default",
			"ttl":         10,
			"max_ttl":     10,
			"period":      1,
			"bound_cidrs": []string{"127.0.0.1/24"},
		},
	}

	resp, err := e.backend.HandleRequest(e.ctx, req)
	if err != nil {
		t.Fatal(err)
	}

	if resp != nil {
		t.Fatal("expected nil response to represent a 204")
	}
}

func (e *testEnv) ReadRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "roles/test",
		Storage:   e.storage,
	}

	resp, err := e.backend.HandleRequest(e.ctx, req)
	if err != nil {
		t.Fatal(err)
	}

	if resp == nil {
		t.Fatal("expected an response")
	}

	if resp.Data["arn"].(Arn) != Arn(e.arn) {
		t.Fatalf("expected arn of %s but received %s", e.arn, resp.Data["arn"])
	}

	if resp.Data["policies"].([]string)[0] != "default" {
		t.Fatalf("expected policy of default but received %s", resp.Data["policies"].([]string)[0])
	}

	if resp.Data["ttl"].(int64) != 10 {
		t.Fatalf("expected ttl of 10 but received %d", resp.Data["ttl"].(time.Duration))
	}

	if resp.Data["max_ttl"].(int64) != 10 {
		t.Fatalf("expected max_ttl of 10 but received %d", resp.Data["max_ttl"].(time.Duration))
	}

	if resp.Data["period"].(int64) != 1 {
		t.Fatalf("expected period of 1 but received %d", resp.Data["period"].(time.Duration))
	}
}

func (e *testEnv) UpdateRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      "roles/test",
		Storage:   e.storage,
		Data: map[string]interface{}{
			"max_ttl": 100,
		},
	}

	resp, err := e.backend.HandleRequest(e.ctx, req)
	if err != nil {
		t.Fatal(err)
	}

	if resp != nil {
		t.Fatal("expected nil response to represent a 204")
	}
}

func (e *testEnv) ReadUpdatedRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "roles/test",
		Storage:   e.storage,
	}

	resp, err := e.backend.HandleRequest(e.ctx, req)
	if err != nil {
		t.Fatal(err)
	}

	if resp == nil {
		t.Fatalf("expected response containing data")
	}

	if resp.Data["arn"] != Arn(e.arn) {
		t.Fatalf("expected arn of %s but received %s", e.arn, resp.Data["arn"])
	}

	if resp.Data["policies"].([]string)[0] != "default" {
		t.Fatalf("expected policy of default but received %s", resp.Data["policies"].([]string)[0])
	}

	if resp.Data["ttl"].(int64) != 10 {
		t.Fatalf("expected ttl of 10 but received %d", resp.Data["ttl"].(time.Duration))
	}

	if resp.Data["max_ttl"].(int64) != 100 {
		t.Fatalf("expected max_ttl of 100 but received %d", resp.Data["max_ttl"].(time.Duration))
	}

	if resp.Data["period"].(int64) != 1 {
		t.Fatalf("expected period of 1 but received %d", resp.Data["period"].(time.Duration))
	}
}

func (e *testEnv) DeleteRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.DeleteOperation,
		Path:      "roles/test",
		Storage:   e.storage,
	}

	resp, err := e.backend.HandleRequest(e.ctx, req)
	if err != nil {
		t.Fatal(err)
	}

	if resp != nil {
		t.Fatal("expected nil response to represent a 204")
	}
}

func (e *testEnv) ListOfOne(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ListOperation,
		Path:      "roles/",
		Storage:   e.storage,
	}

	resp, err := e.backend.HandleRequest(e.ctx, req)
	if err != nil {
		t.Fatal(err)
	}

	if resp == nil {
		t.Fatal("expected response containing data")
	}

	if len(resp.Data["keys"].([]string)) != 1 {
		t.Fatal("1 key should have been returned")
	}

	if resp.Data["keys"].([]string)[0] != "test" {
		t.Fatalf("expected %s but received %s", "test", resp.Data["keys"].([]string)[0])
	}
}

func (e *testEnv) Login(t *testing.T) {
	requestUrl, requestBody, signedHeader, err := cli.DumpCallerIdentityRequest(e.accessKey, e.secretKey, e.secretToken, e.region)
	if err != nil {
		t.Fatal(err)
	}

	loginData := map[string]interface{}{
		"role":          e.roleName,
		"request_url":   base64.StdEncoding.EncodeToString([]byte(requestUrl)),
		"signed_header": signedHeader,
		"request_body":  base64.StdEncoding.EncodeToString(requestBody),
	}

	req := &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      "login",
		Storage:   e.storage,
		Data:      loginData,
		Connection: &logical.Connection{
			RemoteAddr: "127.0.0.1/24",
		},
	}

	resp, err := e.backend.HandleRequest(e.ctx, req)
	if err != nil {
		t.Fatal(err)
	}

	if resp == nil {
		t.Fatal("expected response containing data")
	}

	if resp.Auth == nil {
		t.Fatal("should have received an auth")
	}

	if resp.Auth.Period != time.Second {
		t.Fatalf("expected period of 1 second but received %d", resp.Auth.Period)
	}

	if len(resp.Auth.Policies) != 1 {
		t.Fatalf("expected 1 policy but received %d", len(resp.Auth.Policies))
	}

	if resp.Auth.Policies[0] != "default" {
		t.Fatalf("expected default but received %s", resp.Auth.Policies[0])
	}

	if resp.Auth.Metadata["account_id"] != e.accountId {
		t.Fatalf("expected %s but received %s", e.accountId, resp.Auth.Metadata["account_id"])
	}

	if resp.Auth.Metadata["arn"] != e.arn {
		t.Fatalf("expected arn %s, but received %v", e.arn, resp.Auth.Metadata["arn"])
	}

	if resp.Auth.Metadata["principal_id"] != e.principalId {
		t.Fatalf("expected principal_id %s, but received %v", e.principalId, resp.Auth.Metadata["principal_id"])
	}

	if resp.Auth.Metadata["request_id"] == "" {
		t.Fatalf("expected request_id but received none")
	}

	if resp.Auth.Metadata["role_name"] != e.roleName {
		t.Fatalf("expected role_name %s but received %s", e.roleName, resp.Auth.Metadata["role_name"])
	}

	if resp.Auth.Metadata["type"] != e.userType {
		t.Fatalf("expected type %s but received %s", e.userType, resp.Auth.Metadata["type"])
	}

	if resp.Auth.DisplayName != e.userId {
		t.Fatalf("expected displa yname %s, but received %s", e.userId, resp.Auth.DisplayName)
	}

	if !resp.Auth.LeaseOptions.Renewable {
		t.Fatal("auth should be renewable")
	}

	if resp.Auth.LeaseOptions.TTL != 10*time.Second {
		t.Fatal("ttl should be 10 seconds")
	}

	if resp.Auth.LeaseOptions.MaxTTL != 10*time.Second {
		t.Fatal("max ttl should be 10 seconds")
	}

	if resp.Auth.Alias.Name == "" {
		t.Fatal("expected alias name but received none")
	}

	e.mostRecentAuth = resp.Auth
}

func (e *testEnv) Renew(t *testing.T) {
	req := &logical.Request{
		Operation: logical.RenewOperation,
		Path:      "login",
		Auth:      e.mostRecentAuth,
		Storage:   e.storage,
		Connection: &logical.Connection{
			RemoteAddr: "127.0.0.1/24",
		},
	}

	resp, err := e.backend.HandleRequest(e.ctx, req)
	if err != nil {
		t.Fatal(err)
	}

	if resp == nil {
		t.Fatal("expected response containing data")
	}

	if resp.Auth == nil {
		t.Fatal("should have received an auth")
	}

	if resp.Auth.Period != time.Second {
		t.Fatalf("expected period of 1 second but received %d", resp.Auth.Period)
	}

	if len(resp.Auth.Policies) != 1 {
		t.Fatalf("expected 1 policy but received %d", len(resp.Auth.Policies))
	}

	if resp.Auth.Policies[0] != "default" {
		t.Fatalf("expected default but received %s", resp.Auth.Policies[0])
	}

	if resp.Auth.Metadata["account_id"] != e.accountId {
		t.Fatalf("expected %s but received %s", e.accountId, resp.Auth.Metadata["account_id"])
	}

	if resp.Auth.Metadata["arn"] != e.arn {
		t.Fatalf("expected arn %s, but received %v", e.arn, resp.Auth.Metadata["arn"])
	}

	if resp.Auth.Metadata["principal_id"] != e.principalId {
		t.Fatalf("expected principal_id %s, but received %v", e.principalId, resp.Auth.Metadata["principal_id"])
	}

	if resp.Auth.Metadata["request_id"] == "" {
		t.Fatalf("expected request_id but received none")
	}

	if resp.Auth.Metadata["role_name"] != e.roleName {
		t.Fatalf("expected role_name %s but received %s", e.roleName, resp.Auth.Metadata["role_name"])
	}

	if resp.Auth.Metadata["type"] != e.userType {
		t.Fatalf("expected type %s but received %s", e.userType, resp.Auth.Metadata["type"])
	}

	if resp.Auth.DisplayName != e.userId {
		t.Fatalf("expected displa yname %s, but received %s", e.userId, resp.Auth.DisplayName)
	}

	if !resp.Auth.LeaseOptions.Renewable {
		t.Fatal("auth should be renewable")
	}

	if resp.Auth.LeaseOptions.TTL != 10*time.Second {
		t.Fatal("ttl should be 10 seconds")
	}

	if resp.Auth.LeaseOptions.MaxTTL != 10*time.Second {
		t.Fatal("max ttl should be 10 seconds")
	}

	if resp.Auth.Alias.Name == "" {
		t.Fatal("expected alias name but received none")
	}
}
