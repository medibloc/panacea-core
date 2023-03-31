# Oracle Registration

This document is an instruction to register an oracle to Panacea

## Get Trusted Block Information

To request an oracle registration to Panacea, trusted block information (height and hash), which will be used for [light client](), is required and need to be verified by other registered oracle.

You can get trusted block information by:
```bash
BLOCK=$(panacead q block --node <node-rpc-address>)

HEIGHT=$(echo $BLOCK | jq -r .block.header.height)
HASH=$(echo $BLOCK | jq -r .block_id.hash)
```

If you need more information about public RPC endpoints provided by MediBloc team, you can refer to [this](https://github.com/medibloc/panacea-mainnet#public-endpoints)

## Request to Register Oracle

In addition to the trusted block information, the following arguments are also required in the oracle registration.

| Argument                          | requirement | Description                                                                                          |
|-----------------------------------|-------------|------------------------------------------------------------------------------------------------------|
| oracle-commission-rate            | required    | The desired initial oracle commission rate                                                           |
| oracle-commission-max-rate        | required    | The maximum oracle commission rate. The oracle commission rate cannot be greater than this max rate. |
| oracle-commission-max-change-rate | required    | The maximum rate that an oracle can change once. It will be reset 24 hours after the last change.    |
| oracle-endpoint                   | optional    | The endpoint of oracle to be used                                                                    |

With the above arguments, you can now request to register your oracle to Panacea.

```bash
docker run \
    --device /dev/sgx_enclave \
    --device /dev/sgx_provision \
    -v <directory-you-want>/oracle:/oracle \
    ghcr.io/medibloc/panacea-oracle:main \
    ego run /usr/bin/oracled register-oracle \ 
    --trusted-block-height ${HEIGHT} \
    --trusted-block-hash ${HASH} \
    --oracle-commission-rate 0.1 \
    --oracle-commission-max-rate 0.3 \
    --oracle-commission-max-change-rate 0.01 \
    --oracle-endpoint "<your-oracle-endpoint>"
```

Then, a new key pair, called `Node Key`, and its remote report will be generated.
The `Node Key` is used to transfer the oracle private key securely.
Since the oracle private key must be shared in a very secure way, the key will be encrypted with the public key of the `Node Key` so that only the oracle can decrypt it.

{% hint style="danger" %}
The private key of `Node Key` is also stored in a secure way as a sealed file, named `node_priv_key.sealed` in default.
This `node_priv_key.sealed` file is the only clue to decrypt the encrypted oracle private key, and, hence, we highly recommend you to manage it very carefully in case you need to restore the oracle private key again later.
{% endhint %}

If you have tried to request `register-oracle` before, previously generated `node_priv_key.sealed` would exist.
And the app will ask, 

```
There is an existing node key.
Are you sure to delete and re-generate node key?
```

If you're sure to re-generate the `Node Key` and re-request oracle registration, please enter `y`.
Or, we recommend to back up the existing `node_priv_key.sealed` file.

## Subscribe Approval of Registration

If the transaction for requesting oracle registration succeeds, the newly registered oracle will start subscribing to `ApproveOracleRegistrationEvent`.
Other oracles that are already registered will do some verifications of this registration by checking if:
- correct version of oracle binary is used
- the oracle is running inside an enclave 
- the `Node Key` is generated inside the enclave
- the trusted block information is valid

When the registration is verified successfully, other oracles will send a transaction for approval of the oracle registration.
Since the oracle is already subscribing to the event, it will detect the approval, and then retrieve the oracle private key.
As a result of the oracle private key retrieval, a sealed file named `oracle_priv_key.sealed` (default file name) will be generated.
Using this sealed oracle private key, the oracle is now ready to perform tasks such as verifying data or other oracles.

For more information about what oracle does, please refer the [running an oracle node](5-running-node#what-oracle-does.md).

{% hint style="danger" %}
Like `node_priv_key.sealed`, `oracle_priv_key.sealed` is also very important for working as a valid oracle.
Thus, we highly recommend you to manage the `oracle_priv_key.sealed` file very carefully.
In case an operator loses the `oracle_priv_key.sealed` file for any reason, it can be retrieved again using `get-oracle-key` CLI with the `node_priv_key.sealed` file, which is generated when oracle registration is requested as above.
{% endhint %}

## Manually Retrieve Oracle Private Key

Because of any unforseen reasons, the oracle could stop before the oracle private key is retrieved. 
In this case, you can also retrieve the oracle private key through `get-oracle-key` CLI.
If at least one registered oracle approves this registration, you will be able to retrieve the oracle private key successfully.

```bash
docker run \
    --device /dev/sgx_enclave \
    --device /dev/sgx_provision \
    -v ${ANY_DIR_ON_HOST}:/oracle \
    ghcr.io/medibloc/panacea-oracle:main \
    ego run /usr/bin/oracled get-oracle-key
```

You can check the status of your registration with the uniqueID and oracle address.

```bash
panacead q oracle oracle-registration <unique-id> <oracle-address>
```

If at least one oracle approves this registration, the `encrypted_oracle_priv_key` will not be empty. 

Now, you are ready to run your own oracle. Please refer the [running a node](./5-running-node.md) instructions.
