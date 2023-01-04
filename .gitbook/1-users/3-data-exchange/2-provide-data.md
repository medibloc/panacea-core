# Provide data

Data providers can provide their data that match the requirements of the data deal and in return earn a reward.

The provider request verification of the data to oracle, and submits consent with oracle signature to the Panacea public blockchain.

Since the data is encrypted throughout the whole process of data transmission, no one can get the original data.

The transmission of data and the payment of rewards are managed atomically through the Panacea public blockchain, so consumers can get the data only when the reward payment is completed.

## Post data to oracle

Encrypt data that you wants to provide. 
Encryption is done using your chain account key and oracle public key.
This makes only oracles can decrypt your original data.

You can encrypt data by the following CLI:
```bash
panacead encrypt-data ${input-file-path} ${your-account-key-name}\
  --chain-id ${chain-id}
```

You must specify a JWT issued by yourself in order to prove that you are the data provider.
For that authentication, the JWT must be signed by your (data provider's) chain account private key.

We highly recommend to set the expiration of JWT as short as possible for security reasons.
In near future, the protocol will adopt the 'nonce' concept to improve the security of authentications.

You can issue jwt by the following CLI:
```bash
panacead issue-jwt ${expiration}\
  --chain-id ${chain-id}
```

Using the following REST API, you can post encrypted data to oracle for verification with deal ID you want to provide.
```bash
curl -v -X POST -H "Authorization: Bearer ${jwt}" {oracle-url}/v0/data-deal/deals/{deal-id}/data -d ${encrpyted-data-json}
```

For `encrpyred-data-json`, create a following JSON file.
```json
{
  "provider_address" : "{your_address}",
  "data_hash" : "{data-hash-of-original-data}",
  "encrypted_data_base64" : "{encrypted-data}"
}
```
Through this process, the data is safely and securely transmitted to oracle, and the oracle verifies the data hash and data schema.
For more details about oracle data validation, please see the [Data Validation](../../3-protocol-devs/1-dep-specs/4-data-validation.md) specification.

When the verification is complete, you can get a data certification that includes the oracle's signature.

## Submit consent

Broadcast the following `submit-consent` transaction with certification from oracle.
```bash
panacead submit-consent ${certifiacte-file}\
  --from ${data-provider-addr} \
  --chain-id ${chain-id} \
  --gas auto --gas-adjustment 1.30 --gas-prices 5umed \
  --node ${chain-node-rpc-addr}
```

After you submit consent, Panacea public blockchain verifies the certificate and checks the status of the deal. 

When verification is complete, Panacea makes the data accessible to consumers and makes you can get reward.


## Query consent
You can query a consent with deal ID you provided and its data hash .
```bash
panacead query datadeal consent ${deal-id} ${data-hash} \
  --node ${chain-node-rpc-addr}
```