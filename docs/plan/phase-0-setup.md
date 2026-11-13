# Phase 0: setup

> **Abandoned 2026-11-13.** The FinanceKit entitlement was approved, so this LunchSync bridge will not be built. Kept for reference only; do not implement it.

[Overview](README.md) · [Working rules](working-rules.md) · [API contract](api-contract.md) · Next: [Phase 1](phase-1-external-endpoints.md)

Read the working rules before starting any task here.

**Outcome:** budget-ingest is an empty but complete Go service that builds, tests, logs and releases exactly the way budget-server does. It runs in budget-stack beside budget-server and the web app, and it is deployed in production behind Traefik on `api.lunchmoney.dev` with the private certificate. All it does is answer `/healthz`.

## Progress

Change `[ ]` to `[x]` when a task's PR is merged, or for human tasks when the step is done.

| Done | Task | What | Repo | Who |
|---|---|---|---|---|
| [ ] | P0-I1 | Go module and service skeleton | budget-ingest | Agent |
| [ ] | P0-I2 | Build, local run and Makefile | budget-ingest | Agent |
| [ ] | P0-I3 | CI and release workflows | budget-ingest | Agent |
| [ ] | P0-H1 | First push and first release | — | Human |
| [ ] | P0-X1 | Add budget-ingest to budget-stack | budget-stack | Agent |
| [ ] | P0-X2 | Production wiring and runbooks | infra-prod, budget-ingest | Agent |
| [ ] | P0-H2 | Run the stack, create the certificate, deploy | — | Human |
| [ ] | — | Exit check passed | — | Human |

The reference is `../budget-server` at its current default branch. The rule for the budget-ingest tasks is **copy, then rename**: start from budget-server's file, change only names, ports and what is listed under "Deliberate differences". If something in budget-server looks improvable, leave it the same and note it in the PR.

## What "the same" means

| Area | budget-server reference | budget-ingest |
|---|---|---|
| Language and toolchain | `mise.toml` (Go version), `go.mod` `tool` block (mockery, goimports) | Same versions, same tools |
| Entry point | root `main.go`: set up logging, read config, build dependencies, start the server | Same shape |
| HTTP server | `api/connect.go`: a struct with `Serve()`, fixed port, the same four timeouts, `LoggingMiddleware` wrapping the mux | `api/server.go`, plain `net/http` mux, port `9448` |
| Config | `internal/config/config.go`: `ProcessEnvironmentVariables()` with defaults that suit docker compose | Same function name and style |
| Logging | `internal/logging/` (`logging.go`, `log_data.go`, `logging_wrapper.go`) and `LoggingMiddleware` in `api/middleware.go` | Copied unchanged apart from the import path |
| Dependencies | Vendored for the image build: `go mod vendor`, `/vendor` git-ignored, `go build -mod=vendor` | Same |
| Image | `Dockerfile`: `golang:<ver>-alpine` build stage, `alpine` runtime | Same, minus the migration stage until phase 4 |
| Local run | `docker-compose.yaml` with `serve` and `serve-local` profiles | Same profiles |
| Make targets | `build`, `clean`, `vendor`, `serve`, `serve-local`, `serve-local-rebuild`, `format`, `generate-mocks`, `help` | Same names and behaviour |
| Tests | `go test ./...`, testify, mockery-generated mocks beside the interface (`.mockery.yaml`) | Same |
| CI | `.github/workflows/test.yaml`, `build.yaml`, `renovate.yml` | Same triggers, steps and action versions |
| Dependency updates | `renovate.json` | Identical |
| Releases | A GitHub Release with tag `vX.Y.Z`; `build.yaml` publishes `ghcr.io/carson-networks/<repo>` tagged with the version, the branch, the commit SHA, and `latest-release` on tags | Same, image `ghcr.io/carson-networks/budget-ingest`, first release `v0.1.0` |
| Default branch | `master` | `master` |

### Logging conventions to keep

- logrus with the JSON formatter, level key renamed to `loglevel`, output to stdout, level info.
- One `LogData` per request, carried on the context; handlers add fields and timings to it.
- The middleware emits `Handler.Start` and `Handler.Complete`, with `method`, `path`, `status` and `durationMs`.
- Message names follow `Package.Function.detail` (for example `ConnectServer.Serve.listening`).
- Startup line: `budget-ingest starting`.

### Deliberate differences

Only these. Anything else that differs is a defect in the PR.

