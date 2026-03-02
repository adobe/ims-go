# Contributing

Thanks for choosing to contribute!

The following are a set of guidelines to follow when contributing to this project.

## Development

```bash
# Run all tests
go test ./... -v

# Run tests with race detection
go test -race ./...

# Static analysis
go vet ./...

# Build
go build ./...
```

## CI Pipelines

All pipelines are defined in `.github/workflows/`.

### CI (`ci.yml`)

**Triggers:** Every push to master/main and pull requests.

Runs three parallel jobs:

- **Test** — Runs `go test -race` with coverage across Go 1.21–1.25. Verifies dependency checksums with `go mod verify`. Prints a coverage summary to the log.
- **Lint** — Runs `go vet` and [golangci-lint](https://golangci-lint.run/) for extended static analysis.
- **Build** — Verifies the project compiles.

### PR Title (`pr-title.yml`)

**Triggers:** When a pull request is opened, edited, synchronized or reopened.

Validates that the PR title follows the [Conventional Commits](https://www.conventionalcommits.org) format (e.g. `feat: add token refresh`, `fix(auth): handle expired tokens`). This is enforced because the repository is configured for **squash merging only** — the PR title becomes the commit message on `master`.

### CodeQL (`codeql.yml`)

**Triggers:** Every push, pull requests targeting master/main, and weekly on a cron schedule.

Runs GitHub's CodeQL security analysis to detect vulnerabilities in the Go source code.

### govulncheck (`govulncheck.yml`)

**Triggers:** Every push to master/main, pull requests, and weekly on Monday at 9:00 UTC.

Runs Go's official vulnerability scanner ([govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck)) against all packages. The job fails if any known vulnerabilities affect the code. The weekly schedule catches new vulnerabilities even when the code hasn't changed. If the scheduled scan fails, a GitHub issue labeled `security` is created automatically.

## Repository Settings

- **Squash merge only** — Merge commits and rebase merging are disabled. The PR title is used as the squash commit message, ensuring conventional commit messages land on `master`.
- **Auto-delete branches** — Head branches are automatically deleted after a PR is merged.
- **Renovate** — [Renovate](https://docs.renovatebot.com/) monitors `go.mod` for dependency updates and opens PRs automatically. Patch updates are auto-merged after CI passes. Minor and major updates require manual review. Configuration lives in `renovate.json`.

## Code Of Conduct

This project adheres to the Adobe [code of conduct](../CODE_OF_CONDUCT.md). By participating,
you are expected to uphold this code. Please report unacceptable behavior to
[Grp-opensourceoffice@adobe.com](mailto:Grp-opensourceoffice@adobe.com).

## Have A Question?

Start by filing an issue. The existing committers on this project work to reach
consensus around project direction and issue solutions within issue threads
(when appropriate).

## Contributor License Agreement

All third-party contributions to this project must be accompanied by a signed contributor
license agreement. This gives Adobe permission to redistribute your contributions
as part of the project. [Sign our CLA](https://opensource.adobe.com/cla.html). You
only need to submit an Adobe CLA one time, so if you have submitted one previously,
you are good to go!

## Code Reviews

All submissions should come in the form of pull requests and need to be reviewed
by project committers. Read [GitHub's pull request documentation](https://help.github.com/articles/about-pull-requests/)
for more information on sending pull requests.

Lastly, please follow the [pull request template](PULL_REQUEST_TEMPLATE.md) when
submitting a pull request!

## From Contributor To Committer

We love contributions from our community! If you'd like to go a step beyond contributor
and become a committer with full write access and a say in the project, you must
be invited to the project. The existing committers employ an internal nomination
process that must reach lazy consensus (silence is approval) before invitations
are issued. If you feel you are qualified and want to get more deeply involved,
feel free to reach out to existing committers to have a conversation about that.

## Security Issues

Security issues shouldn't be reported on this issue tracker. Instead, [file an issue to our security experts](https://helpx.adobe.com/security/alertus.html).
