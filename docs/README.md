# Updating the docs


## Docs Build Workflow

The documentation for Panacea is hosted at https://docs.gopanacea.org and built from the files in the `/docs` directory.
It is built using [Docusaurus 2](https://docusaurus.io/), a modern static website generator.

### How It Works

There is a GitHub Action listening for changes in the `/docs` directory for the `master` branch and each supported version branch (e.g. `release/v2.0.x`). Any updates to files in the `/docs` directory will automatically trigger a website deployment. Under the hood, the private website repository has a `make build-docs` target consumed by a Github Action within that repository.

## README

The [README.md](./docs/README.md) is both the README for the repository and the configuration for the layout of the landing page.

## Links

**NOTE:** Strongly consider the existing links - both within this directory
and to the website docs - when moving or deleting files.

Relative links should be used nearly everywhere, due to versioning.
Note that in case of page reshuffling, you must update all links references.

### Full

The full GitHub URL to a file or directory. Used occasionally when it makes sense
to send users to the GitHub.

## Building Locally

Make sure you are in the `docs` directory and run the following commands:

```sh
rm -rf node_modules
```

This command will remove old version of the visual theme and required packages. This step is optional.

```sh
npm install
```

Install the theme and all dependencies.

```sh
npm start
```

Run `pre` and `post` hooks and start a hot-reloading web-server. See output of this command for the URL (it is often https://localhost:3000).

To build documentation as a static website run `npm run build`.

## Search

TODO: We are using [Algolia](https://www.algolia.com) to power full-text search. 