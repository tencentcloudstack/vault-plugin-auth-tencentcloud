package custom

import (
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

const (
	StsApiVersion = "2018-08-13"
	StsApiService = "sts"

	CamApiVersion = "2019-01-16"
	CamApiService = "cam"
)

type Client struct {
	common.Client
}

func NewClient(credential *common.Credential, region string, clientProfile *profile.ClientProfile) (client *Client, err error) {
	client = &Client{}
	client.Init(region).
		WithCredential(credential).
		WithProfile(clientProfile)
	return
}

func NewGetCallerIdentityRequest() (request *GetCallerIdentityRequest) {
	request = &GetCallerIdentityRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo(StsApiService, StsApiVersion, "GetCallerIdentity")
	return
}

func NewGetCallerIdentityResponse() (response *GetCallerIdentityResponse) {
	response = &GetCallerIdentityResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// 获取当前调用者的身份信息
func (c *Client) GetCallerIdentity(request *GetCallerIdentityRequest) (response *GetCallerIdentityResponse, err error) {
	if request == nil {
		request = NewGetCallerIdentityRequest()
	}
	response = NewGetCallerIdentityResponse()
	err = c.Send(request, response)
	return
}

func NewGetUserBasicInfoRequest() (request *GetUserBasicInfoRequest) {
	request = &GetUserBasicInfoRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo(CamApiService, CamApiVersion, "GetUserBasicInfo")
	return
}

func NewGetUserBasicInfoResponse() (response *GetUserBasicInfoResponse) {
	response = &GetUserBasicInfoResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// 获取用户基本信息
func (c *Client) GetUserBasicInfo(request *GetUserBasicInfoRequest) (response *GetUserBasicInfoResponse, err error) {
	if request == nil {
		request = NewGetUserBasicInfoRequest()
	}
	response = NewGetUserBasicInfoResponse()
	err = c.Send(request, response)
	return
}
