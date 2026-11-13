# Phase 7: web UI

[Overview](README.md) · [Working rules](working-rules.md) · [API contract](api-contract.md) · Previous: [Phase 6](phase-6-after-first-insert.md) · Next: [Phase 8](phase-8-hardening.md)

Read the working rules before starting any task here.

**Outcome:** LunchSync is a first-class integration in the web app. You can add a LunchSync account and follow the setup instructions from the browser, accounts are labelled "LunchSync" with their last sync time, and pending transactions are badged.

## Progress

Change `[ ]` to `[x]` when a task's PR is merged, or for human tasks when the step is done.

| Done | Task | What | Repo | Who |
|---|---|---|---|---|
| [ ] | P7-W1 | Regenerate web clients | budget | Agent |
| [ ] | P7-W2 | LunchSync accounts and setup instructions | budget | Agent |
| [ ] | P7-W3 | Pending badge, last synced, stale warning | budget | Agent |
| [ ] | — | Exit check passed | — | Human |

Everything here reads fields and calls RPCs that already work. No budget-server or budget-ingest changes are expected; if one seems necessary, stop and ask.

## Tasks

### P7-W1 Regenerate web clients

- Bump the proto ref in `buf.gen.yaml` (currently `v0.14.0`, well behind production) to the current budget-server release and run `pnpm generate`.
- List in the PR every generated change that does not come from this plan. Fix any compile breaks they cause without changing behaviour.
- `src/connectRPC/connect.ts`: add a `lunchSyncClient` singleton. `src/connectRPC/types.ts`: re-export the new messages.
- No model, query or component changes. The new wire fields are ignored for now.

Acceptance: `pnpm lint`, `pnpm test` and `pnpm build` pass; in budget-stack the app looks and behaves exactly as before.

### P7-W2 LunchSync accounts and setup instructions

- `src/models/account.ts`: add `LunchSync = 3` to `AccountIntegration`; replace the stubbed `integrationFromWireAccount` with a real mapping. Update `account_test.ts` and `src/test/wire.ts`.
- `src/queries/lunchSync.ts`: `lunchSyncMutations.createAccount`, invalidating the accounts list.
- `components/Account/lunchSyncAccountForm.ts` (pure): name, kind (Apple Card / Apple Cash / Apple Savings), starting balance → request, using the contract's type mapping.
- `components/Account/CreateLunchSyncAccountModal/Modal.tsx`: the form, then the setup instructions on success.
- `components/Account/LunchSyncSetup/Instructions.tsx`, a static checklist:
  1. VPN installed, on-demand enabled. *(one-time)*
  2. Root certificate installed and trusted. *(one-time)*
  3. API key generated with `docker exec budget-ingest /budget-ingest-key create <name>`. *(one-time)*
  4. LunchSync: API version v2, key pasted, FinanceKit metadata on. *(one-time)*
  5. Pair each Wallet account with its budget account (or create it from the app) and sync.

  When a LunchSync account already exists, steps 1–4 render collapsed.
- `AddAccountMenu.tsx`: "Add LunchSync account" and "LunchSync setup".
- `accountSyncLabel.ts`: `"LunchSync"`.
- Add a LunchSync account to the `?mock=true` demo data.

Acceptance: unit tests for the form logic and model mapping; component tests for the menu items, create → instructions, collapsed one-time steps, and the label; no CSP change.

### P7-W3 Pending badge, last synced, stale warning

- `src/models/transaction.ts`: `pending: boolean`. `src/models/account.ts`: `lastSyncedAt?: Date`. Update tests and `src/test/wire.ts`.
- Pending badge in `TransactionsTable` and `CategoryTransactionsTable`.
- Accounts table: last-synced time for LunchSync accounts, "Never" before the first sync, a warning when older than 48 hours.
- Confirm LunchSync transactions render correctly in `TransactionsTable` and `AccountTransactionsView`.
- Demo data: one LunchSync transaction; one pending transaction; one stale LunchSync account.

Acceptance: component tests for the badge, "Never", and the stale warning; a pure helper with unit tests decides staleness.

## Exit check

- Create "Apple Card" from the web UI; it appears in the app's pairing list.
- LunchSync accounts are labelled "LunchSync", not "Manual", and show their last sync time.
- A pending purchase is badged, and loses the badge after it settles.
- An account that has not synced for 48 hours shows the warning.
