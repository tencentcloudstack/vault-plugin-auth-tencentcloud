# Tencent Cloud Auth Method

The `tencentcloud` auth method provides an automated mechanism to retrieve a Vault token for Tencent Cloud entities.
Unlike most Vault auth methods, this method does not require manual first-deploying, or provisioning security-sensitive
credentials (tokens, username/password, client certificates, etc), by operators.

## Authentication Workflow

The TencentCloud STS API includes a method,
[`sts:GetCallerIdentity`], which allows you to validate the identity of a client. The client signs a `GetCallerIdentity`
query using the [Tencent Cloud Signature Algorithm v3](https://intl.cloud.tencent.com/document/api/598/32225). It then
submits 2 pieces of information to the Vault server to recreate a valid signed request: the request URL, and the request
headers. The Vault server then reconstructs the query and forwards it on to the Tencent Cloud STS service and validates
the result back.

Each signed Tencent Cloud request includes the current timestamp and a nonce to mitigate the risk of replay attacks.

## Authorization Workflow

The basic mechanism of operation is per-role.

Roles are associated with a role ARN that has been pre-created in Tencent Cloud. Tencent Cloud's console displays each
role's ARN. A role in Vault has a 1:1 relationship with a role in Tencent Cloud, and must bear the same name.

When a client assumes that role and sends its `GetCallerIdentity` request to Vault, Vault matches the arn of its assumed
role with that of a pre-created role in Vault. It then checks what policies have been associated with the role, and
grants a token accordingly.

## Authentication

### Via the CLI

### From Sources

If you prefer to build the plugin from sources, clone the GitHub repository locally.

### Build the plugin

Build the auth method into a plugin using Go.

```shell
$ go build -o vault/plugins/vault-plugin-auth-tencentcloud ./cmd/vault-plugin-auth-tencentcloud/main.go
```

### Configuration

Copy the plugin binary into a location of your choice; this directory must be specified as
the [`plugin_directory`](https://www.vaultproject.io/docs/configuration#plugin_directory) in the Vault configuration
file:

```hcl
plugin_directory = "vault/plugins"
```

Start a Vault server with this configuration file:

```sh
$ vault server -config=vault/server.hcl
```

Once the server is started, register the plugin in the Vault
server's [plugin catalog](https://www.vaultproject.io/docs/internals/plugins#plugin-catalog):

```sh
$ SHA256=$(shasum -a 256 vault/plugins/vault-plugin-auth-tencentcloud | cut -d ' ' -f1)
$ vault plugin register -sha256=$SHA256 auth vault-plugin-auth-tencentcloud
$ vault plugin info auth vault-plugin-auth-tencentcloud
```

You can now enable the tencentCloud auth plugin:

```sh
$ vault auth enable -path=tencentcloud vault-plugin-auth-tencentcloud
Success! Enabled vault-plugin-auth-tencentcloud auth method at: tencentcloud/
```

#### Configure the credentials required to make TencentCLoud API calls

```shell
$ vault write auth/tencentcloud/config/client \
  secret_id="..." \
  secret_key="..."
```

#### Configure the policies on the role.

```shell
$ vault write auth/tencentcloud/role/dev-role arn='qcs::cam::uin/100021543443:roleName/dev-role'
```

#### Perform the login operation

```shell
$ vault write auth/tencentcloud/login \
        role=dev-role \
        region=$IDENTITY_REQUEST_REGION \
        secret_id=$IDENTITY_REQUEST_SECRET_ID \
        secret_key=$IDENTITY_REQUEST_SECRET_KEY \
        token=$IDENTITY_REQUEST_TOKEN
```

For the CAM auth method, generating the signed request is a non-standard operation. The Vault CLI supports generating
this for you:

```shell
$ vault login -method=tencentcloud secret_id=... secret_key=... token=... region=... role=...
```

This assumes you have the Tencent Cloud credentials you would find on an CVM instance using the following call:

```shell
curl 'http://metadata.tencentyun.com/latest/meta-data/cam/security-credentials/$ROLE_NAME'
```

Please note the `$ROLE_NAME` above is case-sensitive and must be consistent with how it's reflected on the instance.

An example of how to generate the required request values for the `login` method can be found found in the Vault CLI
source code

## API

The Tencent Cloud auth method has a full HTTP API. Please see the
[Tencent Cloud  Auth API](https://github.com/tencentcloudstack/vault-plugin-auth-tencentcloud/blob/master/docs/Tencent%20Cloud%20-%20Auth%20Methods%20-%20HTTP%20API.md)
for more details.
