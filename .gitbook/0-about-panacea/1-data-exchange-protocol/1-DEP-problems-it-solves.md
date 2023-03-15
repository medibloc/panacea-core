# DEP and the problems it solves

DEP seeks to address the problems of decentralized data delivery that allows for data validation while ensuring data privacy.

DEP aims to stay true to the five principles below:
- [Data privacy](#data-privacy)
- [Data validation](#data-validation)
- [Incentive](#incentive)
- [Decentralization](#decentralization)
- [Scalability](#scalability)

## Data privacy

Since DEP is targeting sensitive data, the privacy of the data must be ensured throughout all data transfers. 
No data must be exposed to anyone except promised data consumers during the data delivery.

DEP uses data encryption in the whole data delivery process to ensure the privacy of sensitive data.
[ECIES hybrid encryption](https://cryptobook.nakov.com/asymmetric-key-ciphers/ecies-public-key-encryption) is used to ensure data privacy during the entire data delivery process.

## Data validation

As data consumers, it is crucial to validate that the data they receive is correct.
The data must be truthful, untampered, and conform to the requirements the consumer specified.

DEP provides a way to verify data by utilizing oracle running in an secure enclave environment and DID-based verifiable presentation.

## Incentive

By providing data, the data provider can be guaranteed an incentive paid by the data consumer. 
In this case, data delivery and payment must be processed atomically and simultaneously.

DEP ensures this through Panacea blockchain transactions and data encryption.

## Decentralization

DEP aims to be a decentralized protocol with no centralized mediator for data validation and delivery.

Anyone can participate in operating DEP components, such as Panacea blockchain and oracle. By the nature of the proof-of-stake blockchain, any party with less than 2/3 voting power cannot change the behavior of DEP, and anyone with less than 1/3 voting power cannot stop the protocol. Also, anyone can operate oracle nodes to improve the performance of DEP and earn incentives if they have appropriate hardware environments.
The entire source code, protocol specifications, and discussions are open-sourced, so anyone can dive into details and participate in developments. Any protocol changes must go through governments to be deployed in production. 

## Scalability

The DEP aims to have a scalable structure for oracle, which plays a key role in data validation.
Since oracle plays a role in data validation, if the oracle structure is not scalable, it can be fatal to the overall DEP performance.

The DEP has designed the ecosystem in such a way that as the number of oracles increases, data validation request can be divided into multiple oracles.
This ensures that as the number of oracles increases, the overall performance of the DEP increases.

## Further Reading

- To learn more about how DEP works, see [How DEP works](2-How-DEP-works.md).
- For more detailed specification of DEP, see [DEP specifications](../../3-protocol-devs/1-dep-specs/0-overview.md).
- For more information about Oracle, see the [Operate Oracle Nodes](../../5-oracles/1-operate-oracle-nodes/0-overview.md).
