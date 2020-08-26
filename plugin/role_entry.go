package plugin

import (
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-sockaddr"
	"github.com/hashicorp/vault-plugin-auth-tencentcloud/sdk"
	"github.com/hashicorp/vault/sdk/helper/tokenutil"
)

type roleEntry struct {
	tokenutil.TokenParams

	Arn        Arn                           `json:"arn"`
	Policies   []string                      `json:"policies"`
	TTL        time.Duration                 `json:"ttl"`
	MaxTTL     time.Duration                 `json:"max_ttl"`
	Period     time.Duration                 `json:"period"`
	BoundCIDRs []*sockaddr.SockAddrMarshaler `json:"bound_cidrs"`
}

func (r *roleEntry) ResponseData() map[string]interface{} {
	d := map[string]interface{}{
		"arn": r.Arn,
	}

	r.PopulateTokenData(d)

	if len(r.Policies) > 0 {
		d["policies"] = d["token_policies"]
	}

	if len(r.BoundCIDRs) > 0 {
		d["bound_cidrs"] = d["token_bound_cidrs"]
	}

	if r.TTL > 0 {
		d["ttl"] = int64(r.TTL.Seconds())
	}

	if r.MaxTTL > 0 {
		d["max_ttl"] = int64(r.MaxTTL.Seconds())
	}

	if r.Period > 0 {
		d["period"] = int64(r.Period.Seconds())
	}

	return d
}

type Arn string

func GetRoleName(client *sdk.Client, userId string) (string, error) {
	if strings.HasPrefix(userId, "roleSessionName") {
		roleId := userId[:strings.Index(userId, ":roleSessionName")]

		role, err := client.GetRole(roleId)
		if err != nil {
			return "", err
		}

		return *role.RoleName, nil
	}

	var uin int64

	if strings.HasSuffix(userId, "federatedUserName") {
		uinStr := userId[:strings.Index(userId, ":federatedUserName")]

		var err error
		uin, err = strconv.ParseInt(uinStr, 10, 64)
		if err != nil {
			return "", err
		}
	} else {
		var err error
		uin, err = strconv.ParseInt(userId, 10, 64)
		if err != nil {
			return "", err
		}
	}

	userBasicInfo, err := client.GetUserBasicInfo(uin)
	if err != nil {
		return "", err
	}

	return *userBasicInfo.Response.Nickname, nil
}
