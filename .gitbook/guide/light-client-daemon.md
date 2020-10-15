# Light Client Daemon‌

To start a REST server, we need to specify the following parameters:

| Parameter | Type | Default | Required | Description |
| :--- | :--- | :--- | :--- | :--- |
| chain-id | string | null | true | chain id of the full node to connect |
| node | URL | "tcp://localhost:26657" | false | address of the full node to connect |
| laddr | URL | "tcp://localhost:1317" | false | address to run the rest server on |
| trust-node | bool | "false" | false | Whether this LCD is connected to a trusted full node |
| trust-store | DIRECTORY | "$HOME/.lcd" | false | directory for save checkpoints and validator sets |

For example::

```bash
panaceacli rest-server \
    --chain-id=test \
    --laddr=tcp://localhost:1317 \
    --node=tcp://localhost:26657 \
```

The server listens on HTTP by default. You can enable the secure layer by adding the `--tls` flag. By default a self-signed certificate will be generated and its fingerprint printed out. You can configure the server to use a SSL certificate by passing the certificate and key files via the `--ssl-certfile` and `--ssl-keyfile` flags:

```bash
panaceacli rest-server \
    --chain-id=test \
    --laddr=tcp://localhost:1317 \
    --node=tcp://localhost:26657 \
    --tls \
    --ssl-certfile=mycert.pem --ssl-keyfile=mykey.key
```

