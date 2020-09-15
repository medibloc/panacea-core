# Panacea DID Method Specification


## Abstract

Panacea is a public blockchain built by MediBloc to reinvent the healthcare experience.
Panacea also supports DID operations. DIDs are created and stored in the Panacea, and they are used with verifiable credentials.

This specification describes how DIDs are managed on the Panacea.


## Table of Contents

* [DID Method Name](#did-method-name)
* [DID Method Specific Identifier](#did-method-specific-identifier)
    * [Relationship between DIDs and Panacea accounts](#relationship-between-dids-and-panacea-accounts)
* [DID Document Format (JSON-LD)](#did-document-format-json-ld)
* [CRUD Operations](#crud-operations)
    * [Create (Register)](#create-register)
    * [Read](#read)
    * [Update](#update)
    * [Deactivate](#deactivate)
* [Security Considerations](#security-considerations)
    * [Replay Attack](#replay-attack)
* [Privacy Considerations](#privacy-considerations)
* [Reference Implementations](#reference-implementations)
* [References](#references)


## DID Method Name

The namestring is `panacea`.

A DID must begin with the prefix: `did:panacea` in lowercase.


## DID Method Specific Identifier

```
panacea-did = "did:panacea:" network-id ":" idstring
network-id = "mainnet" / "testnet"
idstring = 21*22(base58)
base58 = "1" / "2" / "3" / "4" / "5" / "6" / "7" / "8" / "9" / "A" / "B" /
         "C" / "D" / "E" / "F" / "G" / "H" / "J" / "K" / "L" / "M" / "N" /
         "P" / "Q" / "R" / "S" / "T" / "U" / "V" / "W" / "X" / "Y" / "Z" /
         "a" / "b" / "c" / "d" / "e" / "f" / "g" / "h" / "i" / "j" / "k" /
         "m" / "n" / "o" / "p" / "q" / "r" / "s" / "t" / "u" / "v" / "w" /
         "x" / "y" / "z"
```
To generate `idstring`, Panacea SDK generates a Secp256k1 key-pair, and
encodes the first 16 bytes of its public key into base58.
This gives a length of either 21 or 22 characters, and it means that DIDs are case-sensitive,
even though the prefix is always lower-case.

The Panacea SDK provides a tool for generateing the Secp256k1 key-pair either randomly or from a mnemonic provided by the user.

Example:
```
did:panacea:mainnet:DnreD8QqXAQaEW9DwC16Wh
```

### Relationship between DIDs and Panacea accounts

DIDs are independent of any Panacea accounts.
Panacea accounts are necessary only for sending transactions to Panacea
to create/update/deactivate the DIDs.

It means that Panacea accounts are not used to verify the DID ownership.
To prove the DID ownership, users must include a signature to the transaction.
The signature must be generated with the private key which corresponds to one of the public keys registered in the DID document.
The signature is different from the Panacea transaction signature generated with the private key of the Panacea account. 
The details are described below.


## DID Document Format (JSON-LD)

```json
{
    "@context": "https://www.w3.org/ns/did/v1",
    "id": "did:panacea:mainnet:G3UzSnRRsyApppuHVuaff",
    "verificationMethod": [
        {
            "id": "did:panacea:mainnet:G3UzSnRRsyApppuHVuaff#key1",
            "type": "Secp256k1VerificationKey2018",
            "controller": "did:panacea:mainnet:G3UzSnRRsyApppuHVuaff",
            "publicKeyBase58": "dBuN4i7dqwCLzSX7GHBLsfUoXw5RmWQ3DwQ9Ee4bfh5Y"
        }
    ],
    "authentication": [
        "did:panacea:mainnet:G3UzSnRRsyApppuHVuaff#key1"
    ]
}
```

Currently, the `controller` in the `verificationMethod` must be equal to the [DID subject](https://www.w3.org/TR/2020/WD-did-core-20200907/#dfn-did-subjects).
It would be extended later.

The Key IDs in the `authentication` are references to one of public keys specified in the `verificationMethod`.
The spec of the `authentication` would be extended in the future.

The Panacea DID Document doesn't contain the `service` field currently. It would be extended soon.


## CRUD Operations

### Create (Register)

To create a DID Document in Panacea, the following transaction should be submitted.
```json
{
    "type": "did/MsgCreateDID",
    "value": {
        "did": "did:panacea:mainnet:G3UzSnRRsyApppuHVuaff",
        "document": {
            "@context": "https://www.w3.org/ns/did/v1",
            "id": "did:panacea:mainnet:G3UzSnRRsyApppuHVuaff",
            "verificationMethod": [
                {
                    "id": "did:panacea:mainnet:G3UzSnRRsyApppuHVuaff#key1",
                    "type": "Secp256k1VerificationKey2018",
                    "controller": "did:panacea:mainnet:G3UzSnRRsyApppuHVuaff",
                    "publicKeyBase58": "dBuN4i7dqwCLzSX7GHBLsfUoXw5RmWQ3DwQ9Ee4bfh5Y"
                }
            ],
            "authentication": [
                "did:panacea:mainnet:G3UzSnRRsyApppuHVuaff#key1"
            ]
        },
        "signature": "FLOgUBcMEjKs/o1lgu4Y5Ump/2xee0D0tLsrY9+YVMUD/G/qbSHo3lOJ4Jv2zsDn1grcbIYSQsOvoBTbYXXg3g==",
        "sig_key_id": "did:panacea:mainnet:G3UzSnRRsyApppuHVuaff#key1",
        "from_address": "panacea1d58s72gu0mjkw0lkgyvr0eqzz3mv74awfsjslz"
    }
}
```
The transaction must have a `did` and a `document` which will be stored in the Panacea.

It also must have a `signature` and a `sig_key_id` for proving the ownership of the DID.
The `signature` must be generated from the `document` and the sequence `"0"`.
It must be signed with a private key which corresponds to the public key referred by the `sig_key_id`.
The `sig_key_id` must be one of the key IDs specified in the `authentication` of the `document`.

The source of the `signature` should look like (encoded with Amino):
```json
{
    "data": {
        "@context": ...,
        "id": "did:panacea:...",
        ...
    },
    "sequence": "0"
}
```

The transaction also must contain a `from_address` which is a Panacea account.
Also, it must be signed with the private key of the Panacea account, so that Panacea can verify the transaction.

The transaction fails if the same DID exists or if it has been already deactivated.

### Read

A Panacea DID Document can be looked up by the following query.
```json
{
    "did": "did:panacea:mainnet:G3UzSnRRsyApppuHVuaff"
}
```

If the DID exists (not deactivated yet), the result is:
```json
{
    "document": {
        "@context": "https://www.w3.org/ns/did/v1",
        "id": "did:panacea:mainnet:G3UzSnRRsyApppuHVuaff",
        "verificationMethod": [
            {
                "id": "did:panacea:mainnet:G3UzSnRRsyApppuHVuaff#key1",
                "type": "Secp256k1VerificationKey2018",
                "controller": "did:panacea:mainnet:G3UzSnRRsyApppuHVuaff",
                "publicKeyBase58": "dBuN4i7dqwCLzSX7GHBLsfUoXw5RmWQ3DwQ9Ee4bfh5Y"
            }
        ],
        "authentication": [
            "did:panacea:mainnet:G3UzSnRRsyApppuHVuaff#key1"
        ]
    },
    "sequence": "0"
}
```

The `sequence` is returned along with the `document`.
It must be included in the subsequent transaction (update/deactivate) for preventing transaction replay attacks.

### Update

Only the DID owner can replace the DID Document using the following transaction.

This example is for adding a new public key to the `verificationMethod` and adding a dedicated public key to the `authentication`.
```json
{
    "type": "did/MsgUpdateDID",
    "value": {
        "did": "did:panacea:mainnet:G3UzSnRRsyApppuHVuaff",
        "document": {
            "@context": "https://www.w3.org/ns/did/v1",
            "id": "did:panacea:mainnet:G3UzSnRRsyApppuHVuaff",
            "verificationMethod": [
                {
                    "id": "did:panacea:mainnet:G3UzSnRRsyApppuHVuaff#key1",
                    "type": "Secp256k1VerificationKey2018",
                    "controller": "did:panacea:mainnet:G3UzSnRRsyApppuHVuaff",
                    "publicKeyBase58": "dBuN4i7dqwCLzSX7GHBLsfUoXw5RmWQ3DwQ9Ee4bfh5Y"
                },
                {
                    "id": "did:panacea:mainnet:G3UzSnRRsyApppuHVuaff#key2",
                    "type": "Secp256k1VerificationKey2018",
                    "controller": "did:panacea:mainnet:G3UzSnRRsyApppuHVuaff",
                    "publicKeyBase58": "2BjcxuwijyE1om4991ANiFrwZJ3Ev5YYX9KiPKgaHmGsi"
                }
            ],
            "authentication": [
                "did:panacea:mainnet:G3UzSnRRsyApppuHVuaff#key1",
                {
                    "id": "did:panacea:mainnet:G3UzSnRRsyApppuHVuaff#key3",
                    "type": "Secp256k1VerificationKey2018",
                    "controller": "did:panacea:mainnet:G3UzSnRRsyApppuHVuaff",
                    "publicKeyBase58": "yE1om4991ANiFrwZJ3Ev5YYX9KiPKgaHmGsi2Bjcxuwij"
                }
            ]
        },
        "signature": "xtsQH3D5naHe9IXmhCnohlChwHiD0dx9PI4aPkaJPGoEznYMHmg0aBerg85ai7T2WNxxlc39uFzAxKbI4sbJCA==",
        "sig_key_id": "did:panacea:mainnet:G3UzSnRRsyApppuHVuaff#key1",
        "from_address": "panacea1d58s72gu0mjkw0lkgyvr0eqzz3mv74awfsjslz"
    }
}
```

Like creating DIDs, The `signature` must be generated from the `document` and the `sequence` returned from the Read DID operation.
It must be signed with a private key which corresponds to the public key referred by the `sig_key_id`.
The `sig_key_id` must be one of the key IDs specified in the `authentication` of the `document`.

Whenever submitting this transaction, the user must query the current `sequence` by the Read DID operation.
(The user can also increment the `sequence` manually, but the transaction can be rejected if there are the concurrent transactions with the same `sequence`.)

The source of the `signature` should look like (encoded with Amino):
```json
{
    "data": {
        "@context": ...,
        "id": "did:panacea:...",
        ...
    },
    "sequence": "50"
}
```

The transaction fails if the DID has been already deactivated.

### Deactivate

To deactivate the DID document, the DID owner should send the following transaction.

Panacea doesn't delete the DID document. The document is just deactivated.
This strategy guarantees that malicious users cannot recreate the DID,
because the DID deactivation may be appropriate when a person dies or a business is terminated.
```json
{
    "type": "did/MsgDeactivateDID",
    "value": {
        "did": "did:panacea:mainnet:DnreD8QqXAQaEW9DwC16Wh",
        "signature": "xtsQH3D5naHe9IXmhCnohlChwHiD0dx9PI4aPkaJPGoEznYMHmg0aBerg85ai7T2WNxxlc39uFzAxKbI4sbJCA==",
        "sig_key_id": "did:panacea:mainnet:G3UzSnRRsyApppuHVuaff#key1",
        "from_address": "panacea1d58s72gu0mjkw0lkgyvr0eqzz3mv74awfsjslz"
    }
}
```

The `signature` must be generated from the `did` and the `sequence` returned from the Read DID operation.
It must be signed with a private key which corresponds to the public key referred by the `sig_key_id`.
The `sig_key_id` must be one of the key IDs specified in the `authentication` of the `document`.

The source of the `signature` should look like (encoded with Amino):
```json
{
    "data": "did:panacea:...",
    "sequence": "50"
}
```

The transaction fails if the DID doesn't exist or if it has been already deactivated.

## Security Considerations

### Replay Attack

To prove the DID ownership, Create/Update/Deactivate transactions must contain a `signature` and a `sig_key_id`.
If malicous users can replay the transaction with the same `signature`, the DID document can be modified unexpectedly.

To prevent replay attacks, a `sequence` must be included when generating the `signature`.
The `sequence` can be obtained by the Read DID operation.

The `sequence` is monotonically incremented by the Panacea when the transaction is committed.
That is, malicious users cannot reuse the signature from the previous transaction committed.
The user must generate a new signature from the new `sequence`.

## Privacy Considerations

## Reference Implementations

- Core: https://github.com/medibloc/panacea-core
- SDK: https://github.com/medibloc/panacea-js

## References

- https://w3c-ccg.github.io/did-spec/
