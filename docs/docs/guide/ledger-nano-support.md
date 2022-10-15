---
sidebar_position: 6
---

# Ledger Nano Support

Using a hardware wallet to store your keys improves the security of your assets. 
The [Ledger](https://www.ledger.com/) device acts as an enclave of the seed and private keys,
and the process of signing transaction takes place within it.
The following is a short tutorial on using the Panacea Ledger app with the Panacea CLI.


## Install the Panacea Ledger app

Installing the `Panacea` app on your Ledger device is required before using the Panacea CLI.

1. Install [Ledger Live](https://www.ledger.com/ledger-live/) on your machine.
2. Using Ledger Live, update your Ledger Nano device with the latest firmware.
3. Navigate to the `Manager` menu on the Ledger Live.
4. Connect your Ledger Nano device and allow Ledger Manager from it.
5. On the Ledger Live, search for `Panacea` and install it.


## Panacea CLI + Ledger Nano

Note: You need to [install the `Panacea` app](#install-the-panacea-ledger-app) on your Ledger Nano device before using following this section.

### Before you begin

Install `panacead` by following the [guide](installation.md).

### Add your Ledger key

- Connect and unlock your Ledger Nano device.
- Open the `Panacea` app on your Ledger Nano device.
- Create an account in `panacead` from your Ledger key.

```bash
panacead keys add <keyName> --ledger
```
Note: Be sure to change the `<keyName>` parameter to be a meaningful name. The `--ledger` flag tells `panacead` to use your Ledger to seed the account.

Panacea uses HD wallets. This means you can setup many accounts using the same Ledger seed.
To create another account from your Ledger device, run the following command by changing the integer `<i>`
to some value >= 0 to choose the account for HD derivation.
```bash
panacead keys add <keyName> --ledger --account <i>
```

### Confirm your address

Run the following command to display your address on the Ledger Nano device. Use the `<keyName>` you gave your Ledger key.
```bash
panacead keys show <keyName> -d
```

Confirm that the address displayed on the Ledger Nano device matches that displayed when you added the key.

### Sign a transaction (Send funds)

You are now ready to start signing and sending transactions.
```bash
panacead tx bank send <fromAddress> <toAddress> <amount>umed \
  --node http://54.180.169.37:26657 \
  --chain-id panacea-2 \
  --fees 1000000umed
```
Be sure to set a proper full node address to the `--node` parameter. In the above example, one of the [Panacea Mainnet full nodes](https://github.com/medibloc/panacea-launch) was used.
Also, the `--chain-id` must be set as `panacea-2` which is the current chain ID of the Panacea Mainnet.

For `--fees`, please set `1000000umed` which is the approximate minimum in the current Mainnet network.

You will be prompted to review and approve the transaction on your Ledger Nano device.
Be sure to inspect the transaction displayed on the screen.

### Receive funds

To receive funds to the Panacea account on your Ledger Nano device,
retrieve the address for your Ledger Nano device (the ones with `type: ledger`) with this command:
```bash
panacead keys list

- name: <keyName>
  type: ledger
  address: panacea1...
```

### Support

For further support,
- support@gopanacea.org
