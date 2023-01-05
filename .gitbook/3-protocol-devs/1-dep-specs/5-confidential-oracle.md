# Confidential Oracle

- Status: Draft
- Created: 2023-01-05
- Modified: 2023-01-05
- Authors
    - Youngjoon Lee <yjlee@medibloc.org>
    - Gyuguen Jang <gyuguen.jang@medibloc.org>
    - Hansol Lee <hansol@medibloc.org>
    - Myongsik Gong <myongsik_gong@medibloc.org>
    - Inchul Song <icsong@medibloc.org>

## Synopsis
This document explains that why the oracles must use confidential computing.

## Privacy for Data Exchange
In DEP, anyone can operate oracle node. However, when validating the sensitive data through the general-purpose oracle, it could be occurred several vulnerabilities. 
The one of vulnerabilities is that the sensitive data could be exposed to oracle operator by using malicious binary. The other vulnerabilities is that the data could be stolen from external node.  
So preventing from these vulnerabilities, the oracle must use confidential computing by using Intel SGX Technology in TEE(Trusted Execution Environment).

## Confidential Oracle 
In DEP, anyone can operate oracle nodes but must use confidential computing.
If the operator runs the confidential oracle, the only oracle software can decrypt a data and validate it. It means that the oracle operator can't decrypt a data and it will be not exposed to the operator.
Also, when running the confidential oracle, the only genuine binary should be used so that it can be prevented from the node who uses malicious binary.
Using a genuine binary not only preserves a privacy of data, but also unseal the data decryption key.

### Additional Attempt

- Encryption/Decryption based on unique ID 
- Using Panacea as Single Source of Truth

## Implementation of Confidential Oracle

- Intel SGX with Azure Confidential Computing blah blah...
- EGo blah blah blah..
