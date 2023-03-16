# Data validation

- Status: Draft
- Created: 2023-01-04
- Modified: 2023-03-14
- Authors
  - Gyuguen Jang <gyuguen.jang@medibloc.org>
  - Youngjoon Lee <yjlee@medibloc.org>
  - Hansol Lee <hansol@medibloc.org>
  - Myongsik Gong <myongsik_gong@medibloc.org>
  - Inchul Song <icsong@medibloc.org>
  - Tae Jin Yoon <tj@medibloc.org>


## Synopsis

This document describes the API specification, data verification, and certificate issuance.

### Motivation
Data consumers can define their requirements.
This requirement allows them to define the format of the data they want.
DEP supports `JSON schema` and `Presentation Definition` for data consumers to define these requirements.
Oracle can verify data submitted by providers according to the requirements that data consumers defined.
Oracle issues a certificate upon successful verification, which can be verified by Panacea.

### Definitions
`Data Provider`, `Data Consumer`, and `Oracle` and [JSON schema](https://www.w3.org/2019/wot/json-schema)

## Technical Specification

### API Specification

#### Request URI
```http request
POST /v0/data-deal/deals/{dealID}/data
```

#### Request Headers
```
# Authorization: Bearer {jwtToken}
# Content-Type: application/json
```

TODO: Guide to JWT generate and verify

#### Request Body
```
{
  "provider_address": "panacea1ewugvs354xput6xydl5cd5tvkzcuymkejekwk3",
  "encrypted_data_base64": "3TpdyLXP0xObMWYev8XwqlKjxzWsP5OiQQjq6MooWUwiP37JGqFgl8Rv+a43RMoNieSmBml2AeE1M3sIn39T3R3FD5nqOcFx2MbsnUYHMVASzv5mv53EYx+mP/aPl7pTeMCioQkRqXCyNrj+EJVQEoUdt2DgJwstia3O5pFFRUViKdVJsGIpDX8vY7qQdNuId/beVqWpL5ffSayZnQg=",
  "data_hash": "13341f91f0da76d2adb67b519e2c9759822ceafd193bd26435ba0d5eee4c3a2b"
}
```
| Key                   | Type   | Description                                                      |
|-----------------------|--------|------------------------------------------------------------------|
| provider_address      | string | Data provider's account address                                  |
| encrypted_data_base64 | string | Base64-encoded value after encrypt the original data             |
| data_hash             | string | A hexadecimal string of a SHA256 hash value of the original data |

#### Response Headers
```
# Content-Type: application/json
```

#### Response Body
```json
{
  "unsigned_certificate": {
    "unique_id": "9a3da3162aa592af3c77f1bba2d5635c7b4c065249bd36094fe3c11c73c90618",
    "oracle_address": "panacea1ewugvs354xput6xydl5cd5tvkzcuymkejekwk3",
    "deal_id": 1,
    "provider_address": "panacea1ewugvs354xput6xydl5cd5tvkzcuymkejekwk3",
    "data_hash": "13341f91f0da76d2adb67b519e2c9759822ceafd193bd26435ba0d5eee4c3a2b"
  },
  "signature": "MEQCIEPyGSe9wVIjrUFuzXtQtEc0siwaHkp4QJCMBvC8ttWQAiAUkntIQldgIkIBFdthaTRWXHisDV2Ys/Ufpc9zejUEuQ=="
}
```

| Key                                   | Type   | Description                                                                  |
|---------------------------------------|--------|------------------------------------------------------------------------------|
| unsigned_certificate                  | Object | Unsigned certificate containing data validation information                  |
| unsigned_certificate.unique_id        | string | UniqueID of the oracle that validated the data                               |
| unsigned_certificate.oracle_address   | string | Account address of the oracle that validated the data                        |
| unsigned_certificate.deal_id          | uint   | Deal to whom the provider intends to provide data                            |
| unsigned_certificate.provider_address | string | Data provider's account address                                              |
| unsigned_certificate.data_hash        | string | A hexadecimal string of a SHA256 hash value of the original data             |
| signature                             | string | Base64-encoded string signed `unsigned_certificate` with Oracle private key. |


### Data validation process

#### Data Decryption
Provider's encrypted data(`encrypted_data_base64`) can only be decrypted by oracle.

```
secret_key = SHA256(ECDH(oracle_private_key, provider_public_key))

encrypted_data = Base64.Decode(encrypted_data_base64)

orginal_data = AES256GCM.Decrypt(secret_key, encrypted_data)
```

#### Data Validation
Verify that the original data matches the `data_hash`.
```
compare(data_hash, Hex.Encode(SHA256(orginal_data))
```

Verify that the `provider_address` of the original data matches the JWT auth token issuer of the request header.
```
compare(provider_address, jwtToken.issuer)
```

The deal information can be retrieved from Panacea using the deal ID

Before verifying the data, Oracle should check whether the deal status is valid.
If the Deal's status is invalid, oracle does not perform verification work.
```
compare(deal.status, 'DEAL_STATUS_ACTIVE')
```

Validate original data with `data_schema` or `presentation_definition` specified in the deal.

If the deal has `data_schema` information, then validate json schema as shown below.
```
data_schema = deal.data_schema

JSONSchema.Validate(data_schema, original_data)
```

If the deal has `presentation_definition`, then validate verifiable presentation with presentation definition.
In this case, the `original_data` must be in a verifiable presentation to pass validation.
```
presentation_definition = deal.presentation_definition

ValidateVP(original_data, presentation_definition)
```

#### Data re-encryption and delivery via Consumer Service

If data validation is successful, the data must be re-encrypted and stored on the `consumer service`.

To re-encrypt the data, a symmetric secret key must be used. The symmetric secret key can be derived from the following logic.
```
deal_id_bz = convertUint64ToBigEndian(deal_id)
secret_key = SHA256(append(oracle_private_key, deal_id_bz, data_hash))
```

After encrypting the data with the generated `secretKey`, send it to `consumer service` as HTTP POST request.

```
encrypted_data = AES256GCM.Encrypt(secret_key, orginal_data)
```
```bash
curl -v -X POST -H "Authorization: Bearer ${jwt}" \
  -d "@<encrypted-data-path>" \
  "${consumer-service-endpoint}/v0/deals/${dealId}/data/${dataHash}
```


#### Certificate issuance with a cryptographic signature
If all processes succeed, oracle responds to the Provider by issuing a certificate.

The certificate includes the following contents.
```json
{
  "unsigned_certificate": {
    "unique_id": "9a3da3162aa592af3c77f1bba2d5635c7b4c065249bd36094fe3c11c73c90618",
    "oracle_address": "panacea1ewugvs354xput6xydl5cd5tvkzcuymkejekwk3",
    "deal_id": 1,
    "provider_address": "panacea1ewugvs354xput6xydl5cd5tvkzcuymkejekwk3",
    "data_hash": "13341f91f0da76d2adb67b519e2c9759822ceafd193bd26435ba0d5eee4c3a2b"
  },
  "signature": "MEQCIEPyGSe9wVIjrUFuzXtQtEc0siwaHkp4QJCMBvC8ttWQAiAUkntIQldgIkIBFdthaTRWXHisDV2Ys/Ufpc9zejUEuQ=="
}
```

The certificate requires a signature signed with `oracle_private_key` because Panacea needs to be able to verify that the certificate was generated by a trusted oracle.
The `oracle_public_key` registered in Panacea is a pair with the `oracle_private_key` of a trusted oracle. 
So, Panacea can verify signature of the certificate with `oracle_public_key` to ensure that the certificate is correct.

Signature is a value obtained by signing `unsigned_certificate` with `oracle_private_key`.
```
signature = sign(oracle_private_key, unsigned_certificate)
```


## Backwards Compatibility

Not applicable.

## Forwards Compatibility

If the JSON-LD validation specification is applied in the future, oracle will also be support.

## Example Implementations

TODO: Repo URLs of panacea-core and panacea-oracle with specific tags

## Other Implementations

None at present.

## History

- 2023-01-04: Initial draft finished
- 2023-03-14: add presentation definition validation

## Copyright

All content herein is licensed under [Apache 2.0](https://www.apache.org/licenses/LICENSE-2.0).