- **`.gitignore` must not ignore `docs/`.** budget-server's does; here the plan lives in `docs/plan/`.
- **Image name in compose and the Makefile is `ghcr.io/...`.** budget-server has it misspelled as `ghrc.io` in both.
- **No `CorsMiddleware`.** Nothing calls this service from a browser.
- **No `buf.yaml` workflow.** This repo defines no protos; it only generates clients (added in phase 3).
- **No database pieces yet.** The migration runner, its Dockerfile stage, bob codegen and the `migrate`/`generate` make targets arrive in phase 4 with the first table, copied from budget-server at that point.
- **No `.iml` file.**

## Tasks

### P0-I1 Go module and service skeleton

- `go.mod` for `github.com/carson-networks/budget-ingest` with the same Go version and `tool` block as budget-server; `mise.toml` copied.
- `internal/logging/` copied from budget-server.
- `internal/config/config.go` with `ProcessEnvironmentVariables()`. No variables are needed yet; keep the struct and the function so phase 3 only adds fields.
- `api/middleware.go` with `LoggingMiddleware` and the `responseWriter` it uses; `api/server.go` with a server struct and `Serve()` mirroring `ConnectServer`.
- One route, `GET /healthz`, returning `200 ok`.
- `main.go` mirroring budget-server's.
- `.gitignore`, `.mockery.yaml` (empty `packages` for now), `renovate.json`.
- Root `README.md`: what the service is, how to run it locally, and a link to `docs/plan/README.md`.
- `docs/dev-rules.md`, adapted from `budget-server/plan/dev-rules.md`, plus the "Running and testing locally" rules from this plan's working rules: run and test through budget-stack.

Acceptance:
- `go test ./...` passes, with tests for the logging middleware (fields present, status captured) and `/healthz`.
- A diff of `internal/logging/` against budget-server shows only import-path changes.
- A request to `/healthz` produces `Handler.Start` and `Handler.Complete` JSON lines with `loglevel`, `method`, `path`, `status` and `durationMs`.

### P0-I2 Build, local run and Makefile

- `Dockerfile` following budget-server's build and runtime stages; the command starts `/budget-ingest`.
- `docker-compose.yaml` with `serve` (published image) and `serve-local` (local build) profiles, publishing port `9448`. No Postgres service yet.
- `Makefile` with the targets listed in the table above, same help output style.

Acceptance: `make serve-local` builds and starts the service and `curl localhost:9448/healthz` answers; `make format` leaves the tree unchanged; `make help` lists every target.

### P0-I3 CI and release workflows

- `.github/workflows/test.yaml`, `build.yaml` and `renovate.yml` copied from budget-server with the image name changed. Keep the action versions identical so Renovate groups them the same way in both repos.
- `build.yaml` keeps the same tag set (`type=sha`, branch, tag, `latest-release` on tags), the `go mod vendor` step before the build, and the provenance attestation.
- Add a "Releasing" section to the root README describing the process: merge to `master`, create a GitHub Release with tag `vX.Y.Z`, the image is published by `build.yaml`.

Acceptance: workflow files differ from budget-server's only in the image name; `actionlint` (or a dry read) finds no errors. The agent cannot observe the first CI run on an empty remote; say so in the PR.

### P0-H1 Human: First push and first release

1. Decide the default branch. This plan assumes `master` to match budget-server; the fresh clone is on `main`.
2. Push the first commit (the plan) and merge the three PRs.
3. Confirm the Test and Build workflows are green and the image appears at `ghcr.io/carson-networks/budget-ingest`. Set the package's visibility to match budget-server's.
4. Give the Renovate workflow the same repository permissions it has in budget-server.
5. Create release `v0.1.0`.

### P0-X1 Add budget-ingest to budget-stack

Do this after `v0.1.0` exists. Repo: `../budget-stack`. From here on, budget-stack is where the whole system is run and tested together. All the wiring is put in place now, so later phases only bump image versions.

- `compose.yaml`: add a `budget-ingest` service.
  - Image `ghcr.io/carson-networks/budget-ingest:v0.1.0`, pinned by version like its neighbours so Renovate bumps it.
  - Same hardening as `budget-server`: `restart: unless-stopped`, `read_only: true`, `cap_drop: ALL`, `no-new-privileges`.
  - Networks `default` (to reach `budget-server:9447`) and `db`; `depends_on` Postgres healthy.
  - Environment: `BUDGET_SERVER_URL`, `POSTGRES_ADDRESS`, `POSTGRES_PORT`, `POSTGRES_DB=budget_ingest`, `POSTGRES_USERNAME=budget_ingest`, `POSTGRES_PASSWORD` from a new `BUDGET_INGEST_POSTGRES_PASSWORD` in `.env.example`. The phase 0 image ignores them; they are used from phase 3 and 4.
  - Published as `127.0.0.1:9448:9448`. It is not routed through `edge`: the phone-facing API answers on its own hostname, and the edge Caddy's `SITE_HOSTS` check would reject it.
  - A healthcheck on `/healthz`, in the style of the `edge` service's.
