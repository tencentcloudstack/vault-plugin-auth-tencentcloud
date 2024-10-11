package vault_plugin_auth_tencentcloud

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/vault-plugin-auth-tencentcloud/tools"
	"github.com/hashicorp/vault/api"
)

// vault cli handler
type CLIHandler struct{}

// auth
func (h *CLIHandler) Auth(c *api.Client, m map[string]string) (*api.Secret, error) {
	mount, ok := m["mount"]
	if !ok {
		mount = "tencentcloud"
	}
	role := m["role"]
	sid := m["secret_id"]
	skey := m["secret_key"]
	token := m["token"]
	region := m["region"]
	loginData := tools.GenerateLoginDataV2(role, sid, skey, token, region)
	path := fmt.Sprintf("auth/%s/login", mount)
	secret, err := c.Logical().Write(path, loginData)
	if err != nil {
		return nil, err
	}
	if secret == nil {
		return nil, errors.New("empty response from credential provider")
	}
	return secret, nil
}

// help
func (h *CLIHandler) Help() string {
	help := `
Usage: vault login -method=tencentcloud [CONFIG K=V...]

  The TencentCloud auth method allows users to authenticate with TencentCloud CAM
  credentials.

  The TencentCloud CAM credentials may be specified explicitly via the command line:

      $ vault login -method=tencentcloud secret_id=... secret_key=... token=... region=...

Configuration:

  secret_id=<string>
      Explicit TencentCloud secret id

  secret_key=<string>
      Explicit TencentCloud secret key

  token=<string>
      Explicit TencentCloud token

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
