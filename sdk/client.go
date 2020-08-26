package sdk

import (
	"net/http"

	"github.com/hashicorp/vault-plugin-auth-tencentcloud/sdk/custom"
	"github.com/hashicorp/vault-plugin-auth-tencentcloud/sdk/ratelimit"
	cam "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cam/v20190116"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sts "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sts/v20180813"
)

type Client struct {
	customClient *custom.Client
	camClient    *cam.Client
	stsClient    *sts.Client
}

func NewClient(accessKey, secretKey, secretToken, region string, transport http.RoundTripper) (*Client, error) {
	credential := common.NewCredential(accessKey, secretKey)
	credential.Token = secretToken

	cpf := profile.NewClientProfile()

	customClient, err := custom.NewClient(credential, region, cpf)
	if err != nil {
		return nil, err
	}

	customClient.WithHttpTransport(transport)

	camClient, err := cam.NewClient(credential, region, cpf)
	if err != nil {
		return nil, err
	}

	camClient.WithHttpTransport(transport)

	stsClient, err := sts.NewClient(credential, region, cpf)
	if err != nil {
		return nil, err
	}

	stsClient.WithHttpTransport(transport)

	return &Client{
		customClient: customClient,
		camClient:    camClient,
		stsClient:    stsClient,
	}, nil
}

func (c *Client) GetCallerIdentity() (resp *custom.GetCallerIdentityResponse, err error) {
	request := custom.NewGetCallerIdentityRequest()

	ratelimit.Take(request.GetService())

	return c.customClient.GetCallerIdentity(request)
}

func (c *Client) GetRole(roleId string) (role *cam.RoleInfo, err error) {
	request := cam.NewGetRoleRequest()
	request.RoleId = &roleId

	ratelimit.Take(request.GetService())

	response, err := c.camClient.GetRole(request)
	if err != nil {
		return nil, err
	}

	return response.Response.RoleInfo, nil
}

func (c *Client) GetUserBasicInfo(uin int64) (*custom.GetUserBasicInfoResponse, error) {
	request := custom.NewGetUserBasicInfoRequest()
	request.SubUin = &uin

	ratelimit.Take(request.GetService())

	return c.customClient.GetUserBasicInfo(request)
}
