package clients

import (
	"os"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tcerr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
)

// for tools login to STS auth
type EnvProvider struct {
	secretIdENV  string
	secretKeyENV string
	tokenENV     string
}

// DefaultEnvProvider return a default provider
// The default environment variable name are TENCENTCLOUD_SECRET_ID and TENCENTCLOUD_SECRET_KEY and TOKEN
func DefaultEnvProvider() *EnvProvider {
	return &EnvProvider{
		secretIdENV:  "TENCENTCLOUD_SECRET_ID",
		secretKeyENV: "TENCENTCLOUD_SECRET_KEY",
		tokenENV:     "TENCENTCLOUD_TOKEN",
	}
}

// NewEnvProvider uses the name of the environment variable you specified to get the credentials
func NewEnvProvider(secretIdEnvName, secretKeyEnvName, tokenENVName string) *EnvProvider {
	return &EnvProvider{
		secretIdENV:  secretIdEnvName,
		secretKeyENV: secretKeyEnvName,
		tokenENV:     tokenENVName,
	}
}

// GetCredential
func (p *EnvProvider) GetCredential() (common.CredentialIface, error) {
	secretId, ok1 := os.LookupEnv(p.secretIdENV)
	secretKey, ok2 := os.LookupEnv(p.secretKeyENV)
	token, ok3 := os.LookupEnv(p.tokenENV)
	if !ok1 || !ok2 || !ok3 {
		return nil, envNotSet
	}
	if secretId == "" || secretKey == "" || token == "" {
		return nil, tcerr.NewTencentCloudSDKError(creErr,
			"Environmental variable ("+p.secretIdENV+" or "+
				p.secretKeyENV+" or "+p.secretKeyENV+") is empty", "")
	}
	return common.NewTokenCredential(secretId, secretKey, token), nil
}

var creErr = "ClientError.CredentialError"
var envNotSet = tcerr.NewTencentCloudSDKError(creErr, "could not find environmental variable", "")
