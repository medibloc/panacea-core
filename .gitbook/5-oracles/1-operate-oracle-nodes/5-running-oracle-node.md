# Running an Oracle Node

## Overview

If you have completed previous steps, now is the time to actually run oracle.

The oracle is responsible for validating the data provider's data and 
ensuring that the data is transmitted to the data consumer securely using the `oracle key`.

In this process, oracle operators can earn commissions according to the commission rate registered to Panacea.
The oracle also serves to validate and approve `registration/upgrade` requests from other oracles.

You can see the details in [What oracle does](#what-oracle-does).

## Prerequisites
- [Hardware Requirement](5-oracles/1-operate-oracle-nodes/1-oracle-installation.md)
- Complete [Oracle Registration](5-oracles/1-operate-oracle-nodes/4-oracle-registration.md) or [Genesis Oracle](5-oracles/1-operate-oracle-nodes/3-genesis-oracle.md)
    - Your oracle have to be registered in Panacea
    - The `oracle_priv_key.sealed` must be in the oracle home path

## Start oracle

You can start an oracle with the following CLI:
```bash
docker run \
    --device /dev/sgx_enclave \
    --device /dev/sgx_provision \
    -v ${ANY_DIR_ON_HOST}:/oracle \
    ghcr.io/medibloc/panacea-oracle:latest \
    ego run /usr/bin/oracled start
```
If the oracle is successful started, you will see the following log message:
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

Oracle provides a REST API to validate data provided by a data provider and issue a certificate.

The validation procedure is as follows:
1. Check the address of the data provider who is requesting data validation with JWT
2. Check the status of the related `deal`
3. Decrypt the data provider's encrypted data & check the data hash and the data schema

If the validation passes successfully, the oracle issues a certificate as follows:
1. Generate a `secret key` by combining `oracle private key`, deal ID, and data hash
2. Re-encrypt the data using the `secret key` and put it in `IPFS`
3. Issue a certificate with `oracle private key` signature

Data providers will be able to submit a consent to Panacea with the issued certificate.
Panacea ensures that a commission is paid to the oracle operator who issued the certificate when the data delivery is successfully completed.

### Safely transmit the data accessibility to the data consumer

Oracle provides a REST API to get data accessibility to the data consumer.

After the `submit-consent` transaction succeeds in Panacea, the oracle transmits a secret key that enables data access to the data consumer.
The detailed process is as follows:
1. Check the address of the data consumer who is requesting data access with JWT
2. Check if the requesting consumer is the owner of the `deal` 
3. Check if the `submit-consent` transaction is present and successful 
4. Make `encrypted secret key` using `consumer public key`
5. Respond with `encrypted secret key` to the data consumer

The data consumer can obtain the `secret key` through his/her `private key`, and can decrypt data from `IPFS`.

### Validate and approve registration/upgrade requests of other oracles

The oracle subscribes to `registration/upgrade` events from Panacea.

When another oracle sends a `registration/upgrade`, the running oracle do verifications by checking if:
- a correct version of the oracle binary is used
- the oracle is running inside an enclave
- the `Node Key` is generated inside the enclave
- the trusted block information is valid
- the oracle is registered (at upgrade requests)

When the `registration/upgrade` is verified successfully, the oracle sends a transaction for approval of the oracle `registration/upgrade`.
