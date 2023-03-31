# Update Oracle Information

This document outlines how to edit the oracle information (endpoint and commission rate)

## Command Line for Updating

Using below CLI, you can update your endpoint and/or commission rate

```bash
docker run \
    --device /dev/sgx_enclave \
    --device /dev/sgx_provision \
    -v <directory-you-want>/oracle:/oracle \
    ghcr.io/medibloc/panacea-oracle:main \
    ego run /usr/bin/oracled update-oracle-info \ 
    --oracle-endpoint "<your-new-endpoint>" \ 
    --oracle-commission-rate <your-new-commission-rate>
```

| Flag                              | requirement | Description                 |
|-----------------------------------|-------------|-----------------------------|
| endpoint                          | optional    | The endpoint of oracle      |
| oracle-commission-rate            | optional    | The oracle commission rate. |

As you can see from above table, both flags are optional.
You can choose and enter only the flags you want to change.
It is noted that the oracle commission rate cannot be greater than oracle max change rate.
And once changed, the oracle commission rate cannot be changed within the next 24 hours.

You can confirm the changes you made when the transaction succeeds

```bash
panacead q oracle oracle <your-oracle-address>
```
