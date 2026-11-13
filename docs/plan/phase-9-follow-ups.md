# Phase 9: follow-ups

> **Abandoned 2026-11-13.** The FinanceKit entitlement was approved, so this LunchSync bridge will not be built. Kept for reference only; do not implement it.

[Overview](README.md) · [Working rules](working-rules.md) · [API contract](api-contract.md) · Previous: [Phase 8](phase-8-hardening.md)

Unscheduled. Each is a separate piece of work to be specified before an agent picks it up.

| Done | # | Item |
|---|---|---|
| [ ] | 1 | Ask upstream (`milanave/LunchSync`) for a custom API base URL setting; if it lands, the private CA and DNS override go away. |
| [ ] | 2 | Charges that vanish from Wallet with no `rejected` update: flag pending transactions older than N days. |
| [ ] | 3 | Populate `transactions.pending` from Plaid, which already reads the flag. |
| [ ] | 4 | Keep `merchantCategoryCode` and feed it to default category rules (`budget-server/plan/catergories.md`). |
| [ ] | 5 | Generate the API key from the web UI through an edge route to budget-ingest. |
| [ ] | 6 | Other push sources in budget-ingest, starting with CSV/OFX statement upload. |
