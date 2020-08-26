package plugin

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/hashicorp/vault-plugin-auth-tencentcloud/sdk/custom"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/cidrutil"
	"github.com/hashicorp/vault/sdk/logical"
)

func pathLogin(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "login$",
		Fields: map[string]*framework.FieldSchema{
			"role": {
				Type:     framework.TypeString,
				Required: true,
				Description: `Name of the role against which the login is being attempted.
If "role" is not specified, then the login endpoint looks for a role name in the ARN returned by 
the GetCallerIdentity request. If a matching role is not found, login fails.`,
			},
			"request_url": {
				Type:        framework.TypeString,
				Required:    true,
				Description: "The request url, it must be encoded by base64.",
			},
			"signed_header": {
				Type:     framework.TypeHeader,
				Required: true,
				Description: `The request headers. This must include the headers over which TencentCloud
has included a signature.`,
			},
			"request_body": {
				Type:        framework.TypeString,
				Required:    true,
				Description: "The request body, it must be encoded by base64.",
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

func (b *backend) pathLoginUpdate(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	roleName := data.Get("role").(string)

	role, err := readRole(ctx, req.Storage, roleName)
	if err != nil {
		return nil, err
	}

	if role == nil {
		return nil, fmt.Errorf("entry for role %s not found", roleName)
	}

	requestUrl, err := base64.StdEncoding.DecodeString(data.Get("request_url").(string))
	if err != nil {
		return nil, fmt.Errorf("decode base64 request_url failed: %w", err)
	}

	if _, err := url.Parse(string(requestUrl)); err != nil {
		return nil, fmt.Errorf("parse decoded reuqest_url failed: %w", err)
	}

	header := data.Get("signed_header").(http.Header)
	if len(header) == 0 {
		return nil, errors.New("missing signed_header")
	}

	requestBody, err := base64.StdEncoding.DecodeString(data.Get("request_body").(string))
	if err != nil {
		return nil, fmt.Errorf("decode base64 request_body failed: %w", err)
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, string(requestUrl), bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("generate http request failed: %w", err)
	}

	httpRequest.Header = header

	httpClient := new(http.Client)
	httpClient.Transport = b.transport

	response, err := httpClient.Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("getCallerIdentity failed: %w", err)
	}

	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("getCallerIdentity response status code is %d, not %d", response.StatusCode, http.StatusOK)
	}

	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read getCallerIdentity response failed: %w", err)
	}

	getCallerIdentityResponse := custom.NewGetCallerIdentityResponse()

	if err := getCallerIdentityResponse.ParseErrorFromHTTPResponse(respBody); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(respBody, getCallerIdentityResponse); err != nil {
		return nil, fmt.Errorf("unmarshal getCallerIdentity response failed: %w", err)
	}

	if Arn(*getCallerIdentityResponse.Response.Arn) != role.Arn {
		return nil, errors.New("the caller's arn does not match the role's arn")
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

	auth := &logical.Auth{
		Metadata: map[string]string{
			"account_id":   *getCallerIdentityResponse.Response.AccountId,
			"user_id":      *getCallerIdentityResponse.Response.UserId,
			"arn":          *getCallerIdentityResponse.Response.Arn,
			"type":         *getCallerIdentityResponse.Response.Type,
			"principal_id": *getCallerIdentityResponse.Response.PrincipalId,
			"request_id":   *getCallerIdentityResponse.Response.RequestId,
			"role_name":    roleName,
		},
		DisplayName: *getCallerIdentityResponse.Response.UserId,
		Alias: &logical.Alias{
			Name: *getCallerIdentityResponse.Response.UserId,
		},
	}

	auth.Renewable = true

	role.PopulateTokenAuth(auth)

	return &logical.Response{
		Auth: auth,
	}, nil
}

func (b *backend) pathLoginRenew(ctx context.Context, req *logical.Request, _ *framework.FieldData) (*logical.Response, error) {
	arn := req.Auth.Metadata["arn"]
	if arn == "" {
		return nil, errors.New("unable to retrieve arn from metadata during renewal")
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

	if Arn(arn) != role.Arn {
		return nil, errors.New("the caller's arn does not match the role's arn")
	}

	resp := &logical.Response{Auth: req.Auth}
	resp.Auth.TTL = role.TokenTTL
	resp.Auth.MaxTTL = role.TokenMaxTTL
	resp.Auth.Period = role.TokenPeriod

	return resp, nil
}

const pathLoginSyn = `
Authenticates an CAM entity with Vault.
`

const pathLoginDesc = `
Authenticate TencentCloud entities using an arbitrary CAM principal.

CAM principals are authenticated by processing a signed sts:GetCallerIdentity
request and then parsing the response to see who signed the request.
`
