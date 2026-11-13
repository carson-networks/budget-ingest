# Phase 5: transactions logic

[Overview](README.md) · [Working rules](working-rules.md) · [API contract](api-contract.md) · Previous: [Phase 4](phase-4-accounts.md) · Next: [Phase 6](phase-6-after-first-insert.md)

Read the working rules before starting any task here.

**Outcome:** a sync from the app puts transactions in budget-server with the right sign, date and name, and syncing again adds nothing.

## Progress

Change `[ ]` to `[x]` when a task's PR is merged, or for human tasks when the step is done.

| Done | Task | What | Repo | Who |
|---|---|---|---|---|
| [ ] | P5-S1 | Idempotent batch insert | budget-server | Agent |
| [ ] | P5-I1 | Transaction insert and read | budget-ingest | Agent |
| [ ] | — | Exit check passed | — | Human |

## Tasks

### P5-S1 Idempotent batch insert

- Storage: a transaction insert that uses `ON CONFLICT (id) DO NOTHING` and reports whether a row was written.
- Action `lunchsync_add_transactions.go`: batch insert for one account, returning created and existing IDs.
- Implement `LunchSyncService.CreateTransactions` (messages defined in phase 3), replacing its `Unimplemented` response. `pending` is accepted but not stored until phase 6.
- `FailedPrecondition` when the account is not LunchSync-type; `NotFound` when it is gone; `InvalidArgument` for an unknown category. The whole batch is one database transaction.

Acceptance: tests for a repeated batch (second call: all `existing_ids`, no row changed), a mixed batch, scoping (manual and Plaid accounts refused), and a 500-item batch finishing well inside the 30 s write timeout.

### P5-I1 Transaction insert and read

- Migration:
  ```sql
  CREATE TABLE transaction_ids (
      lm_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
      transaction_id UUID NOT NULL UNIQUE,   -- uuidv5(account_id, external_id)
      account_lm_id BIGINT NOT NULL REFERENCES account_ids(lm_id) ON DELETE CASCADE,
      external_id TEXT NOT NULL,
      removed_at TIMESTAMPTZ,
      UNIQUE (account_lm_id, external_id)
  );
  ```
- `internal/mapping`: amount negated; `payee` → name; `custom_metadata.wallet.merchantName` → merchant; `date` → UTC midnight; `category_id` through `category_ids`; `needs_review` per the contract's status table. Fix one UUIDv5 namespace constant and never change it.
- `POST /v2/transactions`:
  1. Validate the whole batch; any bad item → `400` with one `errMsg` per bad index, nothing written. Items without `manual_account_id` are bad items.
  2. Per account: compute each UUID and upsert its `transaction_ids` row **before** calling budget-server.
  3. One `CreateTransactions` call per account.
  4. Response per the contract: created → `transactions`; existing → `skipped_duplicates` with `reason: "duplicate_external_id"`, the request index, the existing integer ID and the echoed `external_id`. Amounts in the response use LunchSync's sign.
- `GET /v2/transactions` and `/{id}`: mapped transactions only; date range, `manual_account_id`, `limit`/`offset`, `has_more`.
- Raise the request body limit to 8 MB for this route.
- Log one line per request: key ID, account, counts of created / duplicate / failed. No payee or amount at info level.

Acceptance:
- Mapping unit tests, including sign and missing optional fields.
- Contract tests on the P2 insert fixtures.
- Idempotency: same batch twice → second response is all `skipped_duplicates` with the first run's IDs.
- Crash safety: with a mapping row present but no transaction in budget-server, a retry creates it and reports it as inserted.
- End to end in budget-stack: fixture in, transactions visible through budget-server's `TransactionService.ListTransactions`.

## Exit check

- Sync one Wallet account with a short history window: the count in the web UI's transaction list matches the app, purchases are negative, dates and names are right.
- Sync again: nothing new.
- Change a category in the web UI, sync again: the category is unchanged.
