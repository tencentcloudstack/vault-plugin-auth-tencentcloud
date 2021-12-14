package vault_plugin_auth_tencentcloud

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/vault-plugin-auth-tencentcloud/clients"
	stsLocal "github.com/hashicorp/vault-plugin-auth-tencentcloud/sdk/tencentcloud/sts/v20180813"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/cidrutil"
	"github.com/hashicorp/vault/sdk/logical"
)

func pathLogin(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "login$",
		Fields: map[string]*framework.FieldSchema{
			"role": {
				Type:        framework.TypeString,
				Description: roleDescription,
			},
			"identity_request_url": {
				Type:        framework.TypeString,
				Description: requestUrlDescription,
			},
			"identity_request_headers": {
				Type:        framework.TypeHeader,
				Description: requestHeadersDescription,
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.pathLoginUpdate,
			},
		},
		HelpSynopsis:    pathLoginSyn,
		HelpDescription: pathLoginDesc,
	}
}

//checkData
func checkData(data *framework.FieldData) error {
	b64URL := data.Get("identity_request_url").(string)
	if b64URL == "" {
		return errors.New("missing identity_request_url")
	}
	identityReqURL, err := base64.StdEncoding.DecodeString(b64URL)
	if err != nil {
		return fmt.Errorf("failed to base64 decode identity_request_url: %s", err.Error())
	}
	if _, err := url.Parse(string(identityReqURL)); err != nil {
		return fmt.Errorf("error parsing identity_request_url: %s", err.Error())
	}
	header := data.Get("identity_request_headers").(http.Header)
	if len(header) == 0 {
		return errors.New("missing identity_request_headers")
	}
	return nil
}

// pathLoginUpdate
func (b *backend) pathLoginUpdate(ctx context.Context,
	req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	b64URL := data.Get("identity_request_url").(string)
	identityReqURL, err := base64.StdEncoding.DecodeString(b64URL)
	header := data.Get("identity_request_headers").(http.Header)
	if err := checkData(data); err != nil {
		return nil, err
	}
	callerIdentity, err := b.getCallerIdentity(header, string(identityReqURL))
	if err != nil {
		return nil, errwrap.Wrapf("error making upstream request: {{err}}", err)
	}
	if *(callerIdentity.Response.Type) != "CAMRole" {
		return nil, fmt.Errorf(" %s arn types are not supported at this time", *(callerIdentity.Response.Type))
	}
	parsedARN, err := parseARN(*(callerIdentity.Response.Arn))
	if err != nil {
		return nil, errwrap.Wrapf(fmt.Sprintf(
			"unable to parse entity's arn %s due to {{err}}", *callerIdentity.Response.Arn), err)
	}
	if parsedARN.Type != arnAssumedRoleType {
		return nil, fmt.Errorf(
			"only %s arn types are supported at this time, but %s was provided",
			arnAssumedRoleType, parsedARN.Type)
	}
	// get roleName from tencentCloud
	creds, err := readCredConfig(ctx, req.Storage)
	client, err := clients.NewCAMClient(creds.SecretId, creds.SecretKey)
	if err != nil {
		return nil, err
	}
	parsedRoleName, err := client.GetRoleName(parsedARN.RoleId)
	if err != nil {
		return nil, err
	}
	parsedARN.RoleName = parsedRoleName
	roleName := ""
	roleNameIfc, ok := data.GetOk("role")
	if ok {
		roleName = roleNameIfc.(string)
	}
	if roleName == "" {
		roleName = parsedARN.RoleName
	}
	role, err := readRole(ctx, req.Storage, roleName)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, fmt.Errorf("entry for role %s not found", parsedARN.RoleName)
	}
	if len(role.TokenBoundCIDRs) > 0 {
		if req.Connection == nil {
			b.Logger().Warn("token bound CIDRs found but no connection information available for validation")
			return nil, logical.ErrPermissionDenied
		}
		if !cidrutil.RemoteAddrIsOk(req.Connection.RemoteAddr, role.TokenBoundCIDRs) {
			return nil, logical.ErrPermissionDenied
		}
	}
	if !parsedARN.IsMemberOf(role.ARN) {
		return nil, errors.New("the caller's arn does not match the role's arn")
	}
	auth := makeAuth(callerIdentity, roleName)
	role.PopulateTokenAuth(auth)
	return &logical.Response{
		Auth: auth,
	}, nil
}

