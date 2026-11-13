# Phase 1: external endpoints, stubbed

[Overview](README.md) · [Working rules](working-rules.md) · [API contract](api-contract.md) · Previous: [Phase 0](phase-0-setup.md) · Next: [Phase 2](phase-2-phone-test.md)

Read the working rules before starting any task here.

**Outcome:** budget-ingest answers every endpoint the LunchSync app calls, with the exact request and response shapes, validation and error format. Nothing is implemented: in stub mode the endpoints return canned data, otherwise they return "not implemented". No database, no auth, no calls to budget-server.

## Progress

Change `[ ]` to `[x]` when a task's PR is merged, or for human tasks when the step is done.

| Done | Task | What | Repo | Who |
|---|---|---|---|---|
| [ ] | P1-I1 | Every Lunch Money endpoint, with full wire types | budget-ingest | Agent |
| [ ] | — | Exit check passed | — | Human |

Later phases fill in behaviour behind these handlers; they should not need to add an endpoint or change a shape.

## Endpoints

Shapes are in the [API contract](api-contract.md). This phase builds all of them; the last column says where the real behaviour arrives.

| Endpoint | Used for | Implemented in |
|---|---|---|
| `GET /v2/me` | Check the key at setup | Phase 4 |
| `GET /v2/manual_accounts` | List accounts to pair with | Phase 4 |
| `POST /v2/manual_accounts` | Create an account from the app | Phase 4 |
| `GET /v2/categories` | Category mapping screen | Phase 4 |
| `POST /v2/transactions` | Insert a batch, with dedup | Phase 5 |
| `GET /v2/transactions` | List by date range | Phase 5 |
| `GET /v2/transactions/{id}` | Fetch one | Phase 5 |
| `PUT /v2/transactions/{id}` | Update when a charge settles or changes | Phase 6 |
| `PUT /v2/manual_accounts/{id}` | Push a balance | Phase 6 |

The legacy v1 API (`dev.lunchmoney.app`) is out of scope.

## Tasks

### P1-I1 Every Lunch Money endpoint, with full wire types

- `internal/lunchmoney/wire.go`: a Go struct for every request and response in the contract, with JSON tags matching it exactly. Optional fields are pointers or `omitempty` so that an absent field stays absent.
- `internal/lunchmoney/handlers.go`: one handler per endpoint in the table above, routed under `/v2/` with the right method and path parameters. Each handler decodes and validates its request (required fields present, date and amount parseable, IDs numeric) and rejects bad input with `400`.
- Error helper producing the contract's error shape, `{"message": "...", "errors": [{"errMsg": "..."}]}`, used for every non-2xx response. Unknown routes under `/v2/` use it too, with `404`.
- Two behaviours behind the handlers, selected by the environment variable `STUB_MODE`:
  - **`STUB_MODE=true`**: every endpoint returns canned, valid data (two fake accounts, three fake categories, incrementing integer IDs for inserts, an echo for updates). This is what the phone is pointed at in phase 2.
  - **Unset or false** (the default): an endpoint that has no real implementation returns `501` in the error shape, naming the endpoint. It must never report success for work it did not do: the app marks transactions as synced on a 2xx and will not send them again.
- At startup, log one line listing which endpoints are implemented and which are not, and whether stub mode is on.
- Request logging for `/v2/` routes: add the query string to the request's `LogData`. In stub mode only, also log the full request body. Log whether an `Authorization` header was present, never its value.
- No auth, no database and no calls to budget-server yet.

Acceptance:
- A table-driven test covers every endpoint in both modes: stub mode returns a body that decodes into the response struct with every required field set; default mode returns `501` in the error shape.
- Round-trip tests: example JSON for each request and response in the contract decodes into its struct and encodes back without losing or inventing fields.
- Validation tests: malformed date, non-numeric amount and missing required fields each return `400`.
- In budget-stack with `STUB_MODE=true`, `curl` against each endpoint returns the canned response and the request appears in `make logs`.

## Exit check

- In budget-stack with `STUB_MODE=true`, every endpoint in the table answers with canned data and the request appears in the logs.
- With `STUB_MODE` unset, the same endpoints answer `501` and name the endpoint.
