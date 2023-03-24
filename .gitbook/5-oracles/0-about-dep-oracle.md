# About DEP Oracle

This document provides a brief description of oracle's role in the `DEP` ecosystem and a conceptual introduction to why we use the `SGX` environment.
There is also a brief description of oracle reward for active oracle participation in the `DEP` ecosystem.

## Role of Oracle

Oracle is a key component of `data validation` in the `DEP` ecosystem.

### Data Validator

Oracle is a validator that validates data as it transmits from provider to consumer.

Since data encryption is required for data privacy, `Panacea`, a public blockchain, cannot act as a validator.
Instead, the oracle decrypts and validates the provider's data and provides the result in the form of a `certificate`.
Then, data consumer and `Panacea blockchain` can check the `certificate` to determine the validity of the data.

For a detailed process, see [Data Validation](../3-protocol-devs/1-dep-specs/4-data-validation.md#data-validation).


### Mediator for Data Delivery

Oracle also works as a data deliverer.
The oracle is responsible for delivering the verified data to the appropriate consumer.
And, to prevent malicious consumers who just want to get the data without paying the incentive, the oracle is responsible for managing the data decryption key (`secret key`).
The consumer can get `secret key` through the oracle after the provider get incentive.

For a detailed process, see [Data Validation](../3-protocol-devs/1-dep-specs/4-data-validation.md#data-re-encryption-and-delivery-via-consumer-service).

## Oracle based on SGX

The oracle only works on [SGX](https://www.intel.com/content/www/us/en/developer/tools/software-guard-extensions/overview.html) environment. 

In order for the oracle to work as a validator/deliverer, the data it decrypts inside the oracle must be inaccessible to anyone, including the oracle operator.
Also, the private key it uses must also be inaccessible to anyone. 
To achieve this, it was necessary to make the oracle run in `Intel SGX` environment that provides an `enclave` environment.

Below is more information about the operation of oracle in the `SGX` environment:

### Decrypting and validating data in enclaves

`Enclave` is a trusted execution environment, which means that `enclave` memory cannot be read or written from outside. 
This ensures data privacy because even the oracle operator cannot get the original data while the oracle is decrypting and verifying data.

### Storing oracle private key with sealing

In DEP, the oracle private key is important key that should not be leaked.
The encrypted data that the provider sends to the oracle during the `data validation` process, can be decrypted using `oracle private key`.
Also, the signature that is included in the certificate is also generated based on the `oracle private key`.
The `secret key` that allows the consumer to access the data is also created based on the `oracle private key`.

As shown above, the `oracle private key` is information that should not be revealed to anyone, including the oracle operator, because it is the key to ensuring data privacy and making `DEP` work.
In the `SGX` environment, oracle can store the oracle private key in a completely private way using a feature called `sealing`.
Sealed data can only be decrypted within the `enclave`, and even the oracle operator cannot obtain the decrypted data.
Therefore, oracle stores the `oracle private key `in sealed form and only uses it within the `enclave` to prevent the risk of leakage.

### Blocking malicious oracle

In the DEP ecosystem, anyone can run the oracle and participate in the ecosystem.
Running a proper oracle prevents the oracle operator from obtaining the original data or the `oracle private key`, but there may be users who want to use a modified oracle to steal sensitive data.
Using the `unique ID` and `remote report` provided by `SGX`, the DEP ecosystem is designed to block these potentially malicious oracles.

In order for an oracle to act normally, it must go through the `oracle registration` process through Panacea, which checks the `remote report` and `unique ID` of the oracle.
The `unique ID` allows us to verify that the oracle is an oracle running on the valid binary, and the `remote report` allows us to verify that it is an oracle running on the `SGX` environment.
A modified oracle will not be able to pass this part and will not be able to participate in the `DEP` ecosystem.

For more information on this part, see [verify remote report](./1-operate-oracle-nodes/8-verify-remote-report.md), [oracle registration](./1-operate-oracle-nodes/4-oracle-registration.md).

## Oracle Incentive

In the `DEP` ecosystem, oracle is an essential element, and the performance of verification increases as the number of operated oracles increases.
Therefore, the `DEP` ecosystem aims to be an ecosystem where many oracles participate.
However, running an oracle requires considerable operating costs because it must be run in a special environment called `SGX`.
Therefore, oracle incentive is needed to encourage active oracle participation in the `DEP` ecosystem.

Currently, it is designed so that when a provider submits a `certificate` from an oracle, the oracle information is put inside, and when the provider is rewarded, the oracle get rewards with a certain commission rate.
This may be further developed in the future, see [here](../3-protocol-devs/1-dep-specs/7-incentives.md) for more details.

## Further Reading
- If you want to join the ecosystem as an oracle operator, see [Operate Oracle](./1-operate-oracle-nodes/0-overview.md)
- If you want a detailed specification of the DEP, see [DEP Specs](../3-protocol-devs/1-dep-specs/0-overview.md).