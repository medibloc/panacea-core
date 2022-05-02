## Deploy and instantiate NFT smart contract

### Submit proposal (1): store NFT contract
```shell
VALIDATOR=$(panacead keys show {your validator} -a)

MODULE_ADDR=$(panacead q datapool module-addr -o json | jq -r '.address')

panacead tx gov submit-proposal wasm-store cw721_base.wasm \
--title "store NFT contract wasm code" \
--description "store wasm code for x/datapool module" \
--instantiate-only-address $MODULE_ADDR \
--run-as $MODULE_ADDR \
--deposit "10000000000umed" \
--from $VALIDATOR $TX_FLAG -y
```

The module is the only allowed address to instantiate the contract

### Vote yes
```shell
panacead tx gov vote {store proposal id} yes --from $VALIDATOR $TX_FLAG -y
```

### Submit proposal (2): instantiate NFT contract

After store NFT contract passed, instantiate the contract

```shell
INST_MSG=$(jq -n --arg name "curator" --arg symbol "CUR" --arg minter $MODULE_ADDR '{"name": $name, "symbol": $symbol, "minter": $minter}')

panacead tx gov submit-proposal instantiate-contract {code id} "$INST_MSG" \
--label "curator NFT" \
--title "instantiate NFT contract" \
--description "instantiate NFT contract for x/datapool module" \
--run-as MODULE_ADDR \
--admin MODULE_ADDR \
--deposit "100000000umed" \
--from $VALIDATOR $TX_FLAG -y
```

### Vote yes
```shell
panacead tx gov vote {instantiation proposal id} yes --from $VALIDATOR $TX_FLAG -y
```

### Submit proposal (3): change parameter of code ID & NFT contract address

param_change_sample.json (when codeID=1, contractAddress=panacea14hj2tavq8fpesdwxxcu44rty3hh90vhu4mda6e)
```json
{
  "title": "parameter change of datapool module",
  "description": "register code ID and address of NFT contract",
  "changes": [
    {
      "subspace": "datapool",
      "key": "DataPoolCodeId",
      "value": "1"
    },
    {
      "subspace": "datapool",
      "key": "DataPoolNftContractAddress",
      "value": "panacea14hj2tavq8fpesdwxxcu44rty3hh90vhu4mda6e"
    }
  ],
  "deposit": "1000000umed"
}
```

TODO: contract address is hardcoded now.

```shell
panacead tx gov submit-proposal param-change param_change_sample.json --from $VALIDATOR $TX_FLAG -y
```

### Vote yes
```shell
panacead tx gov vote {param-change proposal id} yes --from $VALIDATOR $TX_FLAG -y
```

### Create data pool

```shell
CURATOR=$(panacead keys show {your address or key of curator} -a)
panacead tx datapool create-pool {your deposit} create_pool_sample.json --from $CURATOR $TX_FLAG -y
```

### Query curator NFT
```shell
CONTRACT=$(panacead q datapool params -o json | jq -r '.data_pool_nft_contract_address')
QUERY_TOKEN_INFO=$(jq -n --arg owner $CURATOR '{"tokens":{"owner":$owner}}')
panacead q wasm contract-state smart $CONTRACT $QUERY_TOKEN_INFO -o json
```
result
```json
{
  "data": {
    "tokens":["data_pool_1"]
  }
}
```

## Change parameter of data pool commossion by proposal

proposal_example.json
```json
{
  "title": "title",
  "description": "description",
  "changes": [
    {
      "subspace": "datapool",
      "key": "DataPoolCuratorCommissionRate",
      "value": "0.05"
    }
  ],
  "deposit": "1000000umed"
}
```

```shell
panacead tx gov submit-proposal param-change param_example.json --from $VALIDATOR $TX_FLAG -y
```