# Vault Plugin: Bitcoin Secrets Backend

This is a backend secrets plugin to be used with Hashicorp Vault. This plugin manages secret keys for the Bitcoin blockchain platform.

## Usage

Assuming you have Hashicorp Vault installed, `scripts/dev_start.sh` is a helper script to start up Vault in dev mode and mount this plugin.
Vault will be listening on localhost:8200.

Once the plugin is mounted, you can start writing secrets to it.

### Log In To Vault

```
export VAULT_ADDR=http://localhost:8200
vault login
```

The token is "root" if you've used dev_start.sh to start Vault.

### Creating an Account

`vault write bitcoin/accounts/MyAccountName`

This will create a new account called "MyAccountName". 

### Viewing an Account

`vault read bitcoin/accounts/MyAccountName`

### Viewing All Account Names

`vault list bitcoin/accounts`

### Creating a Signed Payment Transaction

`vault write bitcoin/payments source=MySourceAccountName destination=MyDestinationAccountName amount=35 unsignedTx=01000...`

This will return a signed transaction.

## Running Tests

```
make test
```

Running tests with coverage:
```
make coverage
```

## License

Copyright (c) 2019 ChainFront LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at:

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
