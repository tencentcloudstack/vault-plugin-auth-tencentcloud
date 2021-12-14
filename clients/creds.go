package clients

import (
	"fmt"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
)

// client chainedCreds for Cli
func ChainedCredsToCli(secretId, secretKey, token string) (common.CredentialIface, error) {
	providerChain := []common.Provider{
		NewConfigurationCredentialProvider(&Configuration{secretId, secretKey, token}),
		DefaultEnvProvider(),
		common.DefaultCvmRoleProvider(),
	}
	return common.NewProviderChain(providerChain).GetCredential()
}

// Configuration
type Configuration struct {
	SecretId  string
	SecretKey string
	Token     string
}

// NewConfigurationCredentialProvider
func NewConfigurationCredentialProvider(configuration *Configuration) common.Provider {
	return &ConfigurationProvider{
		Configuration: configuration,
	}
}

// ConfigurationProvider
type ConfigurationProvider struct {
	Configuration *Configuration
}

// GetCredential
func (p *ConfigurationProvider) GetCredential() (common.CredentialIface, error) {
	if p.Configuration.SecretId != "" && p.Configuration.SecretKey != "" {
		return common.NewTokenCredential(p.Configuration.SecretId, p.Configuration.SecretKey, p.Configuration.Token), nil
	} else {
		return nil, ErrNoValidCredentialsFound
	}
}

var (
	ErrNoValidCredentialsFound = fmt.Errorf("no valid credentials were found")
)
