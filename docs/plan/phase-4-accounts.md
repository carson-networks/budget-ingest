# Phase 4: accounts logic

> **Abandoned 2026-11-13.** The FinanceKit entitlement was approved, so this LunchSync bridge will not be built. Kept for reference only; do not implement it.

[Overview](README.md) · [Working rules](working-rules.md) · [API contract](api-contract.md) · Previous: [Phase 3](phase-3-budget-server-endpoints.md) · Next: [Phase 5](phase-5-transactions.md)

Read the working rules before starting any task here.

**Outcome:** accounts work end to end without the web UI. An account created from the app exists in budget-server as a LunchSync-type account, the app's pairing list shows only LunchSync-type accounts, real categories appear in the app's mapping screen, and the API requires a key.

## Progress

Change `[ ]` to `[x]` when a task's PR is merged, or for human tasks when the step is done.

| Done | Task | What | Repo | Who |
|---|---|---|---|---|
| [ ] | P4-S1 | Sync type, provider, and account RPC implementations | budget-server | Agent |
| [ ] | P4-I1 | Database and ID mappings | budget-ingest | Agent |
| [ ] | P4-I2 | Account and category endpoints | budget-ingest | Agent |
| [ ] | P4-I3 | API key and authentication | budget-ingest, infra-prod | Agent |
| [ ] | — | Exit check passed | — | Human |

The web UI is untouched until phase 7, so a LunchSync account shows there as "Manual" for now. To create one without the app, call `LunchSyncService.CreateAccount` directly (Connect accepts a JSON `POST`).

## Tasks

### P4-S1 Sync type, provider, and account RPC implementations

- `SyncType_LunchSync SyncType = 2` in `internal/storage/sync/model.go`.
- No-op provider `internal/sync/providers/lunchsync.go` (`Sync` returns an empty `SyncResult`), registered in `main.go`. This is required: `Orchestrator.Sync` errors on any account whose type has no provider.
- Action `lunchsync_create_account.go`: creates the account and its `syncs` row in one transaction and exposes the new ID.
- Implement `LunchSyncService.CreateAccount` and `ListAccounts` (defined in phase 3), replacing their `Unimplemented` responses. `ListAccounts` returns LunchSync-type accounts only.
- Populate `Account.integration` from the `syncs` row in `AccountService.ListAccounts`.
- No proto changes.

Acceptance:
- Tests: action; provider; a sync-all with one Plaid and one LunchSync account succeeds; `LunchSyncService.ListAccounts` excludes manual and Plaid accounts; `AccountService.ListAccounts` reports all three integrations correctly.
- Deleting a LunchSync account through the existing `DeleteAccount` works.
- No migration is needed in this task. If one seems necessary, stop and ask.

### P4-I1 Database and ID mappings

- Database tooling copied from budget-server, now that there is a first table: `scripts/db_migrations/` (golang-migrate runner, but do not print the connection string as budget-server's does), the `build-sql` Dockerfile stage with the image command running migrations then the server, `bobgen.yaml`, the `migrate` and `generate` make targets, and a Postgres service in the repo's own `docker-compose.yaml`.
- Config: `POSTGRES_ADDRESS|PORT|DB|USERNAME|PASSWORD` in `ProcessEnvironmentVariables()`, with docker-compose defaults.
- First migration:
  ```sql
  -- lm_id is the integer id LunchSync sees.
  CREATE TABLE account_ids (
      lm_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
      account_id UUID NOT NULL UNIQUE
  );
  CREATE TABLE category_ids (
      lm_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
      category_id UUID NOT NULL UNIQUE
  );
  ```
- `internal/storage/` with bob models, reader and writer following budget-server's shape: get-or-create a mapping by UUID, look up by `lm_id`, delete by UUID. Interfaces added to `.mockery.yaml`.
- `/healthz` now also checks the database and budget-server.

Acceptance: storage tests against Postgres for get-or-create (stable ID on repeat, concurrent callers get the same ID) and delete; in budget-stack the migration runs on start against the `budget_ingest` database created in phase 0.

### P4-I2 Account and category endpoints

Replace the `501` responses of these endpoints with real behaviour. Routes, wire structs and validation already exist from phase 1 and do not change. New package `internal/mapping` holds the pure conversions.

- `GET /v2/me`: canned user.
- `GET /v2/manual_accounts`: `LunchSyncService.ListAccounts`. Assign an `account_ids` row to any account seen for the first time; delete rows for accounts that no longer exist. Type mapping per the contract.
- `POST /v2/manual_accounts`: `LunchSyncService.CreateAccount` with the contract's type mapping and `balance` as the starting balance; assign an `account_ids` row; return the created account object.
- `GET /v2/categories?format=flattened`: `CategoryService.ListCategories`, assigning `category_ids` rows. Parents as `is_group: true`; children with `group_id`.
- Map budget-server errors onto the phase 1 error helper: `NotFound` → `404`, `Unavailable`/timeouts → `503`, `Unimplemented` → `501`. Never return `429`.
- The transaction and balance endpoints keep returning `501` until phases 5 and 6.

Acceptance:
- Contract tests replay the P2 fixtures; every response decodes with all required fields.
- Integer IDs are stable across restarts and repeated calls.
- An end-to-end check in budget-stack, with locally built images where releases don't exist yet: create via `POST`, see it in `GET`, and see it in budget-server's `AccountService.ListAccounts` as `LUNCHSYNC`.

### P4-I3 API key and authentication

Last in this phase, once the account endpoints work.

- Migration:
  ```sql
  CREATE TABLE api_keys (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      name TEXT NOT NULL,
      key_hash BYTEA NOT NULL UNIQUE,   -- sha256
      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      last_used_at TIMESTAMPTZ,
      revoked_at TIMESTAMPTZ
  );
  ```
- API key tool `scripts/api_key/main.go`, built into the image as `/budget-ingest-key` the way the migration runner is: `create <name>` (32 random bytes, base64url, printed once, sha256 stored), `list`, `revoke <id>`.
- Bearer middleware on everything under `/v2/`, in stub mode too: missing, unknown or revoked key → `401` in the contract's error shape; a valid key touches `last_used_at`. `/healthz` stays open.
- `GET /v2/me` returns the key's name as `api_key_label`.
- Never log a key or its hash.
- `infra-prod`: remove `STUB_MODE` from the production service.
- `docs/plan/runbook-deploy.md`: add generating a key with `docker exec budget-ingest /budget-ingest-key create <name>` and entering it in the app.

Acceptance: tests for the key lifecycle and each auth outcome; the tool's output works as a bearer token against the stack; a log search after the tests finds no key material.

## Exit check

- Without a key, or with a wrong or revoked one, the app is refused.
- Creating an account from the app creates it in budget-server; it shows in the web UI's account list.
- An account created with a direct `CreateAccount` call appears in the app's pairing list.
- A manual account and a Plaid account do **not** appear in the app.
- The app's category screen lists the real categories.
- Plaid sync still works with a LunchSync account present.
