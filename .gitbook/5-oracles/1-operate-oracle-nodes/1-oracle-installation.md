# Installation

## Hardware Requirement

The oracle only works on [SGX](https://www.intel.com/content/www/us/en/developer/tools/software-guard-extensions/overview.html)-[FLC](https://github.com/intel/linux-sgx/blob/master/psw/ae/ref_le/ref_le.md) environment with a [quote provider](https://docs.edgeless.systems/ego/#/reference/attest) installed.
You can check if your hardware supports SGX and is enabled in the BIOS by following [EGo guide](https://docs.edgeless.systems/ego/#/getting-started/troubleshoot?id=hardware).

## Installation: Use Docker

### Pull an image
You can pull a docker image by following CLI:
```bash
docker pull ghcr.io/medibloc/panacea-oracle:latest
```
It is highly recommended to use a specific Docker image tag instead of `latest`. 
You can find the image tags from the [GitHub Packages page](https://github.com/medibloc/panacea-oracle/pkgs/container/panacea-oracle).

### Run a container using SGX
This is a sample command to show you how to run a container using SGX in your host.
```bash
docker run \
    --device /dev/sgx_enclave \
    --device /dev/sgx_provision \
    -v ${ANY_DIR_ON_HOST}:/oracle \
    ghcr.io/medibloc/panacea-oracle:latest \
    ego run /usr/bin/oracled --help
```
After a successful installation, go to the process of [initializing oracle](./2-oracle-intialization.md).

### How about building from source?

You can build from source by referring to the following [installation-from-source](https://github.com/medibloc/panacea-oracle/blob/main/docs/installation-src.md).

However, we highly recommend installing using docker. 
This is because the uniqueID in `EGo` is sensitive to changes on the Go dependency or local debug environment.

You should check the uniqueID with following CLI when you want to use the binary you built yourself.
```bash
ego sign ${enclave-json-path} # docker image use enclave in <project-dir>/scripts/enclave-prod.json
ego uniqueid ${oracle-binary-path}
```

