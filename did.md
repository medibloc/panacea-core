# Panacea DID Method Specification

## Table of Contents

* [Panacea DID](#panacea-did)
   * [Panacea DID Method Name](#panacea-did-method-name)
   * [Panacea DID Method Specific Identifier](#panacea-did-method-specific-identifier)
* [Panacea DID Document](#panacea-did-document)
* [CRUD Operations](#crud-operations)
   * [Create (Register)](#create-register)
   * [Read](#read)
   * [Update](#update)
   * [Delete](#delete)
* [Security Considerations](#security-considerations)
* [Privacy Considerations](#privacy-considerations)
* [Reference Implementations](#reference-implementations)
* [References](#references)

## Panacea DID

### Panacea DID Method Name

`panacea`

### Panacea DID Method Specific Identifier

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
The `idstring` is a base58-encoded of the first 16 bytes of a 256bit Ed25519 verification key (the public portion of the key pair).
This gives an length of either 21 or 22 characters, and it means that DIDs are case-sensitive,
even though the prefix is always lower-case.

Example:
```
did:panacea:mainnet:DnreD8QqXAQaEW9DwC16Wh
```


## Panacea DID Document

JSON-LD

```
{
    "@context": "https://www.w3.org/ns/did/v1",
    "id": "did:panacea:mainnet:DnreD8QqXAQaEW9DwC16Wh",
    "authentication": [
        "key1"
    ],
    "publicKey": [
        {
            "id": "did:panacea:mainnet:DnreD8QqXAQaEW9DwC16Wh#key1",
            "type": "Ed25519VerificationKey2018",
            "controller": "did:panacea:mainnet:DnreD8QqXAQaEW9DwC16Wh",
            "publicKeyBase58": "H3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPV"
        }
    ],
    "service": [
        {
            "id": "did:panacea:mainnet:DnreD8QqXAQaEW9DwC16Wh#svc1",
            "type": "VerifiableCredentialService",
            "serviceEndpoint": "https://example.com/vc/"
        }
    ]
}
```

## CRUD Operations

### Create (Register)

To create a DID Document in Panacea, the following transaction should be submitted.
```
{
    "type": "create",
    "did": <new DID that is being registered>,
    "document": {
        "publicKey": [{
            "id": <a valid unique identifier>,
            "type": "Ed25519VerificationKey2018",
            "publicKeyBase58": <base58-encoded public key of a Ed25519 verification key-pair>,
        }],
        "authentication": [{
            "publicKey": [<reference to a publicKey object>]
        }],
        "service": [{
            "type": <VerifiableCredentialService, ...>,
            "serviceEndpoint": <A URI for the endpoint>
        }]
    }
}
```

Example:
```
{
    "type": "create",
    "did": "did:panacea:mainnet:DnreD8QqXAQaEW9DwC16Wh",
    "document": {
        "publicKey": [{
            "id": "did:panacea:mainnet:DnreD8QqXAQaEW9DwC16Wh#key1",
            "type": "Ed25519VerificationKey2018",
            "publicKeyBase58": "H3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPV"
        }],
        "authentication": [{
            "publicKey": ["did:panacea:mainnet:DnreD8QqXAQaEW9DwC16Wh#key1"]
        }],
        "service": [{
            "type": "VerifiableCredentialService",
            "serviceEndpoint": "https://example.com/vc/"
        }]
    }
}
```

Possible outcomes include:
- TBD

### Read

A Panacea DID Document can be looked up by the DID using the following transaction.
```
{
    "type": "read",
    "did": <DID to be queried>
}
```

Example:
```
{
    "type": "read",
    "did": "did:panacea:mainnet:DnreD8QqXAQaEW9DwC16Wh"
}
```

### Update

Only the DID owner can replace the DID Document using the following transaction.
```
{
    "type": "update",
    "did": <DID whose document needs to be updated>,
    "document": {
        "publicKey": [{
            "id": <a valid unique identifier>,
            "type": "Ed25519VerificationKey2018",
            "publicKeyBase58": <base58-encoded public key of a Ed25519 verification key-pair>,
        }, {
            "id": <a valid unique identifier>,
            "type": "Ed25519VerificationKey2018",
            "publicKeyBase58": <base58-encoded public key of a Ed25519 verification key-pair>,
        }],
        "authentication": [{
            "publicKey": [<references to publicKey objects>]
        }],
        "service": [{
            "type": <VerifiableCredentialService, ...>,
            "serviceEndpoint": <A URI for the endpoint>
        }]
    }
}
```

Example:
```
{
    "type": "update",
    "did": "did:panacea:mainnet:DnreD8QqXAQaEW9DwC16Wh",
    "document": {
        "publicKey": [{
            "id": "did:panacea:mainnet:DnreD8QqXAQaEW9DwC16Wh#key1",
            "type": "Ed25519VerificationKey2018",
            "publicKeyBase58": "H3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPV"
        }, {
            "id": "did:panacea:mainnet:DnreD8QqXAQaEW9DwC16Wh#key2",
            "type": "Ed25519VerificationKey2018",
            "publicKeyBase58": "VAjZpfkcJCwDwnZn6z3wXmqPVH3C2AVvLMv6gmMNam3u"
        }],
        "authentication": [{
            "publicKey": [
                "did:panacea:mainnet:DnreD8QqXAQaEW9DwC16Wh#key1",
                "did:panacea:mainnet:DnreD8QqXAQaEW9DwC16Wh#key2"
            ]
        }],
        "service": [{
            "type": "VerifiableCredentialService",
            "serviceEndpoint": "https://example.com/vc/"
        }]
    }
}
```

### Delete

[TODO: This paragraph is temporary. Should be decided how to implement this operation.]
Internally, the delete operation is processed by setting the DID Document to null.

To delete the DID document, the DID owner should send the following transaction.
```
{
    "type": "delete",
    "did": <DID whose document is being deleted>
}
```

Example:
```
{
    "type": "delete",
    "did": "did:panacea:mainnet:DnreD8QqXAQaEW9DwC16Wh"
}
```

## Security Considerations

## Privacy Considerations

## Reference Implementations

## References

- https://w3c-ccg.github.io/did-spec/
