package clients

import (
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sts "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sts/v20180813"
)

// NewStsClient init New STS Client
func NewStsClient(secretId, secretKey, token, region string) (*STSClient, error) {
	creds, err := ChainedCredsToCli(secretId, secretKey, token)
	if err != nil {
		return nil, err
	}
	profile := profile.NewClientProfile()
	profile.Language = "en-US"
	profile.HttpProfile.ReqTimeout = 90
	client, err := sts.NewClient(creds, region, profile)
	if err != nil {
		return nil, err
	}
	return &STSClient{client: client}, nil
}

// STSClient STS Client
type STSClient struct {
	client *sts.Client
}

// CallerIdentityRsp caller identity response
type CallerIdentityRsp struct {
	Arn         string
	AccountId   string
	UserId      string
	PrincipalId string
	Type        string
	RequestId   string
}

// GetCallerIdentity get caller identity
func (c *STSClient) GetCallerIdentity() (rsp *CallerIdentityRsp, err error) {
	req := sts.NewGetCallerIdentityRequest()
	callerIdentityRsp, err := c.client.GetCallerIdentity(req)
	if err != nil {
		return nil, err
	}
	return &CallerIdentityRsp{
		Type:        *callerIdentityRsp.Response.Type,
		Arn:         *callerIdentityRsp.Response.Arn,
		AccountId:   *callerIdentityRsp.Response.AccountId,
		UserId:      *callerIdentityRsp.Response.UserId,
		PrincipalId: *callerIdentityRsp.Response.PrincipalId,
		RequestId:   *callerIdentityRsp.Response.RequestId,
	}, nil
}
