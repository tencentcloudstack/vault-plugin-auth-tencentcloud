package vault_plugin_auth_tencentcloud

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/vault-plugin-auth-tencentcloud/clients"
	"github.com/hashicorp/vault-plugin-auth-tencentcloud/tools"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
	sts "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sts/v20180813"
)

const (
	envVarRunAccTests = "VAULT_ACC"

	// This role must have trusted actors enabled on it.
	envVarAccTestRoleARN = "VAULT_ACC_TEST_ROLE_ARN"

	// The secret id and kek and token given must be for someone who is a trusted actor
	// and thus can assume the given role arn.
	envVarAccTestSecretId        = "VAULT_ACC_TEST_SECRET_ID"
	envVarAccTestSecretKey       = "VAULT_ACC_TEST_SECRET_KEY"
	envVarAccTestToken           = "VAULT_ACC_TEST_TOKEN"
	envClientConfigTestSecretId  = "CLIENT_CONFIG_TEST_SECRET_ID"
	envClientConfigTestSecretKey = "CLIENT_CONFIG_TEST_SECRET_KEY"
)

var runAcceptanceTests = os.Getenv(envVarRunAccTests) == "1"

type testEnv struct {
	ctx     context.Context
	storage logical.Storage
	backend logical.Backend

	isAccTest             bool
	arn                   *arn
	secretId              string
	secretKey             string
	token                 string
	clientConfigSecretId  string
	clientConfigSecretKey string
}

// This test doesn't make real API calls. It injects a fauxRoundTripper
// that intercepts outbound http calls and provides a mocked response.
func TestBackend_Integration(t *testing.T) {
	ctx := context.Background()
	arn, err := parseARN("qcs::cam::uin/1000215438890:roleName/elk")
	if err != nil {
		t.Fatal(err)
	}
	e := testEnv{
		ctx:     ctx,
		storage: &logical.InmemStorage{},
		backend: func() logical.Backend {
			client := cleanhttp.DefaultClient()
			client.Transport = &fauxRoundTripper{}
			b := newBackend(client)
			conf := &logical.BackendConfig{
				System: &logical.StaticSystemView{
					DefaultLeaseTTLVal: time.Hour,
					MaxLeaseTTLVal:     time.Hour,
				},
			}
			if err := b.Setup(ctx, conf); err != nil {
				panic(err)
			}
			return b
		}(),
		isAccTest:             false,
		arn:                   arn,
		secretId:              "someSecretId",
		secretKey:             "someSecretKey",
		token:                 "someToken",
		clientConfigSecretId:  "someClientConfigSecretId",
		clientConfigSecretKey: "someClientConfigSecretKey",
	}

	// Exercise all the role endpoints.
	t.Run("AddClientConfig", e.AddClientConfig)
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
	t.Run("LoginSuccess", e.LoginSuccess)
}

// This test makes real API calls. It's intended for developers and a CI
// test runner.
func TestBackend_Acceptance(t *testing.T) {
	if !runAcceptanceTests {
		t.SkipNow()
	}

	ctx := context.Background()
	arn, err := parseARN(os.Getenv(envVarAccTestRoleARN))
	if err != nil {
		t.Fatal(err)
	}
	e := testEnv{
		ctx:     ctx,
		storage: &logical.InmemStorage{},
		backend: func() logical.Backend {
			client := cleanhttp.DefaultClient()
			b := newBackend(client)
			conf := &logical.BackendConfig{
				System: &logical.StaticSystemView{
					DefaultLeaseTTLVal: time.Hour,
					MaxLeaseTTLVal:     time.Hour,
				},
			}
			if err := b.Setup(ctx, conf); err != nil {
				panic(err)
			}
			return b
		}(),
		isAccTest:             true,
		arn:                   arn,
		secretId:              os.Getenv(envVarAccTestSecretId),
		secretKey:             os.Getenv(envVarAccTestSecretKey),
		token:                 os.Getenv(envVarAccTestToken),
		clientConfigSecretId:  os.Getenv(envClientConfigTestSecretId),
		clientConfigSecretKey: os.Getenv(envClientConfigTestSecretKey),
	}

	// Exercise all the role endpoints.
	t.Run("AddClientConfig", e.AddClientConfig)
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
	t.Run("LoginSuccess", e.LoginSuccess)
}

func (e *testEnv) CreateRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.CreateOperation,
		Path:      "role/" + e.arn.RoleName,
		Storage:   e.storage,
		Data: map[string]interface{}{
			"arn":         e.arn.String(),
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
		Path:      "role/" + e.arn.RoleName,
		Storage:   e.storage,
	}
	resp, err := e.backend.HandleRequest(e.ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected response containing data")
	}
	if resp.Data["arn"] != e.arn.String() {
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
		Path:      "role/" + e.arn.RoleName,
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
		Path:      "role/" + e.arn.RoleName,
		Storage:   e.storage,
	}
	resp, err := e.backend.HandleRequest(e.ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatalf("expected response containing data")
	}
	if resp.Data["arn"] != e.arn.String() {
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

func (e *testEnv) AddClientConfig(t *testing.T) {
	req := &logical.Request{
		Operation: logical.CreateOperation,
		Path:      "config/client",
		Storage:   e.storage,
		Data: map[string]interface{}{
			"secret_id":  e.secretId,
			"secret_key": e.secretKey,
		},
	}
	resp, err := e.backend.HandleRequest(e.ctx, req)
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\nerr:%v", resp, err)
	}
	if resp != nil {
		t.Fatal("expected nil response to represent a 204")
	}
}

