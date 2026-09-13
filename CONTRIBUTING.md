# Contributing to NetGuard VPN

We love your input! We want to make contributing to this project as easy and transparent as possible.

## Development Setup

1. Make sure you have Go 1.22+ installed.
2. Clone the repository: `git clone https://github.com/netguard-vpn/netguard.git`
3. Install dependencies: `go mod download`
4. Build the project: `make build`
5. Run tests: `make test`

## Branch Strategy

- The `main` branch is our stable branch.
- For new features and bug fixes, please create a feature branch (`feature/your-feature-name` or `bugfix/issue-description`).

## Pull Request Process

1. Ensure any install or build dependencies are removed before the end of the layer when doing a build.
2. Update the README.md with details of changes to the interface.
3. Verify that your code passes all lint checks (`make lint`).
4. Ensure tests pass (`make test`) and add tests for your new code.

## Commit Message Format

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:
- `feat: add new feature`
- `fix: resolve a bug`
- `docs: update documentation`

## Coding Standards

- Run `make fmt` before committing (gofmt, goimports).
- Pass all checks in `golangci-lint`.
- Ensure new code is covered by unit tests.

## Issue Reporting Guidelines

- Use the provided issue templates.
- Provide step-by-step reproduction instructions for bugs.

## Code of Conduct

Please note that this project is released with a [Contributor Code of Conduct](CODE_OF_CONDUCT.md). By participating in this project you agree to abide by its terms.
