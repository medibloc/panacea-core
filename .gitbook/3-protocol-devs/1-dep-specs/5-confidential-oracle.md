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
Verifying data means that the content of data could be shown to the oracle. However, no one would want their data to be
exposed to the oracle for data verification. There may also be malicious oracles to exploit the data unfairly. To
prevent it, we have designed the DEP oracle with confidential computing. Confidential computing guarantees that only
selected code (binary) can access data in the secure enclave. Also, it could protect a sensitive data from being leaked 
by malicious operators while verifying a data correctly in decentralized system.

## Confidential Oracle

The oracle operators(human) can see the content of sensitive data. For example, if the oracle software stores the
important sensitive data to the disk or storage without sealing it, it can be read by anyone who can access the disk or
storage. So we developed the confidential oracle for preventing the case from the malicious oracle by using below 
sections.

### Technical Specification

For running the DEP oracle with confidential computing, we choose an Intel SGX which offers hardware-based memory
encryption for data security
using [Microsoft Azure Confidential Computing VM](https://learn.microsoft.com/en-us/azure/confidential-computing/overview).
Also, we choose an [EGo](https://www.edgeless.systems/products/ego/) SDK which enables to develop confidential apps
written in Go language.

### Oracle Key

An oracle key is an asymmetric key that should be shared across all oracles. The request from provider and consumer that
oracle receive must be encrypted by the oracle public key. If the secure data encryption and decryption need to be
performed, the corresponding oracle private must be held by every oracle securely. How the all oracle shared the oracle
private key and how they stored the key securely will be described in next section.

### Oracle Key Handshake

The oracle private key is sealed by unique ID which is for the Intel SGX enclaves. If the oracle operator modified the
source code of the oracle, the unique ID will be changed. If the malicious oracle operator modified the source code of
the oracle to print logs of sensitive data or to leak a contents somewhere else, it would cause the serious privacy
problems. It means that the all oracle operators must use the only genuine binary built by selected code.
The only genuine binary can unseal the data in the SGX secure enclave.

If the oracle wants to register, it must use remote report composed of the promised security version and unique ID.
So the one of the oracle approves the registration if the new oracle's remote report is valid.
If you want a register the oracle, you can verify remote report that is valid or not using
[CLI](../../5-oracles/7-verfiy-remote-report.md). After the oracle registration passed, the new oracle can retrieve the shared oracle
private key.

### Sealing Secrets and States

All secrets and important information, such as the shared oracle private key and the state of blockchain light client,
must be sealed before it is stored to the disk, so that anyone outside the enclave cannot read it. The sealing must be
done with the unique ID of the enclave, so that any modified enclave cannot unseal secrets.

So how to we set a correct unique ID and how to know it is not malicious binary? The correct unique ID will be
determined by Panacea governance. We can know the correct unique ID from Panacea as Single Source of Truth(SSOT),
and know that what the correct genuine binary is.
