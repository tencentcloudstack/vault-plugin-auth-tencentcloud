# Vault Plugin: TencentCloud Auth Backend

This is a standalone backend plugin for use with [Hashicorp Vault](https://www.github.com/hashicorp/vault).
This plugin allows authentication to Vault using Cloud Access Management (CAM).

**Please note**: We take Vault's security and our users' trust very seriously. If you believe you have found a security issue in Vault, _please responsibly disclose_ by contacting us at [security@hashicorp.com](mailto:security@hashicorp.com).

## Quick Links
    - Vault Website: https://www.vaultproject.io
    - TencentCloud Auth Docs: https://www.vaultproject.io/docs/auth/tencentcloud.html
    - Main Project Github: https://www.github.com/hashicorp/vault

## Getting Started

This is a [Vault plugin](https://www.vaultproject.io/docs/internals/plugins.html)
and is meant to work with Vault. This guide assumes you have already installed Vault
and have a basic understanding of how Vault works.

Otherwise, first read this guide on how to [get started with Vault](https://www.vaultproject.io/intro/getting-started/install.html).

To learn specifically about how plugins work, see documentation on [Vault plugins](https://www.vaultproject.io/docs/internals/plugins.html).

## Security Model

This authentication model places Vault in the middle of a call between a client and
TencentCloud's "GetCallerIdentity" method. Based on TencentCloud's response, it grants
an access token based on pre-configured roles.

## Usage

Please see [documentation for the plugin](https://www.vaultproject.io/docs/auth/tencentcloud.html)
on the Vault website.

This plugin is currently built into Vault and by default is accessed at `auth/tencentcloud`.
To enable this in a running Vault server:

```sh
$ vault auth enable tencentcloud
Success! Enabled tencentcloud auth method at: tencentcloud/
```

To see all the supported paths, see the [TencentCloud auth backend docs](https://www.vaultproject.io/docs/auth/tencentcloud.html).

## Developing

If you wish to work on this plugin, you'll first need [Go](https://www.golang.org) installed on
your machine (version 1.13+ is required).

For local dev first make sure Go is properly installed, including setting up a [GOPATH](https://golang.org/doc/code.html#GOPATH).
Next, clone this repository into `$GOPATH/src/github.com/hashicorp/vault-plugin-auth-tencentcloud`.
You can then download any required build tools by bootstrapping your environment:

```sh
$ make bootstrap
```

To compile a development version of this plugin, run `make` or `make dev`. This will put the
plugin binary in the `bin` and `$GOPATH/bin` folders. `dev` mode will only generate the binary
for your platform and is faster:

```sh
$ make
$ make dev
```

For local development, use Vault's "dev" mode for fast setup:

```sh
$ vault server -dev -dev-plugin-dir="path/to/plugin/directory"
```

The plugin will automatically be added to the catalog with the name "vault-plugin-auth-tencentcloud".
Run the following command to enable this new auth method as a plugin:

```sh
$ vault auth enable -path=tencentcloud vault-plugin-auth-tencentcloud
Success! Enabled vault-plugin-auth-tencentcloud auth method at: tencentcloud/
```

#### Tests

This plugin has comprehensive [acceptance tests](https://en.wikipedia.org/wiki/Acceptance_testing)
covering most of the features of this auth backend.

If you are developing this plugin and want to verify it is still
functioning (and you haven't broken anything else), we recommend
running the acceptance tests.

**Warning:** The acceptance tests create/destroy/modify *real resources*,
which may incur real costs in some cases. In the presence of a bug,
it is technically possible that broken backends could leave dangling
data behind. Therefore, please run the acceptance tests at your own risk.
At the very least, we recommend running them in their own private
account for whatever backend you're testing.

To run the acceptance tests, you will need a TencentCloud account.

To run the acceptance tests, invoke `make test-acc`:

```sh
$ export VAULT_ACC_TEST_SECRET_ID=YOU 
$ export VAULT_ACC_TEST_SECRET_KEY=YOU SECRET KEY
$ export VAULT_ACC_TEST_TOKEN=YOU SECRET TOKEN (if you run as a CAM role, VAULT_ACC_TEST_TOKEN is required)
$ export CLIENT_CONFIG_TEST_SECRET_ID=CLIENT CONFIG SECRET ID
$ export CLIENT_CONFIG_TEST_SECRET_KEY=CLIENT CONFIG SECRET KEY
$ make test-acc
```

You can also specify a `TESTARGS` variable to filter tests like so:

```sh
$ make test-acc TESTARGS='--run=TestConfig'
```

To run the integration tests, invoke `make test`:

```sh
$ make test
```

You can also specify a `TESTARGS` variable to filter tests like so:

```sh
$ make test TESTARGS='--run=TestConfig'
```