package clients

import (
	"fmt"

	cam "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cam/v20190116"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
)

// init New CAM Client
func NewCAMClient(secretId, secretKey string) (*CAMClient, error) {
	creds, err := ChainedCredsToCli(secretId, secretKey, "")
	if err != nil {
		return nil, err
	}
	profile := profile.NewClientProfile()
	profile.Language = "en-US"
	profile.HttpProfile.ReqTimeout = 90
	client, err := cam.NewClient(creds, regions.Ashburn, profile)
	if err != nil {
		return nil, err
	}
	return &CAMClient{client: client}, nil
}

// CAM Client
type CAMClient struct {
	client *cam.Client
}

// API： GetRoleName
func (c *CAMClient) GetRoleName(roleId string) (roleName string, err error) {
	req := cam.NewGetRoleRequest()
	req.RoleId = &roleId
	roleRsp, err := c.client.GetRole(req)
	fmt.Println(roleRsp.ToJsonString())
	if err != nil {
		return "", err
	}
	return *(roleRsp.Response.RoleInfo.RoleName), nil

}
