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
	"encoding/json"
	tcerr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
)

type GetCallerIdentityRequest struct {
	*tchttp.BaseRequest
}

func (r *GetCallerIdentityRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *GetCallerIdentityRequest) FromJsonString(s string) error {
	f := make(map[string]interface{})
	if err := json.Unmarshal([]byte(s), &f); err != nil {
		return err
	}
	if len(f) > 0 {
		return tcerr.NewTencentCloudSDKError("ClientError.BuildRequestError", "GetCallerIdentityRequest has unknown keys!", "")
	}
	return json.Unmarshal([]byte(s), &r)
}

type GetCallerIdentityResponse struct {
	*tchttp.BaseResponse
	Response *struct {

		// 当前调用者ARN。
		Arn *string `json:"Arn,omitempty" name:"Arn"`

		// 当前调用者所属主账号Uin。
		AccountId *string `json:"AccountId,omitempty" name:"AccountId"`

		// 身份标识。
		// 1. 调用者是云账号时，返回的是当前账号Uin
		// 2. 调用者是角色时，返回的是roleId:roleSessionName
		// 3. 调用者是联合身份时，返回的是uin:federatedUserName
		UserId *string `json:"UserId,omitempty" name:"UserId"`

		// 密钥所属账号Uin。
		// 1. 调用者是云账号，返回的当前账号Uin
		// 2, 调用者是角色，返回的申请角色密钥的账号Uin
		PrincipalId *string `json:"PrincipalId,omitempty" name:"PrincipalId"`

		// 身份类型。
		Type *string `json:"Type,omitempty" name:"Type"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *GetCallerIdentityResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *GetCallerIdentityResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}
