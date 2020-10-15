# Installation

This guide will explain how to install the `panacead` and `panaceacli` entrypoints onto your system.

## Install Go

Install `go` by following the [official docs](https://golang.org/doc/install).

::: tip **Go 1.12+ +** is required for Panacea Core. :::

> _NOTE_: Before installing `panacead` and `panaceacli` binaries, let's add the golang binaries to your `PATH` variable. Open your `.bash_profile` or `.zshrc` and append `$HOME/go/bin` to your PATH variable \(i.e. `export PATH=$HOME/bin:$HOME/go/bin`\).

### Install the binaries

Next, let's install the latest version of Panacea Core. Here we'll use the `master` branch, which contains the latest stable release. If necessary, make sure you `git checkout` the correct released version.

```bash
git clone https://github.com/medibloc/panacea-core
c​d panacea-core
git checkout master
make
```

> _NOTE_: If you have issues at this step, please check that you have the latest stable version of GO installed.

That will install the `panacead` and `panaceacli` binaries. Verify that everything is OK:

```bash
$ panacead version --long
$ panaceacli version --long
```

`panaceacli` for instance should output something similar to:

```text
panacea_core: 0.0.2-25-ga25419b
commit: a25419bd3ffa07dde59f352dd015d2dd859227ef
go_sum_hash: ff10edafbe07b3e3aa879b1204cab0528ab002b0c4b37fa9dbf97e5117339e24
build tags: ledger
go version go1.12.4 darwin/amd64
```

### Build Tags

Build tags indicate special features that have been enabled in the binary.

| Build Tag | Description |
| :--- | :--- |
| ledger | Ledger devices are supported \(hardware wallets\) |

