# Consume data

Data consumers can open data deals by specifying the type, the quantity, and the pricing of the data that they are willing to consume. 

Data deals are registered in the Panacea public blockchain, so all data providers can find data deals, which match their data.


## Create a data deal

Broadcast the following `create-deal` transaction with desired data schema and count, and budget specified in the deal-file in JSON format.
```bash
panacead tx datadeal create-deal ${deal-file-path} \
  --from ${data-consumer-addr} \
  --chain-id ${chain-id} \
  --gas auto --gas-adjustment 1.30 --gas-prices 5umed \
  --node ${chain-node-rpc-addr}
```

For `deal-file-path`, create a following JSON file.
```json
{
  "data_schema": [
    "http://jsonschema.gopanacea.org/vaccination-cert.json"
  ],
  "budget": {
    "denom": "umed",
    "amount": "1000000"
  },
  "max_num_data": 10,
  "consumer_address": "panacea1...",
  "agreement_terms": [
    {
      "id": 1,
      "required": true,
      "title": "Terms of data provision",
      "description": "The provided data will be used for ..."
    }
  ]
}
```
It is very important to set data schema specifically and correctly, so that data being provided can be validated accurately by the oracles.

For more details about data deals, please see the [Data Deal](../../3-protocol-devs/1-dep-specs/2-data-deal.md) specification.

## Query deals

You can query a deal with its deal ID.
```bash
panacead query datadeal deal ${deal-id} \
  --node ${chain-node-rpc-addr}
```
You can also query all deals registered in the chain.
```bash
panacead query datadeal deals \
  --node ${chain-node-rpc-addr}
```


## Query consents

If some data providers have data that fit the requirements of your data deal, they will submit consents to the chain.
The consent means that the data provider has agreed to provide their data to a specific data consumer.
Also, each consent should contain a data validation certificate issued by an oracle, so that data consumers can trust the validity of data.

As soon as data providers submit their consents, you can query all consents submitted to a specific data deal.
```bash
panacead query datadeal consents ${deal-id} \
  --node ${chain-node-rpc-addr}
```
Or, you can query a specific consent, which contains a certain data hash.
```bash
panacead query datadeal consent ${deal-id} ${data-hash} \
  --node ${chain-node-rpc-addr}
```

For more details about data consents, please see the [Data Provider Consent](../../3-protocol-devs/1-dep-specs/3-data-provider-consent.md) specification.


## Access data

A data validation certificate issued by an oracle should contain the methods to access the data.
For example, if the oracle decides to transmit data via [IPFS](https://ipfs.tech/), the certificate will contain a [CID](https://docs.ipfs.io/concepts/content-addressing/) of data.
If so, you can access any IPFS node connected to the public IPFS network, and obtain the data.

For more details about data validation certificates, please see the [Data Validation](../../3-protocol-devs/1-dep-specs/4-data-validation.md) specification.

In general, the data transmitted is encrypted by oracles, so that only a specific data consumer is able to decrypt it.
Using the following REST API, you can get a secret key of each data from the oracle that issued the data validation certificate.
```bash
curl -v -X GET -H "Authorization: Bearer ${jwt}" \
  "${oracle-url}/v0/data-deal/secret-key?deal-id=${deal-id}&data-hash=${data-hash}"
```
You must specify a JWT issued using your account key in order to prove that you are the data consumer who created the data deal.
For this authentication, the JWT must be signed by your (data consumer's) chain account private key.

We highly recommend to set the expiration of JWT as short as possible for security reasons.
You can use the `panacead` CLI to issue JWTs easily by the following command.
In near future, the protocol will adopt the 'nonce' concept to improve the security of authentications.
```bash
panacead issue-jwt ${expiration-duration} --from ${your-account-key-name}

# e.g.
# panacead issue-jwt 10s --from panacea1zqum...
```

Please note that the returned secret key is also encrypted, so that only the specific data consumers can decrypt the key using his/her chain account private key.
Nevertheless, we highly recommend you to communicate with oracles who provide an HTTPS endpoint with SSL/TLS encryption.

Using the encrypted secret key that you obtained from the oracle, you can decrypt data by the following CLI.
```bash
panacead decrypt-data ${input-file-path} ${your-account-key-name} ${encrypted-secret-key} \
  --node ${chain-node-rpc-addr}
```
This command will decrypt the secret key using your account key first, and decrypt the data using the decrypted secret key.
So, please note that you should specify the `your-account-key-name` which is registered in the `panacead` keyring.
