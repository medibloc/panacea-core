# User Flow

- Status: Draft
- Created: 2023-01-04
- Modified: 2023-01-04
- Authors
    - Youngjoon Lee <yjlee@medibloc.org>
    - Gyuguen Jang <gyuguen.jang@medibloc.org>
    - Hansol Lee <hansol@medibloc.org>
    - Myongsik Gong <myongsik_gong@medibloc.org>
    - Inchul Song <icsong@medibloc.org>


## Synopsis

This document describes the API specification and verification process for data verification in Oracle.

### Motivation

Oracle can verify whether the data provided by the provider is in the form the consumer wants.
The Consent created after successful verification contains the oracle information that verified the data.

### Definitions
`Data Provider`, `Data Consumer`, and `Oracle` and [JSON Schema](https://www.w3.org/2019/wot/json-schema)

## Technical Specification

TODO: blahblah with diagrams (e.g. seq diagrams)

### API Sepc

#### Request URI
```http request
POST /deals/{dealID}/data
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
| data_hash             | string | The value of hashing the original data with the SHA256 algorithm |

#### Response Headers
```
# Content-Type: application/json
```

#### ResponseBody
```json
{
  "unsigned_certificate": {
    "cid": "QmeSiVLXWagUv9sLEHvbsUJy8rm7r5BoP2pAXrbx4pdbWi",
    "unique_id": "9a3da3162aa592af3c77f1bba2d5635c7b4c065249bd36094fe3c11c73c90618",
    "oracle_address": "panacea1ewugvs354xput6xydl5cd5tvkzcuymkejekwk3",
    "deal_id": 1,
    "provider_address": "panacea1ewugvs354xput6xydl5cd5tvkzcuymkejekwk3",
    "data_hash": "13341f91f0da76d2adb67b519e2c9759822ceafd193bd26435ba0d5eee4c3a2b"
  },
  "signature": "MEQCIEPyGSe9wVIjrUFuzXtQtEc0siwaHkp4QJCMBvC8ttWQAiAUkntIQldgIkIBFdthaTRWXHisDV2Ys/Ufpc9zejUEuQ=="
}
```

| Key                                   | Type   | Description                                                      |
|---------------------------------------|--------|------------------------------------------------------------------|
| unsigned_certificate                  | Object | Unsigned certificate containing data validation information      |
| unsigned_certificate.cid              | string | A content identifier of IPFS                                     |
| unsigned_certificate.unique_id        | string | UniqueID of the oracle that validated the data                   |
| unsigned_certificate.oracle_address   | string | Account address of the oracle that validated the data            |
| unsigned_certificate.deal_id          | int    | Deal to whom the provider intends to provide data                |
| unsigned_certificate.provider_address | string | Data provider's account address                                  |
| unsigned_certificate.data_hash        | string | The value of hashing the original data with the SHA256 algorithm |
| signature                             | string | The value of unsigned_certificate signed with OracleKey          |


### Data validation process

#### Data Decryption
Provider's encrypted data(encrypted_data_base64) can only be decrypted by Oracle.

TODO: Guide to decrypt data
```
secret_key = SHA256(ECDH(oracle_private_key, provider_public_key))

encrypted_data = Base64.Decode(encrypted_data_base64)

orgin_data = AES256GCM.Decrypt(secret_key, encrypted_data)
```

#### Data Validation
After hashing the original data with SHA256, it is encoded with HEX and compared with `data_hash`.
```
data_hash == HEX.Encode(SHA256(origin_data))
```

Verify that the `provider_address` of the original data matches the JWT Token issuer of the request header.
```
origin_data.provider_address == jwtToken.issuer
```

Deal is retrieved from the blockchain using the dealID got from the URI path.
```
deal = getDealFromBlockchain(deal_id)
```

Before verifying the data, it is checked whether the deal status is valid.
If the Deal's status is invalid, Oracle does not perform verification work.
```
deal.status == 'DEAL_STATUS_ACTIVE'
```

Validate original data with `data_schema` extracted from Deal. 
We currently support JSON Schema.
```
data_schema = deal.data_schema

JSONSchema.Validate(data_schema, orgin_data)
```

#### Data re-encryption and store


#### Consent creation and oracle sign


## Backwards Compatibility

Not applicable.

## Forwards Compatibility

TODO: blahblah

## Example Implementations

TODO: Repo URLs of panacea-core and panacea-oracle with specific tags

## Other Implementations

None at present.

## History

- 2023-01-04: Initial draft finished

## Copyright

All content herein is licensed under [Apache 2.0](https://www.apache.org/licenses/LICENSE-2.0).