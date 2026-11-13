# Phase 2: phone setup and test

[Overview](README.md) · [Working rules](working-rules.md) · [API contract](api-contract.md) · Previous: [Phase 1](phase-1-external-endpoints.md) · Next: [Phase 3](phase-3-budget-server-endpoints.md)

Read the working rules before starting any task here.

**Outcome:** the unmodified LunchSync app completes a sync against budget-ingest in stub mode, the device-only questions are answered, and the endpoint shapes are corrected against real traffic. If the app will not connect, stop: the rest of the plan is void.

## Progress

Change `[ ]` to `[x]` when a task's PR is merged, or for human tasks when the step is done.

| Done | Task | What | Repo | Who |
|---|---|---|---|---|
| [ ] | P2-X1 | Turn on stub mode in production | infra-prod | Agent |
| [ ] | P2-H1 | Trust, DNS, VPN and device tests | — | Human |
| [ ] | P2-I1 | Reconcile the API with what the phone sent | budget-ingest | Agent |
| [ ] | — | Exit check passed | — | Human |

This is the go/no-go. Nothing in budget-server or the web UI has been touched yet.

## Tasks

### P2-X1 Turn on stub mode in production

Repo: `../infra-prod`.

- Bump `budget-ingest` to the phase 1 release and set `STUB_MODE=true` on it, with a comment that it is removed in phase 4.

Acceptance: `docker compose config` validates.

### P2-H1 Human: Trust, DNS, VPN and device tests

1. Deploy P2-X1. The certificate and route are already in place from phase 0.
2. Add the DNS override `api.lunchmoney.dev` → Traefik on the home resolver.
3. Install and trust the root on the iPhone (Settings → General → About → Certificate Trust Settings).
4. In LunchSync: API v2, FinanceKit metadata on, any string as the token. Pair one Wallet account with a fake account and sync.
5. Set up the VPN with on-demand and the DNS override, and repeat off Wi-Fi.
6. Fill in the results table below and save the stub's logged request bodies to `budget-ingest/internal/lunchmoney/testdata/` (redact merchant names and amounts if desired).

### P2-I1 Reconcile the API with what the phone sent

After P2-H1. The phase 1 shapes were written from the app's source; this checks them against real traffic.

- Add the captured request bodies as fixtures under `internal/lunchmoney/testdata/` and a test that decodes each into its wire struct with unknown fields disallowed.
- For every mismatch (a field the app sends that the struct lacks, a type that differs, an endpoint or query parameter not in the contract), fix the wire structs and update [api-contract.md](api-contract.md) in the same PR.
- If the traffic shows data budget-server will need that the plan's `LunchSyncService` messages lack, record it in the contract so phase 3 defines them correctly.
- Record the answers from the results table that later tasks depend on (balance meaning and sign, pending lifecycle, rejected updates) in the contract.

Acceptance: every captured request decodes cleanly; the contract and the structs agree; the PR lists each mismatch found, or states there were none.

## Exit check

| Question | Answer |
|---|---|
| Connects on home Wi-Fi | |
| Connects with iCloud Private Relay on | |
| Connects on cellular without VPN | |
| Connects over VPN, including a background sync | |
| Balance sent for Apple Card: amount owed or available credit; sign | |
| Balance sent for Cash and Savings matches Wallet | |
| A pending purchase arrives as an insert, then an update when it settles | |
| A rejected charge arrives as an update with `wallet.status: "rejected"` (leave blank if none occurred) | |

Go/no-go: ______
