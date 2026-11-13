# LunchSync sync: plan

> **Abandoned 2026-11-13.** The FinanceKit entitlement was approved, so this LunchSync bridge will not be built. Kept for reference only; do not implement it.

Status: abandoned · 2026-11-13 (proposed 2026-11-11)

## Goal

Get Apple Card, Apple Cash and Apple Savings transactions into budget-server without building an iOS app or obtaining the FinanceKit entitlement. The unmodified [LunchSync](https://github.com/milanave/LunchSync) app reads FinanceKit on the phone and pushes to the Lunch Money API. A new service, **budget-ingest**, answers as that API and forwards what arrives to budget-server over ConnectRPC.

This replaces `budget-stack/docs/apple-sync-plan.md`.

```
┌────────── iPhone ──────────┐      ┌───────────────────── home server ─────────────────────┐
│ Wallet ─► FinanceKit       │      │ Traefik: Host(api.lunchmoney.dev), private-CA cert     │
│             │              │ HTTPS│        │                                               │
│  LunchSync (App Store) ────┼─────►│        ▼                                               │
│  Bearer <API key>          │ VPN  │ budget-ingest ──ConnectRPC──► budget-server            │
└────────────────────────────┘      │   │  Lunch Money v2 API         LunchSyncService       │
      DNS override:                 │   ▼                                 ▼                  │
      api.lunchmoney.dev → Traefik  │ Postgres db: budget_ingest    Postgres db: budget      │
                                    └────────────────────────────────────────────────────────┘
```

Data only flows phone → server. Nothing on the server can ask the phone to sync; LunchSync's own background push decides when that happens.

## How it works for the user

1. In the web UI, **Add account → LunchSync account**. Enter a name and pick Apple Card, Apple Cash or Apple Savings. This creates an account of the LunchSync sync type.
2. The UI shows **one-time setup instructions**: VPN, trusting the certificate, generating the API key from budget-ingest, and the LunchSync app settings. Adding a second or third account skips this; the instructions stay reachable from the menu.
3. In the LunchSync app, **pair each Wallet account** with a budget account. The list offered contains only LunchSync-type accounts. Creating the account from the app instead also works; it is created as a LunchSync-type account.
4. The app **syncs transactions and balances** into budget-server, in the background from then on.

## Components

| Piece | Repo | Role |
|---|---|---|
| budget-ingest | `carson-networks/budget-ingest` | Speaks Lunch Money v2 to the phone, ConnectRPC to budget-server. Owns the API key and the integer-ID mappings in its own database. |
| `LunchSyncService` | budget-server | New ConnectRPC service with LunchSync-specific account and transaction RPCs. New sync type with a no-op provider. Pending flag on transactions. |
| Web UI | budget | LunchSync account creation, setup instructions, last-synced and pending display. |
| Routing, certificate, DNS | infra-prod | Traefik router for the spoofed host, private CA, DNS override, database and role for budget-ingest. |

## Design decisions

1. **Target Lunch Money API v2 only.** One spoofed host, and dedup is a single insert call.
2. **Separate repo and service.** budget-ingest is the only thing reachable on the spoofed host. budget-server stays behind the edge and has no auth, by intent; budget-ingest reaches it on the internal Docker network only.
3. **Dedicated RPCs, not the existing ones.** `LunchSyncService` has its own `CreateAccount` and `CreateTransactions`, scoped to LunchSync-type accounts.
4. **Idempotency by client-supplied ID.** budget-ingest derives each transaction's UUID deterministically (UUIDv5 of account ID and Apple's transaction ID). `CreateTransactions` reports per item whether it was created or already existed. budget-server never stores an external ID.
5. **Separate database on the same Postgres instance** (`budget_ingest`) with its own role and migrations. No foreign keys into budget tables.
6. **Accounts can be created from the web UI or from the app.** Both go through `LunchSyncService.CreateAccount`, so either way the result is a LunchSync-type account.
7. **No-op sync provider.** Required: `Orchestrator.Sync` errors on any account whose sync type has no registered provider, which would break Plaid sync-all too.
8. **Last-synced time lives in budget-server**, updated by every write RPC, so the web UI can show it.
9. **Pending is shown.** budget-ingest reads the Apple status from the metadata LunchSync sends; budget-server stores a `pending` flag; the UI badges it.
10. **Rejected charges are removed automatically** when the metadata says so. Charges that simply vanish from Wallet are deleted by hand.
11. **Balance is passed through unchanged** and checked against the Wallet app during the spike.
12. **API key comes from a small command in the budget-ingest image** (`/budget-ingest-key create`), printed once and stored hashed. The web UI only shows the instruction. (Alternative if this is too manual: an edge route so the UI can request a key.)
13. **Name-constrained private CA**, limited to `api.lunchmoney.dev`, so trusting it on the phone can't be used to intercept anything else.

