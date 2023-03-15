# What is DEP?

## Overview

DEP is a protocol for delivering sensitive data in a decentralized way while considering the privacy of the data.
In order to deliver sensitive data (e.g., medical data) in a decentralized environment, 
data encryption is essential, and as a trade-off, it makes difficult to verify whether the data is valid or not.
DEP goes beyond these limitations and aims for a protocol that allows sensitive data to be delivered securely and validation of the data.

With DEP, data consumers can specify the type of data they want in a deal, 
and data providers can be guaranteed that their data is delivered securely with data privacy. 
With DEP, data providers can take ownership of their data and receive the proper compensation,
while data consumers can gain access to sensitive data that would otherwise be difficult to collect 
and utilize for various research or business purposes.

## DEP Component
The DEP ecosystem consists of consumer, provider, Panacea blockchain, and oracle as components.

### Data Consumer
A data consumer is an individual or an organization who wants to consume certain kinds of data for specific purposes, with or without paying.
The consumer creates a deal that specifies the terms of the data they want to get and reward the provider who delivers the data they want.

### Data Provider
A data provider is an individual or an organization that holds data and a permission to provide data to data consumers to obtain benefits, such as incentives or services.

### Panacea Blockchain Validator
A Panacea blockchain validator is a blockchain node operator that participates in the blockchain consensus process to guarantee the integrity of the whole process of DEP, which includes consuming data, providing data, validating data, and executing payments.

### Oracle
An oracle is a data validator that guarantees validity and integrity of the data before data is delivered from data providers to data consumers.
Data verification is essential to ensure the atomicity of data delivery and incentive payments.

## Further Reading

- For an overview of the problems that DEP solves, see [DEP and the problems it solves](1-DEP-problems-it-solves.md) and [How DEP works](2-How-DEP-works.md).
- For more detailed specification of DEP, see [DEP specifications](../../3-protocol-devs/1-dep-specs/0-overview.md).
- For more information about Oracle, see the [Operate Oracle Nodes](../../5-oracles/1-operate-oracle-nodes/0-overview.md).