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

Licensed under the Apache License, Version 2.0.
