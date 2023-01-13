# Running Node

## Overview

If you have completed the previous steps, now is the step to actually run oracle.

The oracle is responsible for validating the provider's data and ensuring that the data is transmitted securely using the `oracle key`.
In this process, oracle operators can earn commissions according to the commission rate registered to Panacea.
The oracle also serves to validate and approve `registration/upgrade` requests from other oracles.

You can see the details in [What oracle does](#what-oracle-does).

## Prerequisites
- [Hardware Requirement](5-oracles/1-operate-oracle-nodes/1-oracle-installation.md)
- Complete [Oracle Registration](5-oracles/1-operate-oracle-nodes/4-oracle-registration.md) or [Genesis Oracle](5-oracles/1-operate-oracle-nodes/3-genesis-oracle.md)
    - Your oracle have to be registered in Panacea
    - The `oracle_priv_key.sealed` must be in the oracle home path

## Start oracle

You can start oracle with following CLI:
```bash
docker run \
    --device /dev/sgx_enclave \
    --device /dev/sgx_provision \
    -v ${ANY_DIR_ON_HOST}:/oracle \
    ghcr.io/medibloc/panacea-oracle:latest \
    ego run /usr/bin/oracled start
```
If oracle start is successful, you can see the following log message like example below:
```
EGo v1.1.0 (4625a610928f4f4b1ea49262c363376b1e574b6c)
[erthost] loading enclave ...
[erthost] entering enclave ...
[ego] starting application ...
time="2023-01-11T07:40:31Z" level=info msg="successfully connect to IPFS node"
time="2023-01-11T07:40:31Z" level=info msg="dialing to Panacea gRPC endpoint: http://127.0.0.1:9090"
time="2023-01-11T07:40:31Z" level=info msg="Panacea event subscriber is started"
time="2023-01-11T07:40:31Z" level=info msg="subscribe RegisterOracleEvent. query: message.action = 'RegisterOracle'"
time="2023-01-11T07:40:31Z" level=info msg="subscribe UpgradeOracleEvent. query: message.action = 'UpgradeOracle'"
time="2023-01-11T07:40:31Z" level=info msg="HTTP server is started: 127.0.0.1:8081"
```

## What oracle does

### Validate provider data & issue certificate

Oracle provides a REST API to validate data from a provider and issue a certificate.

Validation procedure is as follows:
1. Check the address of the requested provider with JWT
2. Check the status of related `deal`
3. Decrypt provider's encrypted data & check data hash and data schema

If validation passes successfully, oracle issues a certificate as follows:
1. Generate `secret key` by combining `oracle private key`, deal ID, and data hash
2. Re-encrypt the data using the `secret key` and put it in `IPFS`
3. Issue a certificate with `oracle private key` signature

Providers will be able to submit consent to Panacea through the issued certificate.
Panacea ensures that a commission is paid to the oracle operator who issued the certificate when the data delivery is successfully completed.

### Safely transmit provider data accessibility to consumer

Oracle provides a REST API to get data accessibility to consumer.

After the `submit-consent` transaction succeeds in Panacea, oracle transmits secret key that enables data access to the consumer.
The detailed process is as follows:
1. Check the address of the requested consumer with JWT
2. Check if the requested consumer is the owner of the `deal` 
3. Check if the `submit-consent` transaction succeeds 
4. Make `encrypted secret key` using `consumer public key`
5. Response with `encrypted secret key` to consumer

Consumer can obtain the `secret key` through his/her `private key`, and can decrypt data from `IPFS`.

### Validate and approve registration/upgrade requests of other oracles

Oracle subscribes `registration/upgrade` events from Panacea.

When another oracle sends a `registration/upgrade`, the running oracle do some verifications by checking if:
- correct version of oracle binary is used
- the oracle is running inside an enclave
- the `Node Key` is generated inside the enclave
- the trusted block information is valid
- the oracle is registered (at upgrade requests)

When the `registration/upgrade` is verified successfully, oracle send a transaction for approval of the oracle `registration/upgrade`.