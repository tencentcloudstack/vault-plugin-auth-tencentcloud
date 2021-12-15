
# Tencent Cloud Auth Method

The `tencentcloud` auth method provides an automated mechanism to retrieve
a Vault token for Tencent Cloud entities. Unlike most Vault auth methods, this
method does not require manual first-deploying, or provisioning
security-sensitive credentials (tokens, username/password, client certificates,
etc), by operators. 

## Authentication Workflow

The TencentCloud STS API includes a method,
[`sts:GetCallerIdentity`],
which allows you to validate the identity of a client. The client signs
a `GetCallerIdentity` query using the [Tencent Cloud Signature Algorithm v3](https://intl.cloud.tencent.com/document/api/598/32225). It then
submits 2 pieces of information to the Vault server to recreate a valid signed
request: the request URL, and the request headers. The Vault server then
reconstructs the query and forwards it on to the Tencent Cloud STS service and validates
the result back.

Each signed Tencent Cloud request includes the current timestamp and a nonce to mitigate
the risk of replay attacks.


## Authorization Workflow

The basic mechanism of operation is per-role.

Roles are associated with a role ARN that has been pre-created in Tencent Cloud.
Tencent Cloud's console displays each role's ARN. A role in Vault has a 1:1 relationship
with a role in Tencent Cloud, and must bear the same name.

When a client assumes that role and sends its `GetCallerIdentity` request to Vault,
Vault matches the arn of its assumed role with that of a pre-created role in Vault.
It then checks what policies have been associated with the role, and grants a
token accordingly.

## Authentication

### Via the CLI

#### Enable Tencent Cloud authentication in Vault.

```shell-session
$ vault auth enable tencentcloud
```

#### Configure the credentials required to make TENCENTCLOUD API calls

```shell-session
$ vault write auth/tencentcloud/config/client \
  secret_id="..." \
  secret_key="..."
```

#### Configure the policies on the role.

```shell-session
vault write auth/tencentcloud/role/dev-role arn='qcs::cam::uin/100021543443:roleName/dev-role'
```

#### Perform the login operation

```shell-session
$ vault write auth/tencentcloud/login \
        role=dev-role \
        identity_request_url=$IDENTITY_REQUEST_URL_BASE_64 \
        identity_request_headers=$IDENTITY_REQUEST_HEADERS_BASE_64
```

For the CAM auth method, generating the signed request is a non-standard
operation. The Vault CLI supports generating this for you:

```shell-session
$ vault login -method=tencentcloud secret_id=... secret_key=... token=... region=... role=...
```

This assumes you have the Tencent Cloud credentials you would find on an CVM instance using the
following call:

```
curl 'http://metadata.tencentyun.com/latest/meta-data/cam/security-credentials/$ROLE_NAME'
```

Please note the `$ROLE_NAME` above is case-sensitive and must be consistent with how it's reflected
on the instance.

An example of how to generate the required request values for the `login` method
can be found found in the
Vault CLI source code
## API

The Tencent Cloud auth method has a full HTTP API. Please see the
[Tencent Cloud  Auth API](https://github.com/tencentcloudstack/vault-plugin-auth-tencentcloud/blob/master/docs/Tencent%20Cloud%20-%20Auth%20Methods%20-%20HTTP%20API.md) for more
details.
