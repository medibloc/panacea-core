# Panacea DID Method Specification

## Table of Contents

* [Panacea DID](#panacea-did)
   * [Panacea DID Method Name](#panacea-did-method-name)
   * [Panacea DID Method Specific Identifier](#panacea-did-method-specific-identifier)
* [Panacea DID Document](#panacea-did-document)
* [CRUD Operations (REST)](#crud-operations-rest)
   * [Create](#create)
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
panacea-did = "did:panacea:" network-id ":" idstring ":" checksum
network-id = "mainnet" | "testnet"
idstring = k*HEXDIG
checksum = 4*HEXDIG
```
- `idstring`: The transaction hash of the transaction that was submitted to create the DID on Panacea.
- `checksum`: The first 4 bytes of `SHA3-256(network-id ":" idstring)`


## Panacea DID Document

JSON-LD

```
{
    "@context": "https://www.w3.org/ns/did/v1",
    "id": "did:panacea:mainnet:B9334E0F2032DA0748225438D1A67012CA398E3568B68F40016959D80D3AF5D9:16B66EC9",
    "authentication": [
        "key1"
    ],
    "publicKey": [
        {
            "id": "key1"
            "type": "Ed25519VerificationKey2018",
            "controller": "did:panacea:mainnet:B9334E0F2032DA0748225438D1A67012CA398E3568B68F40016959D80D3AF5D9:16B66EC9",
            "publicKeyBase58": "H3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPV"
        }
    ],
    "service": [
        {
            "id": "service1",
            "type": "VerifiableCredentialService",
            "serviceEndpoint": "https://example.com/vc/"
        }
    ]
}
```

## CRUD Operations (REST)

TODO: Should be written in the Swagger format

### Create

- URL: `/did/`
- Method: `/did/`
- Content Type: `application/json`
- URL Parameters: None
- Request Body: A DID Document
```
{
    "@context": "https://www.w3.org/ns/did/v1",
    "id": "did:panacea:mainnet:B9334E0F2032DA0748225438D1A67012CA398E3568B68F40016959D80D3AF5D9:16B66EC9",
    "authentication": [
        "key1"
    ],
    "publicKey": [
        {
            "id": "key1"
            "type": "Ed25519VerificationKey2018",
            "controller": "did:panacea:mainnet:B9334E0F2032DA0748225438D1A67012CA398E3568B68F40016959D80D3AF5D9:16B66EC9",
            "publicKeyBase58": "H3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPV"
        }
    ]
}
```
- Success Response
    - Code: `200 OK`
    - Body: Unsigned message envelope
    ```
    {
        "mode": "plain",
        "message": {
            "operation": "create",
            "did": "did:panacea:mainnet:B9334E0F2032DA0748225438D1A67012CA398E3568B68F40016959D80D3AF5D9:16B66EC9",
            "didDocumentBase64": "cXdlcm9pam9pamdkamFk",
            "timestamp": "2020-12-25T15:31:42.123Z"
        }
    }
    ```
- Error Response Codes: TO BE DESCRIBED

### Read

- URL: `/did/`
- Method: `GET`
- Content Type: `application/json`
- URL Paraemeters: None
- Request Body:
```
{
    "did": "did:panacea:mainnet:B9334E0F2032DA0748225438D1A67012CA398E3568B68F40016959D80D3AF5D9:16B66EC9"
}
```
- Success Response
    - Code: `200 OK`
    - Body: The latest DID Document in a plain form
    ```
    {
        "@context": "https://www.w3.org/ns/did/v1",
        "id": "did:panacea:mainnet:B9334E0F2032DA0748225438D1A67012CA398E3568B68F40016959D80D3AF5D9:16B66EC9",
        "authentication": [
            "key1"
        ],
        "publicKey": [
            {
                "id": "key1"
                "type": "Ed25519VerificationKey2018",
                "controller": "did:panacea:mainnet:B9334E0F2032DA0748225438D1A67012CA398E3568B68F40016959D80D3AF5D9:16B66EC9",
                "publicKeyBase58": "H3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPV"
            }
        ],
        "service": [
            {
                "id": "service1",
                "type": "VerifiableCredentialService",
                "serviceEndpoint": "https://example.com/vc/"
            }
        ]
    }
    ```
- Error Response Codes: TO BE DESCRIBED

### Update

- URL: `/did/`
- Method: `PUT`
- Content Type: `application/json`
- URL Paraemeters: None
- Request Body: A DID Document
```
{
    "@context": "https://www.w3.org/ns/did/v1",
    "id": "did:panacea:mainnet:B9334E0F2032DA0748225438D1A67012CA398E3568B68F40016959D80D3AF5D9:16B66EC9",
    "authentication": [
        "key1"
    ],
    "publicKey": [
        {
            "id": "key1"
            "type": "Ed25519VerificationKey2018",
            "controller": "did:panacea:mainnet:B9334E0F2032DA0748225438D1A67012CA398E3568B68F40016959D80D3AF5D9:16B66EC9",
            "publicKeyBase58": "H3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPV"
        }
    ]
}
```
- Success Response
    - Code: `200 OK`
    - Body: Unsigned message envelope
    ```
    {
        "mode": "plain",
        "message": {
            "operation": "update",
            "did": "did:panacea:mainnet:B9334E0F2032DA0748225438D1A67012CA398E3568B68F40016959D80D3AF5D9:16B66EC9",
            "didDocumentBase64": "cXdlcm9panFld2xyanFsa3dlO3JqcQ==",
            "timestamp": "2020-12-25T17:31:42.123Z"
        }
    }
    ```
- Error Response Codes: TO BE DESCRIBED

### Delete

- URL: `/did/`
- Method: `DELETE`
- Content Type: `application/json`
- URL Paraemeters: None
- Request Body: A DID Document
```
{
    "@context": "https://www.w3.org/ns/did/v1",
    "id": "did:panacea:mainnet:B9334E0F2032DA0748225438D1A67012CA398E3568B68F40016959D80D3AF5D9:16B66EC9",
    "authentication": [
        "key1"
    ],
    "publicKey": [
        {
            "id": "key1"
            "type": "Ed25519VerificationKey2018",
            "controller": "did:panacea:mainnet:B9334E0F2032DA0748225438D1A67012CA398E3568B68F40016959D80D3AF5D9:16B66EC9",
            "publicKeyBase58": "H3C2AVvLMv6gmMNam3uVAjZpfkcJCwDwnZn6z3wXmqPV"
        }
    ]
}
```
- Success Response
    - Code: `200 OK`
    - Body: Unsigned message envelope
    ```
    {
        "mode": "plain",
        "message": {
            "operation": "delete",
            "did": "did:panacea:mainnet:B9334E0F2032DA0748225438D1A67012CA398E3568B68F40016959D80D3AF5D9:16B66EC9",
            "didDocumentBase64": "cXdlcm9panFld2xyanFsa3dlO3JqcQ==",
            "timestamp": "2020-12-25T18:31:42.123Z"
        }
    }
    ```
- Error Response Codes: TO BE DESCRIBED


## Security Considerations

## Privacy Considerations

## Reference Implementations

## References

- https://w3c-ccg.github.io/did-spec/
