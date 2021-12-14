# Tencent Cloud Auth Method (API)

This is the API documentation for the Vault Tencent Cloud auth method. For general information about the usage and
operation of the Tencent Cloud method, please see
the [Vault Tencent Cloud auth method documentation](/docs/auth/tencentcloud).

This documentation assumes the Tencent Cloud auth method is mounted at the `/auth/tencentcloud`
path in Vault. Since it is possible to enable auth methods at any location, please update your API calls accordingly.

## Create Role

Registers a role. Only entities using the role registered using this endpoint will be able to perform the login
operation.

| Method | Path                        |
| :----- | :-------------------------- |
| `POST` | `/auth/tencentcloud/role/:role` |

### Parameters

- `role` `(string: <required>)` - Name of the role. Must correspond with the name of the role reflected in the arn.
- `arn` `(string: <required>)` - The role's arn.

- `token_ttl` `(integer: 0 or string: "")` - The incremental lifetime for generated tokens. This current value of this
  will be referenced at renewal time.
- `token_max_ttl` `(integer: 0 or string: "")` - The maximum lifetime for generated tokens. This current value of this
  will be referenced at renewal time.
- `token_policies` `(array: [] or comma-delimited string: "")` - List of policies to encode onto generated tokens.
  Depending on the auth method, this list may be supplemented by user/group/other values.
- `token_bound_cidrs` `(array: [] or comma-delimited string: "")` - List of CIDR blocks; if set, specifies blocks of IP
  addresses which can authenticate successfully, and ties the resulting token to these blocks as well.
- `token_explicit_max_ttl` `(integer: 0 or string: "")` - If set, will encode an explicit max TTL onto the token. This
  is a hard cap even if token_ttl and token_max_ttl would otherwise allow a renewal.
- `token_no_default_policy` `(bool: false)` - If set, the default policy will not be set on generated tokens; otherwise
  it will be added to the policies set in token_policies.
- `token_num_uses` `(integer: 0) - The maximum number of times a generated token may be used (within its lifetime); 0
  means unlimited. If you require the token to have the ability to create child tokens, you will need to set this value
  to 0.
- `token_period` `(integer: 0 or string: "")` - The period, if any, to set on the token.
- `token_type` `(string: "")` - The type of token that should be generated. Can be service, batch, or default to use the
  mount's tuned default (which unless changed will be service tokens). For token store roles, there are two additional
  possibilities: default-service and default-batch which specify the type to return unless the client requests a
  different type at generation time.

### Sample Payload

```json
{
  "arn": "qcs::cam::uin/100021543888:roleName/hastrustedactors",
  "policies": [
    "dev",
    "prod"
  ]
}
```

### Sample Request

```shell-session
$ curl \
    --header "X-Vault-Token: ..." \
    --request POST \
    --data @payload.json \
    http://127.0.0.1:8200/v1/auth/tencentcloud/role/dev-role
```

## Read Role

Returns the previously registered role configuration.

| Method | Path                        |
| :----- | :-------------------------- |
| `GET`  | `/auth/tencentcloud/role/:role` |

### Parameters

- `role` `(string: <required>)` - Name of the role.

### Sample Request

```shell-session
$ curl \
    --header "X-Vault-Token: ..." \
    http://127.0.0.1:8200/v1/auth/tencentcloud/role/dev-role
```

### Sample Response

```json
{
  "data": {
    "qcs::cam::uin/100021543888:roleName/hastrustedactors",
    "policies": [
      "default",
      "dev",
      "prod"
    ],
    "ttl": 3600000,
    "max_ttl": 3600000,
    "period": 0
  }
}
```

## List Roles

Lists all the roles that are registered with the method.

| Method | Path                   |
| :----- | :--------------------- |
| `LIST` | `/auth/tencentcloud/roles` |

### Sample Request

```shell-session
$ curl \
    --header "X-Vault-Token: ..." \
    --request LIST \
    http://127.0.0.1:8200/v1/auth/tencentcloud/roles
```

### Sample Response

```json
{
  "data": {
    "keys": [
      "dev-role",
      "prod-role"
    ]
  }
}
```

## Delete Role

Deletes the previously registered role.

| Method   | Path                        |
| :------- | :-------------------------- |
| `DELETE` | `/auth/tencentcloud/role/:role` |

### Parameters

- `role` `(string: <required>)` - Name of the role.

### Sample Request

```shell-session
$ curl \
    --header "X-Vault-Token: ..." \
    --request DELETE \
    http://127.0.0.1:8200/v1/auth/tencentcloud/role/dev-role
```

## Login

Fetch a token. This endpoint verifies the signature of the signed GetCallerIdentity request.

| Method | Path                   |
| :----- | :--------------------- |
| `POST` | `/auth/tencentcloud/login` |

### Parameters

- `role` `(string: <required>)` - Name of the role.
- `identity_request_url` `(string: <required>)` - Base64-encoded HTTP URL used in the signed request.
- `identity_request_headers` `(string: <required>)` - Base64-encoded, JSON-serialized representation of the sts:
  GetCallerIdentity HTTP request headers. The JSON serialization assumes that each header key maps to either a string
  value or an array of string values (though the length of that array will probably only be one).

### Sample Payload

```json
{
  "role": "dev-role",
  "identity_request_url": "...",
  "identity_request_headers": "..."
}
```

### Sample Request

```shell-session
$ curl \
    --request POST \
    --data @payload.json \
    http://127.0.0.1:8200/v1/auth/tencentcloud/login
```

### Sample Response

```json
{
  "request_id": "34a3d3f7-5ab4-2673-e9a7-bdfb6235e5b5",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": null,
  "wrap_info": null,
  "warnings": null,
  "auth": {
    "client_token": "s.mQXkXZTtaXbgJ6swYkKEqdaxasdasda",
    "accessor": "OABKIM1ktcwTxyw9Wm0UZgpm",
    "policies": [
      "default,dev"
    ],
    "token_policies": [
      "default,dev"
    ],
    "metadata": {
      "account_id": "1252588437728950",
      "arn": "qcs:sts::1252588437728950:assumed-role/3123123123761253761",
      "identity_type": "CAMRole",
      "principal_id": "1252588437728950",
      "request_id": "AB13042E-EB70-591A-AEA4-8B744CA1531C",
      "role_id": "dev-role",
      "role_name": "dev-role",
      "user_id": "3123123123761253761/root-dev-role12312323213-9423"
    },
    "lease_duration": 2764800,
    "renewable": true,
    "entity_id": "0aaaf804-e123-67d3-2104-8e95eb896de8",
    "token_type": "service",
    "orphan": true
  }
}
```
