## Deploy and Register NFT smart contract

```shell
TX_FLAG=(--chain-id {your chainID} --gas auto --gas-prices 5umed --gas-adjustment 1.3)
SENDER=$(panacead keys show {your address or key of sender} -a)
panacead tx datapool register-contract cw721_base.wasm --from $SENDER $TX_FLAG -y
```

In this example, the `SENDER` doesn't do anything except paying tx fees for registration of NFT contract. 

## Create data pool

```shell
CURATOR=$(panacead keys show {your address or key of curator} -a)
panacead tx datapool create-pool create-pool.json --from $CURATOR $TX_FLAG -y
```

## Upgrade NFT contract
```shell
panacead tx datapool upgrade-contract cw721_new.wasm --from $SENDER $TX_FLAG -y
```

## Query curator NFT
```shell
CONTRACT=$(panacead q datapool get-contract | cut -d' ' -f2)
QUERY_TOKEN_INFO=$(jq -n --arg owner $CURATOR '{"tokens":{"owner":$owner}}')
panacead q wasm contract-state smart $CONTRACT $QUERY_TOKEN_INFO -o json
```
### results
```json
{
  "data": {
    "tokens":["data_pool_0"]
  }
}
```