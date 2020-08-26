package custom

import (
	"encoding/json"

	tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
)

type GetCallerIdentityRequest struct {
	*tchttp.BaseRequest
}

func (r *GetCallerIdentityRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

func (r *GetCallerIdentityRequest) FromJsonString(s string) error {
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

func (r *GetCallerIdentityResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type GetUserBasicInfoRequest struct {
	*tchttp.BaseRequest

	// 子用户UIN，查询主账号时填主账号Uin
	SubUin *int64 `json:"SubUin,omitempty" name:"SubUin"`
}

func (r *GetUserBasicInfoRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

func (r *GetUserBasicInfoRequest) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

type GetUserBasicInfoResponse struct {
	*tchttp.BaseResponse

	Response *struct {
		// 昵称
		Nickname *string `json:"Nickname,omitempty" name:"Nickname"`

		// 手机号码
		PhoneNumber *string `json:"PhoneNumber,omitempty" name:"PhoneNumber"`

		// 邮箱号码
		Email *string `json:"Email,omitempty" name:"Email"`

		// 用户类型
		Type *string `json:"Type,omitempty" name:"Type"`

		// 手机是否验证标记
		PhoneFlag *int64 `json:"PhoneFlag,omitempty" name:"PhoneFlag"`

		// 邮件是否验证标记
		EmailFlag *int64 `json:"EmailFlag,omitempty" name:"EmailFlag"`

		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId *string `json:"RequestId,omitempty" name:"RequestId"`
	} `json:"Response"`
}

func (r *GetUserBasicInfoResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

func (r *GetUserBasicInfoResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}
