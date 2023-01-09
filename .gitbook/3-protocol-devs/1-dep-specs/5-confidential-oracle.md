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

## Privacy Vulnerabilities of Oracle
In DEP, anyone can operate the oracle node. However, when verifying the sensitive data through the general-purpose oracle, it could be occurred several privacy vulnerabilities.
The one of vulnerabilities is that the sensitive data could be exposed to the oracle operator who uses malicious binary.
The other vulnerabilities is that the data could be stolen from external node or attack.
So preventing from these vulnerabilities, the oracle must use confidential computing which is operated in [TEE(Trusted Execution Environment)](https://en.wikipedia.org/wiki/Trusted_execution_environment) using Intel SGX.

## Confidential Oracle
In DEP, anyone can operate oracle nodes but must use confidential computing.
If the operator runs the confidential oracle, the only oracle software can decrypt a data and verified it. It means that the oracle operator can't decrypt a data and it will be not exposed to the operator.
Also, when running the confidential oracle, the only genuine binary built in Intel SGX should be used so that it can be prevented from the node who uses malicious binary. It means that all the oracles run the same binary.
As a result, using a genuine binary can preserve a privacy of data.

### Additional Efforts

Using a confidential oracle could reinforce a security, but the oracle was developed with additional efforts for protecting a data leak.
When running an oracle with the genuine binary, the data encryption and decryption are performed by oracle private key which is sealed by unique ID located in Intel SGX.
So if the binary is changed or corrupted, the data encryption and decryption would not work.

So how to we set a correct unique ID and how to know it is not malicious binary? The answer is that the unique ID will be determined by Panacea governance.
We can know the correct unique ID from Panacea as Single Source of Truth(SSOT), and know that what the correct genuine binary is.
Because of the reason, the oracle operator who wants to steal data could not use malicious binary and it could be prevented from potential data leakage.


## Implementation of Confidential Oracle

- Intel SGX with Microsoft Azure Confidential Computing

The confidential oracle operator must run the nodes in TEE that uses Intel SGX technology with Microsoft Azure Confidential Computing Virtual Machine(VM).
You can refer to the details of Azure Confidential Computing in this [link](https://learn.microsoft.com/en-us/azure/confidential-computing/overview).

- EGo

EGo is an open-source SDK that enable to develop confidential apps written in Go language. The oracle in DEP used a EGo so that operator needs to install EGo modules.
You can refer the source code and details of EGo following in this [link](https://github.com/edgelesssys/ego).