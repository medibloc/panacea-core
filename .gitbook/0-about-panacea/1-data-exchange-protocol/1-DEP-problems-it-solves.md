# DEP and the problems it solves

DEP seeks to address the problems of decentralized data delivery that allows for data validation while ensuring data privacy:

- [Data privacy](#data-privacy)
- [Data validation](#data-validation)
- [Incentive](#incentive)
- [Decentralization](#decentralization)
- [Scalability](#scalability)

## Data privacy

Since DEP is targeting sensitive data, the privacy of the data must be ensured throughout all data transfers. 
It is essential that no data is exposed in the process of data delivery.

DEP uses data encryption in all data delivery processes to ensure the privacy of sensitive data. 
Symmetric and asymmetric key encryption is utilized to ensure that the privacy of data during all data delivery processes.

## Data validation

As a data consumer, it is important to validate that the data you receive is the right data.
The data must be truthful, uncorrupted, and conform to the requirements the consumer propose.

DEP provides a way to verify data by utilizing oracle running in an enclave environment, did-based verifiable presentation.

## Incentive

By providing data, the data provider can be guaranteed with an incentive offered by the data consumer. 
In this case, the delivery of data and the paying incentive must be processed atomically and simultaneously. 

DEP ensures this through the panacea blockchain Tx and data encryption.

## Decentralization

DEP aims to be a decentralized protocol with no centralized data validator and deliverer.

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
