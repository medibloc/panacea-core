## Deploy and instantiate NFT smart contract

### Setting voter
```shell
TX_FLAG=(--chain-id {your chainID} --gas auto --gas-prices 5umed --gas-adjustment 1.3)
VOTER=$(panacead keys show {your voter} -a)
panacead tx staking delegate ${validator address} 1000000umed --from $VOTER $TX_FLAG -y
```

### Submit proposal (1): store NFT contract
```shell
PROPOSER=$(panacead keys show {your proposer} -a)

MODULE_ADDR="panacea1xacc5pqnn00vf4mf8qvhe3y7k0xj4ky2hxgzvz"  // TODO: add GetParam query to get module address

panacead tx gov submit-proposal wasm-store cw721_base.wasm \
--title "store NFT contract wasm code" \
--description "store wasm code for x/datapool module" \
--instantiate-only-address $MODULE_ADDR \
--run-as $MODULE_ADDR \
--deposit "10000000000umed" \
--from $PROPOSER $TX_FLAG -y
```

The module is the only allowed address to instantiate the contract

### Vote yes
```shell
panacead tx gov vote {store proposal id} yes --from $VOTER $TX_FLAG -y
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
--from $PROPOSER $TX_FLAG -y
```

### Vote yes
```shell
panacead tx gov vote {instantiation proposal id} yes --from $VOTER $TX_FLAG -y
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
      "key": "datapoolcodeid",
      "value": "1"
    },
    {
      "subspace": "datapool",
      "key": "datapoolnftcontractaddress",
      "value": "panacea14hj2tavq8fpesdwxxcu44rty3hh90vhu4mda6e"
    }
  ],
  "deposit": "1000000umed"
}
```

```shell
panacead tx gov submit-proposal param-change param_change_sample.json --from $PROPOSER $TX_FLAG -y
```

### Vote yes
```shell
panacead tx gov vote {param-change proposal id} yes --from $VOTER $TX_FLAG -y
```

### Create data pool

```shell
CURATOR=$(panacead keys show {your address or key of curator} -a)
panacead tx datapool create-pool create_pool_sample.json --from $CURATOR $TX_FLAG -y
```

### Query curator NFT
```shell
CONTRACT="panacea14hj2tavq8fpesdwxxcu44rty3hh90vhu4mda6e"
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

## Change parameter of data pool deposit by proposal

proposal_example.json
```json
{
  "title": "title",
  "description": "description",
  "changes": [
    {
      "subspace": "datapool",
      "key": "datapooldeposit",
      "value": { "denom": "umed", "amount": "20000000" }
    }
  ],
  "deposit": "1000000umed"
}
```

```shell
panacead tx gov submit-proposal param-change param_example.json --from {proposer account} --chain-id {your chainID} --gas auto --gas-prices 5umed --gas-adjustment 1.3 -y
```