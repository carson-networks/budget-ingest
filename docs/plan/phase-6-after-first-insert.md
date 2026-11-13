# Phase 6: after the first insert

[Overview](README.md) · [Working rules](working-rules.md) · [API contract](api-contract.md) · Previous: [Phase 5](phase-5-transactions.md) · Next: [Phase 7](phase-7-web-ui.md)

Read the working rules before starting any task here.

**Outcome:** the data stays correct as things change. A pending charge takes its final amount when it settles, a rejected charge is removed, balances match Wallet, and budget-server records when each account last synced.

## Progress

Change `[ ]` to `[x]` when a task's PR is merged, or for human tasks when the step is done.

| Done | Task | What | Repo | Who |
|---|---|---|---|---|
| [ ] | P6-S1 | Pending, updates, deletes, balance, last-synced | budget-server | Agent |
| [ ] | P6-I1 | Updates, pending, rejected, balances | budget-ingest | Agent |
| [ ] | — | Exit check passed | — | Human |

The pending flag and last-synced time are stored and returned by the API in this phase; they become visible in the web UI in phase 7. Check them here through the API or the database.

## Tasks

### P6-S1 Pending, updates, deletes, balance, last-synced

- Migration `000009` (confirm the next free number):
  ```sql
  ALTER TABLE transactions ADD COLUMN pending BOOLEAN NOT NULL DEFAULT FALSE;
  ALTER TABLE syncs ADD COLUMN last_synced_at TIMESTAMPTZ;
  ```
- Carry `pending` through the transaction model, insert, update and list, and populate `Transaction.pending` and `NewTransaction.pending` (fields defined in phase 3).
- Populate `Account.last_synced_at` (defined in phase 3) from `syncs.last_synced_at`.
- Implement the remaining `LunchSyncService` RPCs, replacing their `Unimplemented` responses. No proto changes.
  - `UpdateTransaction`: partial update (amount, name, merchant, date, category, needs_review, pending). Absent fields are untouched.
  - `DeleteTransaction`
  - `SetAccountBalance(account_id, balance)`
- All three verify the target belongs to a LunchSync account. `CreateTransactions`, `UpdateTransaction`, `DeleteTransaction` and `SetAccountBalance` each set `syncs.last_synced_at`.
- Actions with tests: `lunchsync_update_transaction.go`, `lunchsync_delete_transaction.go`, `lunchsync_set_account_balance.go`.

Acceptance: migration applies and rolls back; tests for partial update semantics, scoping, and `last_synced_at` advancing; existing Plaid and manual behaviour unchanged (all existing tests pass untouched).

### P6-I1 Updates, pending, rejected, balances

- `mapping`: `custom_metadata.wallet.status` → `pending` per the contract's "Pending and rejected" table. Absent metadata → `pending = false` on insert, untouched on update.
- `POST /v2/transactions`: send `pending`. A `rejected` item is not sent to budget-server; its mapping row is created with `removed_at` set and it is returned as inserted.
- `PUT /v2/transactions/{id}`: partial `UpdateTransaction`. `rejected` → `DeleteTransaction` and set `removed_at`. A missing `category_id` leaves the category alone.
- A tombstoned transaction pushed again is reported under `skipped_duplicates`; an update to one returns success and does nothing.
- `PUT /v2/manual_accounts/{id}`: `SetAccountBalance`, value passed through unchanged unless the P2 results table says otherwise; in that case stop and ask how to transform it.
- No endpoint returns `501` any more. Keep `STUB_MODE` for local testing.
- When budget-server answers `NotFound`, delete the stale mapping row and return `404` naming what was deleted in the web UI.

Acceptance: mapping tests for every status value and for missing metadata; tests for pending → booked, rejected on insert, rejected on update, re-push of a tombstone, and stale mapping cleanup; contract tests on the P2 update and balance fixtures.

## Exit check

- A new purchase is stored with `pending` true; after it settles the same transaction has its final amount and `pending` false.
- Balances for each account in the web UI match the Wallet app.
- `last_synced_at` advances after a background sync with the app closed.
- If a rejected charge occurs, it disappears from the web UI.
- No endpoint returns `501`.
