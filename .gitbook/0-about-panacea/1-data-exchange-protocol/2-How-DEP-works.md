# How DEP Works

In this guide, you can see how we solved the problem described on [DEP and the problems it solves](./1-DEP-problems-it-solves.md);

- [For data privacy](#for-data-privacy)
- [For data validation](#for-data-validation)
- [Atomic incentive](#atomic-incentive-distribution)
- [Decentralization & Scalability](#decentralization--scalability)

## DEP flow

Before reading on, if you want to know the entire flow of a DEP, check out this [user-flow](../../3-protocol-devs/1-dep-specs/1-user-flow.md)

## For data privacy

### Data privacy between provider & oracle

In the process of [Data Validation](../../3-protocol-devs/1-dep-specs/4-data-validation.md), data provider first sends data to oracle for data verification and receives a certificate.
In this process, the provider encrypts the data with the public key of the oracle and its own private key.
The oracle private key is sealed in the enclave environment, so the data cannot be decrypted by anyone except the oracle.

### Data privacy in oracle

In the process of [Data Validation](../../3-protocol-devs/1-dep-specs/4-data-validation.md), the oracle decrypts the encrypted data with oracle private key and obtains the original data.
However, the oracle we defined must run in an enclave environment, and the enclave memory cannot be read or written outside the enclave.
Therefore, it is impossible to leak the original data and data privacy is continuously protected.

### Data privacy between provider & consumer

After completing the [Data Validation](../../3-protocol-devs/1-dep-specs/4-data-validation.md), oracle encrypts the data and provides this encrypted to the consumer.
At this time, a symmetric key based on the oracle private key is generated, so the data cannot be decrypted until the consumer also receives this symmetric key.
After the [Submit Consent](../../3-protocol-devs/1-dep-specs/3-data-provider-consent.md) and [Incentive Distribution](../../3-protocol-devs/1-dep-specs/6-incentives.md) process is completed, the consumer can request the symmetric key from the oracle.
The oracle checks whether the incentive distribution process is completed from the panacea blockchain, and provides the secret key encrypted based on the consumer's public key.
As a result, it is impossible for the provider to obtain the original data before the incentive is paid, and the secret key cannot be obtained except by the consumer.

## For data validation

Oracle validates the data after receiving encrypted data from the provider.
After the basic deal status, data hash validation, oracle checks whether the data satisfies the requirements specified in the deal.
Here are two types of data requirements that can be specified in the deal (json schema validation, presentation definition validation).

[Json schema](https://json-schema.org) validation checks whether the data satisfies the conditions for the schema definition.
[Presentation definition](https://identity.foundation/presentation-exchange/#presentation-definition) validation checks that the data is in the form of a did-based verifiable presentation and satisfies the presentation definition.
This validation can additionally verify that the data was created from a verifiable credential issued by a Certificate Authority, and thus can verify that the data is authenticated by the Certificate Authority.

## Atomic incentive distribution

After submitting consent to Panacea, the provider can get the incentive through [Incentive Distribution](../../3-protocol-devs/1-dep-specs/6-incentives.md) process.
In this process, panacea blockchain checks the consent information and immediately completes the incentive payment to the provider, and at the same time enables the consumer to obtain the secret key to decrypt the provided data.
Thus, the provider is guaranteed to be rewarded for providing the data, and the consumer is guaranteed to access the data after completing the payment.

## Decentralization & Scalability

The DEP components, panacea blockchain and oracle, are open decentralized ecosystem where everyone can participate.
Panacea blockchain is a decentralized ecosystem where anyone can participate as a full node as a public blockchain, and tx is confirmed through the consensus of validators. 
In addition, oracle can also participate through the oracle registration process of the panacea blockchain.

Panacea blockchain validators can get block generation rewards by participating in the consensus, and oracle can get commissions as incentives are paid out.
This reward system can make a healthy decentralized ecosystem.

Providers or consumers can be guaranteed to get the same response no matter which oracle they request.
Therefore, consumers and providers can choose one of several registered oracles to perform [Data Validation](../../3-protocol-devs/1-dep-specs/4-data-validation.md).
This DEP structure that allows the validation process to be divided into the oracles.
As a result, data validation in oracle can be processed in parallel by multiple oracles and the overall performance of DEP improves as the number of oracles increases.

## Further Reading

- For more detailed specification of DEP, see [DEP specifications](../../3-protocol-devs/1-dep-specs/0-overview.md).
- For more detailed information of data privacy, see [Data Validation](../../3-protocol-devs/1-dep-specs/4-data-validation.md).
- For more detailed information of deal requirements, see [Data Deal](../../3-protocol-devs/1-dep-specs/2-data-deal.md).
- For more information about Oracle, see the [Operate Oracle Nodes](../../5-oracles/1-operate-oracle-nodes/0-overview.md).