// makeAuth
func makeAuth(callerIdentity *stsLocal.GetCallerIdentityResponse, roleName string) (auth *logical.Auth) {
	return &logical.Auth{
		Metadata: map[string]string{
			"role_id":       roleName,
			"arn":           *(callerIdentity.Response.Arn),
			"account_id":    *(callerIdentity.Response.AccountId),
			"user_id":       *(callerIdentity.Response.UserId),
			"principal_id":  *(callerIdentity.Response.PrincipalId),
			"type":          *(callerIdentity.Response.Type),
			"request_id":    *(callerIdentity.Response.RequestId),
			"identity_type": *(callerIdentity.Response.Type),
			"role_name":     roleName,
		},
		DisplayName: *(callerIdentity.Response.PrincipalId),
		Alias: &logical.Alias{
			Name: *(callerIdentity.Response.PrincipalId),
		},
	}
}

// pathLoginRenew
func (b *backend) pathLoginRenew(ctx context.Context,
	req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	// The arn set in metadata earlier is the assumed-role arn.
	arn := req.Auth.Metadata["arn"]
	if arn == "" {
		return nil, errors.New("unable to retrieve arn from metadata during renewal")
	}
	parsedARN, err := parseARN(arn)
	if err != nil {
		return nil, err
	}

	roleName, ok := req.Auth.Metadata["role_name"]
	if !ok {
		return nil, errors.New("error retrieving role_name during renewal")
	}

	role, err := readRole(ctx, req.Storage, roleName)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, errors.New("role entry not found")
	}

	if !parsedARN.IsMemberOf(role.ARN) {
		return nil, errors.New("the caller's arn does not match the role's arn")
	}

	resp := &logical.Response{Auth: req.Auth}
	resp.Auth.TTL = role.TokenTTL
	resp.Auth.MaxTTL = role.TokenMaxTTL
	resp.Auth.Period = role.TokenPeriod
	return resp, nil
}

// getCallerIdentity
func (b *backend) getCallerIdentity(header http.Header, rawURL string) (*stsLocal.GetCallerIdentityResponse, error) {
	/*
		Here we need to ensure we're actually hitting the TencentCloud service, and that the caller didn't
		inject a URL to their own service that will respond as desired.
	*/
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "https" {
		return nil, fmt.Errorf(`expected "https" url scheme but received "%s"`, u.Scheme)
	}
	q := u.Query()
	if header["X-Tc-Action"][0] != "GetCallerIdentity" {
		return nil, fmt.Errorf("query Action must be GetCallerIdentity but received %s", q.Get("Action"))
	}
	request, err := http.NewRequest(http.MethodPost, rawURL, strings.NewReader("{}"))
	if err != nil {
		return nil, err
	}
	if u.Host != "sts.tencentcloudapi.com" {
		return nil, fmt.Errorf(`expected host of "sts.tencentcloudapi.com" but received "%s"`, u.Host)
	}
	request.Header = header
	response, err := b.identityClient.Do(request)
	if err != nil {
		return nil, errwrap.Wrapf("error making request: {{err}}", err)
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		b, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, errwrap.Wrapf("error reading response body: {{err}}", err)
		}
		return nil, fmt.Errorf("received %d checking caller identity: %s", response.StatusCode, b)
	}
	result := &stsLocal.GetCallerIdentityResponse{}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	result.FromJsonString(string(body))
	return result, nil
}

const (
	roleDescription = `Name of the role against which the login is being attempted.
If 'role' is not specified, then the login endpoint looks for a role name in the ARN returned by
the GetCallerIdentity request. If a matching role is not found, login fails.`
	requestUrlDescription     = `Base64-encoded full URL against which to make the TencentCloud request.`
	requestHeadersDescription = `The request headers. This must include the headers over which TencentCloud
has included a signature.`
	pathLoginSyn  = `Authenticates an RAM entity with Vault.`
	pathLoginDesc = `
Authenticate TencentCloud entities using an arbitrary RAM principal.
RAM principals are authenticated by processing a signed sts:GetCallerIdentity
request and then parsing the response to see who signed the request.
`
)
