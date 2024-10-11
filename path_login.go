package vault_plugin_auth_tencentcloud

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/vault-plugin-auth-tencentcloud/clients"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/cidrutil"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
)

func pathLogin(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "login$",
		Fields: map[string]*framework.FieldSchema{
			"role": {
				Type:        framework.TypeString,
				Description: roleDescription,
			},
			"region": {
				Type:        framework.TypeString,
				Description: requestRegionDescription,
			},
			"secret_id": {
				Type:        framework.TypeString,
				Description: requestSecretIdDescription,
			},
			"secret_key": {
				Type:        framework.TypeString,
				Description: requestSecretKeyDescription,
			},
			"token": {
				Type:        framework.TypeString,
				Description: requestTokenDescription,
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

// checkData
func checkData(data *framework.FieldData) error {
	secretId := data.Get("secret_id").(string)
	if secretId == "" {
		return errors.New("missing secret id")
	}
	secretKey := data.Get("secret_key").(string)
	if secretKey == "" {
		return errors.New("missing secret key")
	}
	token := data.Get("token").(string)
	if token == "" {
		return errors.New("missing token")
	}
	return nil
}

// pathLoginUpdate
func (b *backend) pathLoginUpdate(ctx context.Context,
	req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	if err := checkData(data); err != nil {
		return nil, err
	}
	sId := data.Get("secret_id").(string)
	sKey := data.Get("secret_key").(string)
	token := data.Get("token").(string)
	region := data.Get("region").(string)
	if region == "" {
		region = regions.Ashburn
	}

	stsClient, err := clients.NewStsClient(sId, sKey, token, region)
	if err != nil {
		return nil, err
	}
	ciRsp, err := stsClient.GetCallerIdentity()
	if err != nil {
		return nil, err
	}
	if ciRsp.Type != "CAMRole" {
		return nil, fmt.Errorf(" %s arn types are not supported at this time", ciRsp.Type)
	}
	parsedARN, err := parseARN(ciRsp.Arn)
	if err != nil {
		return nil, errwrap.Wrapf(fmt.Sprintf(
			"unable to parse entity's arn %s due to {{err}}", ciRsp.Arn), err)
	}
	if parsedARN.Type != arnAssumedRoleType {
		return nil, fmt.Errorf(
			"only %s arn types are supported at this time, but %s was provided",
			arnAssumedRoleType, parsedARN.Type)
	}

	// get roleName from tencentCloud
	camClient, err := clients.NewCAMClient(sId, sKey, token)
	if err != nil {
		return nil, err
	}
	parsedRoleName, err := camClient.GetRoleName(parsedARN.RoleId)
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
	auth := makeAuth(ciRsp, roleName)
	role.PopulateTokenAuth(auth)
	return &logical.Response{
		Auth: auth,
	}, nil
}

// makeAuth
func makeAuth(callerIdentity *clients.CallerIdentityRsp, roleName string) (auth *logical.Auth) {
	return &logical.Auth{
		Metadata: map[string]string{
			"role_id":       roleName,
			"arn":           callerIdentity.Arn,
			"account_id":    callerIdentity.AccountId,
			"user_id":       callerIdentity.UserId,
			"principal_id":  callerIdentity.PrincipalId,
			"type":          callerIdentity.Type,
			"request_id":    callerIdentity.RequestId,
			"identity_type": callerIdentity.Type,
			"role_name":     roleName,
		},
		DisplayName: callerIdentity.PrincipalId,
		Alias: &logical.Alias{
			Name: callerIdentity.PrincipalId,
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

const (
	roleDescription = `Name of the role against which the login is being attempted.
If 'role' is not specified, then the login endpoint looks for a role name in the ARN returned by
the GetCallerIdentity request. If a matching role is not found, login fails.`

	requestRegionDescription    = `Region parameter, used to identify the region whose data you want to operate.`
	requestSecretIdDescription  = `Temporary certificate key ID. The maximum length is 1024 bytes.`
	requestSecretKeyDescription = `Temporary certificate key. The maximum length is 1024 bytes.`
	requestTokenDescription     = `The length of the token depends on the binding policy and is no longer than 4096 bytes.`

	pathLoginSyn  = `Authenticates an RAM entity with Vault.`
	pathLoginDesc = `
Authenticate TencentCloud entities using an arbitrary RAM principal.
RAM principals are authenticated by processing a signed sts:GetCallerIdentity
request and then parsing the response to see who signed the request.
`
)
