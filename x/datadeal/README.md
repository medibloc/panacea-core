## Data Deal Commands

### Create a Deal
```shell
CREATOR=$(panacead keys show {your deal creator} -a)

panacead tx datadeal create-deal --deal-file {your deal json file path} --from $CREATOR --chain-id {your chain ID}
```

### Sell Data
```shell
SELLER=$(panacead keys show {your seller} -a)

panacead tx datadeal sell-data --data-verification-certificate-file {your data cert json file path} --from $SELLER --chain-id {your chain ID}
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