package vault_plugin_auth_tencentcloud

import (
	"context"
	"fmt"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	configClientStoragePath = "config/client"
	secretId                = "secret_id"
	secretKey               = "secret_key"
)

type clientConfig struct {
	SecretId  string `json:"secret_id"`
	SecretKey string `json:"secret_key"`
}

func pathConfigClient(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: configClientStoragePath,
		Fields: map[string]*framework.FieldSchema{
			secretId: {
				Type:        framework.TypeString,
				Description: "Secret Id for the account used to make TencentCloud API requests.",
			},
			secretKey: {
				Type:        framework.TypeString,
				Description: "Secret Key for the account used to make TencentCloud API requests.",
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.CreateOperation: &framework.PathOperation{
				Callback: b.pathConfigWrite,
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.pathConfigWrite,
			},
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.pathConfigRead,
			},
			logical.DeleteOperation: &framework.PathOperation{
				Callback: b.pathConfigDelete,
			},
		},
		ExistenceCheck:  b.pathConfigExistenceCheck,
		HelpSynopsis:    pathConfigRootHelpSyn,
		HelpDescription: pathConfigRootHelpDesc,
	}
}

// pathConfigWrite
func (b *backend) pathConfigWrite(ctx context.Context,
	req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	creds, err := readCredConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if creds == nil {
		if req.Operation == logical.UpdateOperation {
			return nil, fmt.Errorf("config not found during update operation")
		}
		creds = new(clientConfig)
	}

	if secretIdIfc, ok := data.GetOk(secretId); ok {
		creds.SecretId = secretIdIfc.(string)
	}
	if secretKeyIfc, ok := data.GetOk(secretKey); ok {
		creds.SecretKey = secretKeyIfc.(string)
	}
	err = writeCredConfig(ctx, creds, req.Storage)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// pathConfigRead
func (b *backend) pathConfigRead(ctx context.Context,
	req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	creds, err := readCredConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if creds == nil {
		return nil, nil
	}
	return &logical.Response{
		Data: map[string]interface{}{
			secretId:  creds.SecretId,
			secretKey: creds.SecretKey,
		},
	}, nil
}

// pathConfigDelete
func (b *backend) pathConfigDelete(ctx context.Context,
	req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	if err := req.Storage.Delete(ctx, configClientStoragePath); err != nil {
		return nil, err
	}
	return nil, nil
}

// pathConfigExistenceCheck
func (b *backend) pathConfigExistenceCheck(ctx context.Context,
	req *logical.Request, data *framework.FieldData) (bool, error) {
	config, err := readCredConfig(ctx, req.Storage)
	if err != nil {
		return false, err
	}

	return config != nil, nil
}

func readCredConfig(ctx context.Context, storage logical.Storage) (*clientConfig, error) {
	entry, err := storage.Get(ctx, configClientStoragePath)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}
	creds := &clientConfig{}
	if err = entry.DecodeJSON(creds); err != nil {
		return nil, err
	}
	return creds, nil
}

func writeCredConfig(ctx context.Context, config *clientConfig, s logical.Storage) error {
	entry, err := logical.StorageEntryJSON(configClientStoragePath, config)

	if err != nil {
		return err
	}

	err = s.Put(ctx, entry)
	if err != nil {
		return err
	}
	return nil
}

const (
	pathConfigRootHelpSyn = `
    Configure the secret id and key for the account used to make TencentCloud API requests.
    `
	pathConfigRootHelpDesc = `
    Before doing anything, the TencentCloud backend needs credentials that are able
    to call STS AssumeRole. This endpoint is used to configure those credentials.
    `
)
