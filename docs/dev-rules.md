# Dev rules

Adapted from budget-server's `plan/dev-rules.md`. Where this repo and budget-server differ, the differences are listed in [phase 0](plan/phase-0-setup.md#deliberate-differences); everything else should match.

## Testing

- When writing tests, make sure that they pass before completion.
- Don't use the message field on `assert` or `require`.
- Unit tests stay in this repo (`go test ./...`) and must not need the stack. They are necessary but not sufficient: a change that crosses a service boundary is not done until it has been exercised in budget-stack.

## General

- Don't forget to regenerate the mocks (`make generate-mocks`).
- `make format` to verify formatting changes.
- Go is installed with mise (see `mise.toml`).
- Avoid `go run` or `go build`; build with docker instead.
- DB enums are explicit integers, never `iota`.
- Comments: one-line "why" only.

## Running and testing locally

Run and test in an integrated way, through budget-stack (`../budget-stack`), not service by service.

- **budget-stack is the local environment.** Anything beyond unit tests — booting the service, calling an endpoint, checking a migration, clicking through the web UI — is done with the whole stack up: Postgres, budget-server, budget-ingest, the web app and the edge proxy.
- **Use local builds of what you changed.** `make start-local` in budget-stack builds images from the sibling checkouts and runs them in place of the pinned releases; `make rebuild-local s=budget-ingest` rebuilds and restarts this service. `make start` runs the pinned releases only.
- **Before budget-stack knows about budget-ingest** (phase 0 tasks P0-I1 to P0-I3), use this repo's own `make serve-local` and `make serve-local-rebuild`.
- **When a change spans repos,** check out each repo's branch side by side and run them together in the stack before opening any of the PRs. Say in each PR which branches of the other repos it was run against.
- **Do not create another multi-service setup.** This repo's `docker-compose.yaml` exists to match budget-server's structure; it is not where integration is checked. If the stack is missing something, add it to budget-stack.
- **Record what you ran.** Each PR description lists the stack commands used and what was observed.
