# Contributing to KubeGuard

First off, thank you for considering contributing to KubeGuard! We value all our contributors and appreciate any help, whether it's reporting bugs, suggesting new features, or submitting Pull Requests.

## How to Report Bugs
If you find a bug, please create an issue on GitHub. Include:
- A clear, descriptive title.
- Steps to reproduce the issue.
- Expected vs. actual behavior.
- Version of KubeGuard and Kubernetes you are using.

## Suggesting Features
We welcome new ideas! Open an issue with a clear description of the feature you'd like to see, the problem it solves, and potential use cases.

## Submitting Pull Requests
1. Fork the repository and create your branch from `main`.
2. Write clear, descriptive commit messages.
3. Ensure your code follows the standard Go style guide. Run `make lint` (which runs `gofmt -l .` and `go vet ./...`).
4. Add tests for your changes. Run `make test` and `make test-race`.
5. Ensure your PR description explains the problem, the solution, and any testing performed.

## Code Style
- Follow Go conventions.
- Use `gofmt` to format code.
- Ensure `go vet ./...` passes.

## Testing Requirements
All PRs should include tests. Aim for good code coverage. You can run tests using the `Makefile` targets provided.

## Code of Conduct
By participating in this project, you agree to abide by the [Code of Conduct](CODE_OF_CONDUCT.md).
