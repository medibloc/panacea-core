# Confidential Oracle

- Status: Draft
- Created: 2023-01-06
- Modified: 2023-01-06
- Authors
    - Inchul Song <icsong@medibloc.org>
    - Youngjoon Lee <yjlee@medibloc.org>
    - Gyuguen Jang <gyuguen.jang@medibloc.org>
    - Hansol Lee <hansol@medibloc.org>
    - Myongsik Gong <myongsik_gong@medibloc.org>

## Synopsis

This document explains that why the oracles must use confidential computing.

## Motivation

In DEP, the data must be verified as explained in the [data validation documentation](./4-data-validation.md).
Since the data is sensitive, we designed the off-chain oracle. Verifying data means that the content of data could be
shown to the oracle.
It also means that the oracle operators(human) can see the content of sensitive data. For example, if the oracle
software stores the important sensitive data to the disk or storage without sealing it,
it can be read by anyone who can access the disk or storage. Plus, if the malicious oracle operator modified the source
code of the oracle to print logs of sensitive data or to leak a contents somewhere else,
it would cause the serious privacy problems.
To prevent it, we have designed the DEP oracle with confidential computing. Confidential computing guarantees that only
selected code (binary) can access data in the secure enclave. Also it could protect a sensitive data from being
leaked by malicious operators while verifying a data correctly in decentralized system.

## Confidential Oracle

For running the DEP oracle with confidential computing, we choose an Intel SGX which offers hardware-based memory
encryption for data security
using [Microsoft Azure Confidential Computing VM](https://learn.microsoft.com/en-us/azure/confidential-computing/overview).
Also, we choose an [EGo](https://www.edgeless.systems/products/ego/) SDK which enables to develop confidential apps
written in Go language.

The data encryption and decryption are performed by the oracle private key. The oracle private key is sealed by unique
ID which is for the Intel SGX enclaves. If the oracle operator modified the source code of the oracle, the unique ID
will be changed.
It means that the all oracle operators must use the only genuine binary built by selected code. The only genuine binary
can unseal the data in the SGX secure enclave.

If the oracle wants to register, it must use remote report composed of the promised security version and unique ID.
So the one of the oracle approves the registration it the new oracle's remote report is valid.
If you want a register the oracle, you can verify remote report using below CLI before submitting a registration.

```bash
ego run oracled verify-report [report-file-path]
```

So how to we set a correct unique ID and how to know it is not malicious binary? The correct unique ID will
be determined by Panacea governance.
We can know the correct unique ID from Panacea as Single Source of Truth(SSOT), and know that what the correct genuine
binary is.

