# Phase 8: hardening

> **Abandoned 2026-11-13.** The FinanceKit entitlement was approved, so this LunchSync bridge will not be built. Kept for reference only; do not implement it.

[Overview](README.md) · [Working rules](working-rules.md) · [API contract](api-contract.md) · Previous: [Phase 7](phase-7-web-ui.md) · Next: [Phase 9](phase-9-follow-ups.md)

Read the working rules before starting any task here.

**Outcome:** every Apple account syncs in the background for a week with no duplicates and no manual steps, and a failure is noticed.

## Progress

Change `[ ]` to `[x]` when a task's PR is merged, or for human tasks when the step is done.

| Done | Task | What | Repo | Who |
|---|---|---|---|---|
| [ ] | P8-X1 | Operations | infra-prod, budget-stack | Agent |
| [ ] | P8-H1 | Go live on all accounts | — | Human |
| [ ] | — | Exit check passed | — | Human |

## Tasks

### P8-X1 Operations

- Add `budget_ingest` to the existing Postgres backup configuration.
- Uptime check on budget-ingest's `/healthz`.
- An alert when no LunchSync account has synced in 48 hours (expired certificate, broken DNS, dropped VPN, or an app update that changed the API all look like this).
- `budget-ingest/docs/plan/runbook-cert-renewal.md`: reissuing the leaf from the offline root.
- Replace `budget-stack/docs/apple-sync-plan.md` with a pointer to `budget-ingest/docs/plan/README.md`.

### P8-H1 Human: Go live on all accounts

- Take a database dump, pair the remaining Wallet accounts, enable background sync.
- Turn off App Store auto-update for LunchSync.
- Set a calendar reminder a month before the leaf certificate expires.

## Exit check

One week of background sync across all accounts: no duplicates, balances correct, no manual intervention.
