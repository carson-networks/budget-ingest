# Working rules

> **Abandoned 2026-11-13.** The FinanceKit entitlement was approved, so this LunchSync bridge will not be built. Kept for reference only; do not implement it.

Rules for the implementing agent. They apply to every task in every phase. Back to the [overview](README.md).

Each phase is a vertical slice across every system and ends with something a person can see working on the phone. Do not start a phase until the previous phase's exit check has passed.

## Systems

| Key | Repo | Local path | What changes there |
|---|---|---|---|
| **S** | `carson-networks/budget-server` | `../budget-server` | Sync type, `LunchSyncService`, storage |
| **I** | `carson-networks/budget-ingest` | this repo | The Lunch Money-compatible bridge |
| **W** | `carson-networks/budget` | `../budget` | Web UI |
| **X** | `carson-networks/infra-prod`, `carson-networks/budget-stack` | `../infra-prod`, `../budget-stack` | Routing, certificate, compose |

Task IDs are `P<phase>-<system><n>`. Do a phase's tasks in the order its progress table lists them.

## Who does what

- **Agent** tasks are everything not marked otherwise.
- **Human** tasks need the phone, the root CA key, DNS, or production access. Do not attempt them. Prepare whatever they need (files, exact commands, a checklist), then stop and report.
- A phase's exit check is always run by the human, after releasing and deploying that phase. Finish all agent tasks, summarise what to test, and stop.

## Rules for every task

- One branch and one PR per task per repo, branched from that repo's default branch. Never push to a default branch, never merge, never tag a release.
- Never commit keys, certificates, API keys or `.env` values. Never print an API key to a log.
- Do not copy code from the LunchSync repository; it has no licence. Reading it to learn the wire format is fine.
- Comments: one-line "why" only, no long explanatory blocks.
- Do not weaken or delete an existing test to make a change pass.
- Each phase file opens with a progress table. Do not tick a row yourself: the human ticks it when the task's PR is merged. If a task turns out to need splitting or a new task appears, add a row and say so in the PR.
- A task is done when its acceptance items hold and the repo's checks pass. Paste the check output in the PR description. If a check cannot be run, say so; do not claim it passed.
- Stop and ask instead of guessing when: a task conflicts with the contract or README, an existing behaviour would have to change for Plaid or manual accounts, or a migration would touch existing rows.

## Repo checks

| Repo | Run before opening a PR | Conventions |
|---|---|---|
| budget-server | `make format`, `go test ./...`, then `make start-local` in budget-stack and exercise the change there | `plan/dev-rules.md`. Build with docker, not `go run`. After storage interface changes: `make generate-mocks`. After migrations: `make migrate` then `make generate`. After proto changes: `make proto-generate`. No message argument on `assert`/`require`. DB enums are explicit integers, never `iota`. |
| budget-ingest | `make format`, `go test ./...`, then `make start-local` in budget-stack and exercise the change there | `docs/dev-rules.md`. Same as budget-server in every respect; see [phase 0](phase-0-setup.md) for what "the same" covers and the short list of deliberate differences. |
| budget | `pnpm lint`, `pnpm test`, `pnpm build`, then `make start-local` in budget-stack and check the change in the browser | `spec/arch.md`. Server state only through `src/queries/`; only `src/connectRPC/` imports generated code; logic in pure functions with unit tests. |

## Running and testing locally

Run and test in an integrated way, through budget-stack (`../budget-stack`), not service by service.

- **budget-stack is the local environment.** Anything beyond unit tests — booting a service, calling an endpoint, checking a migration, clicking through the web UI — is done with the whole stack up: Postgres, budget-server, budget-ingest, the web app and the edge proxy.
- **Use local builds of what you changed.** `make start-local` in budget-stack (added in task P0-X1) builds images from the sibling checkouts and runs them in place of the pinned releases. `make start` runs the pinned releases only.
- **One exception:** tasks P0-I1 to P0-I3 run before the stack knows about budget-ingest, so they use the repo's own `make serve-local`. Everything from P0-X1 onward uses the stack.
- **Unit tests stay in their repo** (`go test ./...`, `pnpm test`) and must not need the stack. They are necessary but not sufficient: a task that changes behaviour across a service boundary is not done until it has been exercised in the stack.
- **When a task spans repos,** check out each repo's branch side by side and run them together in the stack before opening any of the PRs. Say in each PR which branches of the other repos it was run against.
- **Do not create another multi-service setup.** The per-repo compose files exist to match budget-server's structure; they are not where integration is checked. If the stack is missing something a task needs, add it to budget-stack as part of that task.
- **Record what you ran.** Each PR description lists the stack commands used and what was observed (for example the request made and the log lines or rows that resulted).

## Fixed shapes

Both API surfaces are defined before any behaviour is built behind them, and later phases should not need to change them.

- **Lunch Money side (budget-ingest):** endpoints, wire structs and validation are built in phase 1 and corrected against real traffic in phase 2. Later phases change what happens behind a handler, not its shape.
- **budget-server side:** the whole `LunchSyncService` and every new field are defined in phase 3. budget-ingest (phase 3) and the web UI (phase 7) generate clients from budget-server's protos with buf, pinned to a release tag.

If a later task does need a shape or proto change, keep it additive, say so at the top of the PR, and for protos expect a new budget-server tag and a repin before dependent tasks can merge.

## Releases and deploys

Every phase ends with the human merging its PRs, tagging releases, and bumping the pinned versions in budget-stack and production. This is implied and not listed as a task in each phase. While a phase is in progress, run unreleased work together with `make start-local`.

## Patterns to follow in budget-server

| Need | Copy the shape of |
|---|---|
| Sync provider | `internal/sync/providers/plaid.go`, registered in `main.go` |
| Action + test | `internal/operator/actions/plaid_add_transaction.go` and its `_test.go` |
| Connect handler | `internal/connecthandlers/v1/plaid/`, registered in `api/connect.go` |
| Storage reader/writer | `internal/storage/plaid/` |
| Migration | `scripts/db_migrations/migrations/000005_add_plaid_tables.*.sql` |
