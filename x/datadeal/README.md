## Data Deal Commands

### Create a Deal
```shell
CREATOR=$(panacead keys show {your deal creator} -a)

panacead tx datadeal create-deal {your deal json file path} --from $CREATOR --chain-id {your chain ID}
```
Deal JSON File Example
```json
{
  "data_schema": [
    "https://www.json.ld"
  ],
  "budget": "10000000umed",
  "max_num_data": 10000,
  "trusted_oracles": [
    "...."
  ]
}
```

### Sell Data
```shell
SELLER=$(panacead keys show {your seller} -a)

panacead tx datadeal sell-data {your data cert json file path} --from $SELLER --chain-id {your chain ID}
```
Data Verification Certificate JSON File Example
```json
{
  "unsigned_cert": {
    "deal_id": "1",
    "data_hash": "....",
    "encrypted_data_url": "....",
    "oracle_address": "....",
    "requester_address": "...."
  },
  "signature": "...."
}
```

### Deactivate Deal
```shell
## The only creator can deactivate the deal
panacead tx datadeal deactivate-deal {deal ID} --from $CREATOR --chain-id {your chain ID}
```

### Query a Deal
```shell
panacead q datadeal deal {deal ID} --chain-id {your chain ID}
```