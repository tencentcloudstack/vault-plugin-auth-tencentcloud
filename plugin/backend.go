package plugin

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/hashicorp/vault-plugin-auth-tencentcloud/sdk"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	debug := conf.Logger.IsDebug()

	if !debug {
		env := strings.ToLower(os.Getenv("VAULT_LOG_LEVEL"))
		debug = env == "trace" || env == "debug"
	}

	b := newBackend(&sdk.LogRoundTripper{
		Debug: debug,
	})

	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}

	return b, nil
}

func newBackend(transport http.RoundTripper) *backend {
	b := &backend{
		transport: transport,
	}

	b.Backend = &framework.Backend{
		AuthRenew: b.pathLoginRenew,
		Help:      backendHelp,
		PathsSpecial: &logical.Paths{
			Unauthenticated: []string{
				"login",
			},
		},
		Paths: []*framework.Path{
			pathLogin(b),
			pathListRoles(b),
			pathRole(b),
		},
		BackendType: logical.TypeCredential,
	}

	return b
}

type backend struct {
	*framework.Backend

	transport http.RoundTripper
}

const backendHelp = `
That TencentCloud CAM auth method allows entities to authenticate based on their identity and pre-configured roles.
`
