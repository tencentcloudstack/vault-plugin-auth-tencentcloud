// Copyright (c) 2017-2018 THL A29 Limited, a Tencent company. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v20180813

import (
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

const APIVersion = "2018-08-13"

type Client struct {
	common.Client
}

// Deprecated
func NewClientWithSecretId(secretId, secretKey, region string) (client *Client, err error) {
	cpf := profile.NewClientProfile()
	client = &Client{}
	client.Init(region).WithSecretId(secretId, secretKey).WithProfile(cpf)
	return
}

func NewClient(credential common.CredentialIface, region string, clientProfile *profile.ClientProfile) (client *Client, err error) {
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
	request.Init().WithApiInfo("sts", APIVersion, "GetCallerIdentity")
	return
}

func NewGetCallerIdentityResponse() (response *GetCallerIdentityResponse) {
	response = &GetCallerIdentityResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// GetCallerIdentity
// 获取当前调用者的身份信息。
//
// 接口支持主账号，子账号长期密钥以及AssumeRole，GetFederationToken生成的临时凭据的身份获取。
//
// 可能返回的错误码:
//  AUTHFAILURE_ACCESSKEYILLEGAL = "AuthFailure.AccessKeyIllegal"
//  INTERNALERROR_GETSEEDTOKENERROR = "InternalError.GetSeedTokenError"
//  INVALIDPARAMETER_ACCESSKEYNOTSUPPORT = "InvalidParameter.AccessKeyNotSupport"
func (c *Client) GetCallerIdentity(request *GetCallerIdentityRequest) (response *GetCallerIdentityResponse, err error) {
	if request == nil {
		request = NewGetCallerIdentityRequest()
	}
	response = NewGetCallerIdentityResponse()
	err = c.Send(request, response)
	return
}
