package cli

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/vault/api"
)

type CliHandler struct{}

const mountPath = "tencentcloud"

func (c *CliHandler) Auth(client *api.Client, m map[string]string) (*api.Secret, error) {
	mount := m["mount"]
	if mount == "" {
		mount = mountPath
	}

	role := m["role"]

	accessKey := m["access_key"]
	secretKey := m["secret_key"]
	secretToken := m["security_token"]
	region := m["region"]

	requestUrl, requestBody, signedHeader, err := DumpCallerIdentityRequest(accessKey, secretKey, secretToken, region)
	if err != nil {
		return nil, err
	}

	loginData := map[string]interface{}{
		"role":          role,
		"request_url":   base64.StdEncoding.EncodeToString([]byte(requestUrl)),
		"signed_header": signedHeader,
		"request_body":  base64.StdEncoding.EncodeToString(requestBody),
	}

	path := fmt.Sprintf("auth/%s/login", mount)

	secret, err := client.Logical().Write(path, loginData)
	if err != nil {
		return nil, err
	}

	if secret == nil {
		return nil, errors.New("empty response from credential provider")
	}

	return secret, nil
}

func (c *CliHandler) Help() string {
	help := `
Usage: vault login -method=tencentcloud [CONFIG K=V...]

  The TencentCloud auth method allows users to authenticate with TencentCloud CAM credentials.

  The TencentCloud CAM credentials may be specified explicitly via the command line:

      $ vault login -method=tencentcloud access_key=... secret_key=... security_token=... region=...

Configuration:

  access_key=<string>
      Explicit TencentCloud access key ID

  secret_key=<string>
      Explicit TencentCloud secret key

  security_token=<string>
      Explicit TencentCloud security token when credential type is AssumeRole

  region=<string>
	  Explicit TencentCloud region

  mount=<string>
      Path where the TencentCloud credential method is mounted. This is usually provided
      via the -path flag in the "vault login" command, but it can be specified
      here as well. If specified here, it takes precedence over the value for
      -path. The default value is "tencentcloud".

  role=<string>
      Name of the role to request a token against
`

	return strings.TrimSpace(help)
}