func (e *testEnv) DeleteRole(t *testing.T) {
	req := &logical.Request{
		Operation: logical.DeleteOperation,
		Path:      "role/" + e.arn.RoleName,
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

func (e *testEnv) EmptyList(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ListOperation,
		Path:      "role/",
		Storage:   e.storage,
	}
	resp, err := e.backend.HandleRequest(e.ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected response containing data")
	}
	if resp.Data["keys"] != nil {
		t.Fatal("no keys should have been returned")
	}
}

func (e *testEnv) ListOfOne(t *testing.T) {
	req := &logical.Request{
		Operation: logical.ListOperation,
		Path:      "role/",
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
	if resp.Data["keys"].([]string)[0] != e.arn.RoleName {
		t.Fatalf("expected %s but received %s", e.arn.RoleName, resp.Data["keys"].([]string)[0])
	}
}

func (e *testEnv) getIsAccTestCreds(t *testing.T) (creds common.CredentialIface) {
	profile := profile.NewClientProfile()
	profile.Language = "en-US"
	profile.HttpProfile.ReqTimeout = 90
	origCreds := common.NewCredential(e.secretId, e.secretKey)
	client, err := sts.NewClient(origCreds, regions.Ashburn, profile)
	if err != nil {
		t.Fatal(err)
	}
	uid, err := uuid.GenerateUUID()
	if err != nil {
		t.Fatal(err)
	}

	req := sts.NewAssumeRoleRequest()
	arnFull := e.arn.String()
	req.RoleArn = &arnFull
	sessionName := strings.Replace(uid, "-", "", -1)
	req.RoleSessionName = &sessionName
	resp, err := client.AssumeRole(req)
	if err != nil {
		t.Fatal(err)
	}
	creds, err = clients.NewConfigurationCredentialProvider(&clients.Configuration{
		SecretId:  *(resp.Response.Credentials.TmpSecretId),
		SecretKey: *(resp.Response.Credentials.TmpSecretKey),
		Token:     *(resp.Response.Credentials.Token),
	}).GetCredential()
	return
}

func (e *testEnv) LoginSuccess(t *testing.T) {

	var creds common.CredentialIface
	var err error

	if e.isAccTest {
		creds = e.getIsAccTestCreds(t)
	} else {
		creds, err = clients.NewConfigurationCredentialProvider(&clients.Configuration{
			// dummy creds are fine
			SecretId:  e.clientConfigSecretId,
			SecretKey: e.clientConfigSecretKey,
			Token:     e.token,
		}).GetCredential()
	}

	if err != nil {
		t.Fatal(err)
	}

	data, err := tools.GenerateLoginData(e.arn.RoleName, creds, "us-west-2")
	if err != nil {
		t.Fatal(err)
	}
	req := &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      "login",
		Storage:   e.storage,
		Data:      data,
		Connection: &logical.Connection{
			RemoteAddr: "127.0.0.1/24",
		},
	}
	resp, err := e.backend.HandleRequest(e.ctx, req)
	e.checkResp(t, resp, err)
}

// checkResp
func (e *testEnv) checkResp(t *testing.T, resp *logical.Response, err error) {
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
	if resp.Auth.Metadata["account_id"] != e.arn.Uin {
		t.Fatalf("expected %s but received %s", e.arn.Uin, resp.Auth.Metadata["account_id"])
	}
	if resp.Auth.Metadata["role_id"] == "" {
		t.Fatal("expected role_id but received none")
	}
	assumedRoleArn, err := parseARN(resp.Auth.Metadata["arn"])
	if err != nil {
		t.Fatal(err)
	}
	if !assumedRoleArn.IsMemberOf(e.arn) {
		t.Fatalf("assumed role arn of %s is not a member of role arn of %s", assumedRoleArn, e.arn)
	}

	if resp.Auth.Metadata["principal_id"] == "" {
		t.Fatal("expected principal_id but received none")
	}
	if resp.Auth.Metadata["request_id"] == "" {
		t.Fatalf("expected request_id but received none")
	}
	if resp.Auth.Metadata["role_name"] != e.arn.RoleName {
		t.Fatalf("expected %s but received %s", e.arn.RoleName, resp.Auth.Metadata["role_name"])
	}
	if resp.Auth.DisplayName == "" {
		t.Fatal("expected displayname but received none")
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

type fauxRoundTripper struct{}

// This simply returns a spoofed successful response from the GetCallerIdentity endpoint.
func (f *fauxRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	respBody := map[string]map[string]string{
		"Response": {
			"Type":        "CAMRole",
			"AccountId":   "1000262888",
			"UserId":      "461168601842741888:roleSessionName",
			"PrincipalId": "10002618888",
			"Arn":         "qcs::sts:1000262888:assumed-role/461168601842741***",
			"RequestId":   "1c875b55-128b-4152-9e73-0984fd489ba2",
		},
	}
	b, err := json.Marshal(respBody)
	if err != nil {
		return nil, err
	}
	resp := &http.Response{
		Body:       ioutil.NopCloser(bytes.NewReader(b)),
		StatusCode: 200,
	}
	return resp, nil
}