## What was verified in the LunchSync source

Read at commit `7ec2cba` (2026-07-25).

| Fact | Where |
|---|---|
| No certificate pinning: plain `URLSession.shared` | `Api/LunchMoneyAPI.swift`, `Api/LunchMoneyAPIV2.swift` |
| Host is hard-coded, no base-URL setting | `https://api.lunchmoney.dev/v2` (default), `https://dev.lunchmoney.app/v1` (legacy) |
| `Authorization: Bearer <token>` on every call | both clients |
| v2 expects server-side dedup on `(manual_account_id, external_id)` | `createResponse(fromV2:)` |
| Inserts are batched, up to 500 per request | `SyncBroker.createBatchSize` |
| Debits are positive, credits negative | `Api/AppleWallet.swift` |
| Balance is Apple's raw amount, no sign applied | `Api/AppleWallet.swift` |
| Any status other than `booked` is treated as pending | `Api/AppleWallet.swift` |
| Apple's status travels only in `custom_metadata.wallet.status`, and only when the app's metadata setting is on | `Models/WalletMetadata.swift`, `SyncBroker.walletMetadata` |
| The v2 client never deletes transactions | `LunchMoneyAPIV2` |

Not verified: behaviour on a real device, and whether the app sends an update when a charge becomes `rejected`. Phase 2 covers both.

## Phases

Do not start a phase until the previous one's exit check has passed. Each phase ends with a release and deploy, which is implied and not listed as a task.

| Done | # | Phase | Outcome |
|---|---|---|---|
| [ ] | 0 | [Setup](phase-0-setup.md) | budget-ingest mirrors budget-server, runs in budget-stack, and is deployed behind Traefik with the private certificate |
| [ ] | 1 | [External endpoints, stubbed](phase-1-external-endpoints.md) | budget-ingest answers every endpoint the app calls with its final shape; none implemented |
| [ ] | 2 | [Phone setup and test](phase-2-phone-test.md) | The real app syncs against stub mode; shapes corrected against real traffic. Go/no-go |
| [ ] | 3 | [budget-server endpoints, synced to ingest](phase-3-budget-server-endpoints.md) | Every `LunchSyncService` RPC and new field exists and budget-ingest can reach them; none implemented |
| [ ] | 4 | [Accounts logic](phase-4-accounts.md) | Accounts created from the app exist in budget-server; only LunchSync accounts are listed; a key is required |
| [ ] | 5 | [Transactions logic](phase-5-transactions.md) | A sync inserts transactions correctly; a second sync adds nothing |
| [ ] | 6 | [After the first insert](phase-6-after-first-insert.md) | Settled amounts, pending flag, rejected charges removed, balances, last-synced |
| [ ] | 7 | [Web UI](phase-7-web-ui.md) | Add LunchSync account, setup instructions, label, last-synced, pending badge |
| [ ] | 8 | [Hardening](phase-8-hardening.md) | A week of unattended background sync, with backups and alerting |
| [ ] | 9 | [Follow-ups](phase-9-follow-ups.md) | Unscheduled |

The order is deliberate: wiring first, then both API surfaces as empty shells with the phone test in between, then behaviour one slice at a time, and the web UI last so no UI is built for something that isn't proven. Until phase 7, LunchSync accounts are created from the app and show as "Manual" in the web UI.

Each phase file opens with a progress table listing its tasks; tick the phase here when its exit check has passed.

Supporting files:

- [Working rules](working-rules.md): what the implementing agent must follow in every task — repos, checks, human-only steps, patterns to copy.
- [API contract](api-contract.md): what LunchSync sends and what it must get back.

This plan and its runbooks live in `budget-ingest/docs/plan/`.

## Risks

| Risk | Mitigation |
|---|---|
| A LunchSync update adds pinning or changes endpoints | Turn off auto-update for the app; contract tests mirror the Swift structs; the stale-sync indicator shows when it stops. |
| DNS override doesn't apply (cellular, Private Relay, encrypted DNS) | Tested in phase 2; on-demand VPN carries the override away from home. |
| Metadata setting turned off in the app | Sync still works; pending and rejected handling silently stop. Called out in the setup instructions. |
| Version coupling between budget-ingest and budget-server | budget-ingest pins a proto tag; deploy budget-server first. |
| Certificate expires (825 days at most on iOS) | Calendar reminder and a renewal runbook. |
| No licence on the LunchSync repo | We run the App Store build unmodified and copy none of its code. |