- Database: an init script mounted into the Postgres container that creates the `budget_ingest` database and a `budget_ingest` role that owns it and has no privileges on the budget database. Document in the README the one-off command for a stack whose volume already exists (init scripts only run on an empty volume).
- Local-build mode, so changes can be run together before anything is released:
  - `compose.local.yaml`, an override that gives `budget-server`, `budget-ingest` and `web` a `build:` context pointing at the sibling checkouts (`../budget-server`, `../budget-ingest`, `../budget`) and a `:local` image tag.
  - Make targets `start-local` (build from the sibling checkouts, then start with the override) and `rebuild-local` (rebuild and restart one service, `make rebuild-local s=budget-ingest`). The Go images build with `-mod=vendor`, so the target runs `go mod vendor` in each Go checkout first.
  - A missing sibling checkout fails with a message naming the path, not a Docker error.
- `README.md`: add budget-ingest to the description of what the stack runs, with its port and the note that it is separate from the edge proxy; add a "Developing across repos" section explaining `start` versus `start-local`.
- Leave `Caddyfile`, the `edge` service and the existing services unchanged.

Acceptance:
- `docker compose config` validates.
- On a fresh volume, `make start` brings up every service healthy, the `budget_ingest` database and role exist, `curl 127.0.0.1:9448/healthz` answers, and the web app at `http://localhost:8080` behaves as before.
- The documented one-off command creates the database on an existing volume.
- `make start-local` runs the three services from local source: a visible change in a sibling checkout shows up after `make rebuild-local`.

### P0-X2 Production wiring and runbooks

Repos: `../infra-prod` and this one. Prepare everything production needs so that later phases only change image versions and one environment variable. Do not create certificates or touch the running system.

- `infra-prod/app-prod/budget/docker-compose.yml`: add a `budget-ingest` service at `v0.1.0`, hardened like its neighbours, on the `proxy`, `default` and `db` networks, with the same environment variables as in budget-stack.
- Traefik labels on it: rule ``Host(`api.lunchmoney.dev`)``, entrypoint `https`, `tls=true`, **no** `certresolver`, service port `9448`. This is the only route to budget-ingest; budget-server stays reachable only through the edge.
- Traefik file provider: a `tls.certificates` entry pointing at a mounted certificate and key path for `api.lunchmoney.dev`.
- `docs/plan/runbook-ca.md`: exact `openssl` commands for a root CA name-constrained to DNS `api.lunchmoney.dev`; a leaf for that name with SAN, `serverAuth` EKU and 825-day validity; exporting the root as a profile installable on iOS; where each file goes and which never leave offline storage.
- `docs/plan/runbook-deploy.md`: SQL to create the `budget_ingest` database and role in production; where the certificate and key are mounted (key mode `0600`, kept out of git); deploy order (budget-server, then budget-ingest, then web); how to check the route without DNS, using `curl --resolve api.lunchmoney.dev:443:<traefik-ip> --cacert root.pem https://api.lunchmoney.dev/healthz`; rollback (remove the DNS override; the app then reaches the real Lunch Money and fails auth harmlessly).

Acceptance: `docker compose config` validates in infra-prod; both runbooks can be followed by pasting commands; no secret, key or certificate is committed.

### P0-H2 Human: Run the stack, create the certificate, deploy

1. Merge the budget-stack PR; run `make start` and `make start-local` on a real checkout.
2. Run the CA runbook. Keep the root key offline. Set a calendar reminder a month before the leaf expires.
3. Run the deploy runbook: create the database and role, mount the leaf certificate, deploy.
4. Check the route with the `curl --resolve` command from the runbook.

## Exit check

- A push to a branch produces an image tagged with the branch name and SHA; release `v0.1.0` produces one tagged `v0.1.0` and `latest-release`.
- The image logs `budget-ingest starting` as a JSON line and answers `/healthz`.
- In budget-stack, `make start` runs budget-ingest beside the existing services and nothing that worked before is broken.
- In production, `https://api.lunchmoney.dev/healthz` answers over the private certificate when resolved to Traefik.
