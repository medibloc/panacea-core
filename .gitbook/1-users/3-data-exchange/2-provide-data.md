# Provide data

Data providers can provide their data that match the requirements of the data deal and, in return, earn a reward in MED.

Since only the data verified by oracle can be provided to a deal, providers should first request data verification to oracle.

As a result of verification, oracle will issue a certificate; then provider can provide their data by submitting a consent with the certificate to Panacea.

In the whole process of data transmission, the data is encrypted so that no one can access the original data.

The transmission of data and the payment of rewards are managed atomically through Panacea, so consumers can get the data only when the reward payment is completed.

## Request data verification to oracle

Before requesting data verification to oracle, you (data provider) should encrypt the data to preserve privacy.
Encryption can be done using your chain account key and oracle public key which is registered in Panacea.
This makes only oracles can decrypt and verify your original data in secure area.
For more details about data secure area, please see [Confidential Oracle](../../3-protocol-devs/1-dep-specs/5-confidential-oracle.md).

You can encrypt data by the following CLI:
```bash
panacead encrypt-data ${input-file-path} ${your-account-key-name}
```

You must specify a JWT issued by yourself in order to prove that you are the data provider.
For that authentication, the JWT must be signed by your chain account private key.

We highly recommend to set the expiration of JWT as short as possible for security reasons.
In the near future, the protocol will adopt the 'nonce' concept to improve the security of authentications.

You can issue JWT by the following CLI:
```bash
panacead issue-jwt ${expiration-duration} --from ${your-account-key-name}

# e.g.
# panacead issue-jwt 10s --from panacea1zqum...
```

Using the following REST API, you can post encrypted data to oracle for verification with deal ID you want to provide.
```bash
curl -v -X POST -H "Authorization: Bearer ${jwt}" ${oracle-url}/v0/data-deal/deals/${deal-id}/data -d ${encrpyted-data-json-path}
```

For `encrpyred-data-json`, create a following JSON file.
```json
{
  "provider_address" : "{your-address}",
  "data_hash" : "{data-hash-of-original-data}",
  "encrypted_data_base64" : "{encrypted-data}"
}
```
You have to use data hash of original data with SHA-256 hash function. For example, you can get hash by following CLI:
```bash
sha256sum ${original-data-path}
```

Through this process, the data is safely and securely transmitted to oracle, and the oracle verifies the data hash and data schema.
For more details about oracle data validation, please see the [Data Validation](../../3-protocol-devs/1-dep-specs/4-data-validation.md) specification.

When the verification is completed, you can get a data certificate that includes the oracle's signature.
The certificate form is like below:
```json
{
  "unsigned_certificate" : {
    "cid" : "{ipfs-cid}",
    "unique_id" : "{oracle-unique-id}",
    "oracle_address" : "{oracle-address}",
    "deal_id": <deal-id>,
    "provider_address" : "{your-address}",
    "data_hash" : "{data-hash}"
  },
  "signature" : "{oracle-signature}"
}
```
Now you can use this certificate in the next step.

## Submit consent

Broadcast the following `submit-consent` transaction with the certificate from oracle and agreements of terms for data provision.

**consent.json**

```json
{
  "deal_id": <deal-id>,
  "certificate": {
    ...
  },
  "agreements": [
    {
      "id": 1,
      "agreement": true
    }
  ]
}
```

```bash
panacead submit-consent ${consent-file-path} \
  --from ${data-provider-addr} \
  --chain-id ${chain-id} \
  --gas auto --gas-adjustment 1.30 --gas-prices 5umed \
  --node ${chain-node-rpc-addr}
```

After you submit consent, Panacea verifies the certificate and checks the status of the deal. 

When the verification is completed, Panacea makes the data accessible to consumers and transmitts rewards to you.

## Query consent
You can query a consent by the deal ID and data hash you provided.
```bash
panacead query datadeal consent ${deal-id} ${data-hash} \
  --node ${chain-node-rpc-addr}
```
