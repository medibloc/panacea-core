# Contributing

* [General Procedure](#general-procedure)
* [Development Procedure](#development-procedure)
    * [Testing](#testing)
    * [Pull Requests](#pull-requests)
    * [Requesting Reviews](#requesting-reviews)
    * [Updating Documentation](#updating-documentation)


Thank you for considering making contributions to the Panacea and related repositories!

## General Procedure

Contributing to this repo can mean many things, such as participating in
discussion or proposing code changes. To ensure a smooth workflow for all
contributors, the general procedure for contributing has been established:

1. Start by browsing [new issues](https://github.com/medibloc/panacea-core/issues) and [discussions](https://github.com/medibloc/panacea-core/discussions). If you are looking for something interesting or if you have something in your mind, there is a chance it had been discussed.
2. Determine whether a GitHub issue or discussion is more appropriate for your needs:
   1. If want to propose something new that requires specification or an additional design, or you would like to change a process, start with a [new discussion](https://github.com/medibloc/panacea-core/discussions/new). With discussions, we can better handle the design process using discussion threads. A discussion usually leads to one or more issues.
   2. If the issue you want addressed is a specific proposal or a bug, then open a [new issue](https://github.com/medibloc/panacea-core/issues/new/choose).
   3. Review existing [issues](https://github.com/medibloc/panacea-core/issues) to find an issue you'd like to help with.
3. Participate in thoughtful discussion on that issue.
4. If you would like to contribute:
   1. Ensure that the proposal has been accepted.
   2. Ensure that nobody else has already begun working on this issue. If they have,
      make sure to contact them to collaborate.
   3. If nobody has been assigned for the issue and you would like to work on it,
      make a comment on the issue to inform the community of your intentions
      to begin work.
5. To submit your work as a contribution to the repository follow standard GitHub best practices. See [development procedure](#development-procedure) below.

**Note:** For very small or blatantly obvious problems such as typos, you are
not required to an open issue to submit a PR, but be aware that for more complex
problems/features, if a PR is opened before an adequate design discussion has
taken place in a GitHub issue, that PR runs a high likelihood of being rejected.

## Development Procedure

* The latest state of development is on `main`.
* `main` must never fail `make lint build test`.
* No `--force` onto `main` (except when reverting a broken commit, which should seldom happen).
* Create a branch to start work:
    * Fork the repo (core developers must create a branch directly in the `panacea-core` repo),
    branch from the HEAD of `main`, make some commits, and submit a PR to `main` in the `panacea-core` repo.
    * For core developers working within the `panacea-core` repo, follow branch name conventions to ensure a clear
    ownership of branches: `{issue#}-branch-name`. It is also recommended to use the [`Create a branch`](https://docs.github.com/en/issues/tracking-your-work-with-issues/creating-a-branch-for-an-issue) feature in the GitHub Issues.

All pull requests are merged to the `main` after being reviewed, based on the [trunk-based development](https://trunkbaseddevelopment.com/).

### Testing

Tests can be executed by running `make test` at the top level of the `panacea-core` repository.

### Pull Requests

Before submitting a pull request:

* merge the latest main `git merge origin/main`,
* run `make lint build test` to ensure that all checks and tests pass. This is also ran automatically by [GitHub Actions](./.github/workflows) in the `panacea-core` repo.

Then:

1. If you have something to show, **start with a `Draft` PR**. It's good to have early validation of your work and we highly recommend this practice. A Draft PR also indicates to the community that the work is in progress.
   Draft PRs also helps the core team provide early feedback and ensure the work is in the right direction.
2. When the code is complete, change your PR from `Draft` to `Ready for Review`.
3. Go through the actions for each checkbox present in the PR template description. The PR actions are automatically provided for each new PR.
4. Be sure to include a relevant changelog entry in the `Unreleased` section of `CHANGELOG.md`. The entry should be on top of all others changes in the section.

PRs must have a category prefix that is based on the type of changes being made (for example, `fix`, `feat`,
`refactor`, `docs`, and so on). The *type* must be included in the PR title as a prefix (for example,
`fix: <description>`). This convention ensures that all changes that are committed to the base branch follow the
[Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification.
Additionally, each PR should only address a single issue.

Pull requests are merged by [code owners](./CODEOWNERS).
All pull requests must be squashed before being merged, and also be rebased on top of the `main`.

### Requesting Reviews

[Code owners](./CODEOWNERS) are marked automatically as the reviewers.
All PRs require at least two review approvals before they can be merged.

All PRs are squashed and merged by code owners after being reviewed.
Please note that PRs based on outdated `main` cannot be merged. Those PRs must be updated or rebased on top of the `main` by PR authors.

### Updating Documentation

If you open a PR, it is mandatory to update the relevant documentation in [`/.gitbook`](.gitbook).