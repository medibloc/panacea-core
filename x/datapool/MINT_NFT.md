### Deploy and Register NFT smart contract

```shell
TX_FLAG=(--chain-id {your chainID} --gas auto --gas-prices 5umed --gas-adjustment 1.3)
MINTER=$(panacead keys show {your address or key of minter} -a)
panacead tx datapool deploy-and-register-contract cw721_base.wasm --from $MINTER $TX_FLAG -y
```

### Create data pool

```shell
CURATOR=$(panacead keys show {your address or key of curator} -a)
panacead tx datapool create-pool create-pool.json --from $CURATOR $TX_FLAG -y
```

### Query curator NFT
```shell
CONTRACT=$(panacead q datapool get-contract | cut -d' ' -f2)
QUERY_TOKEN_INFO=$(jq -n --arg owner $CURATOR '{"tokens":{"owner":$owner}}')
panacead q wasm contract-state smart $CONTRACT $QUERY_TOKEN_INFO -o json
```

query result : 
```json
{
  "data": {
    "tokens":["data_pool_0"]
  }
}
```