# Phase 3: budget-server endpoints, synced to ingest

[Overview](README.md) · [Working rules](working-rules.md) · [API contract](api-contract.md) · Previous: [Phase 2](phase-2-phone-test.md) · Next: [Phase 4](phase-4-accounts.md)

Read the working rules before starting any task here.

**Outcome:** budget-server exposes the whole `LunchSyncService` and every new field, and budget-ingest has a generated client that can reach them. Nothing is implemented and nothing calls them yet: every RPC returns "unimplemented" and the new fields stay empty.

## Progress

Change `[ ]` to `[x]` when a task's PR is merged, or for human tasks when the step is done.

| Done | Task | What | Repo | Who |
|---|---|---|---|---|
| [ ] | P3-S1 | `LunchSyncService` and new fields, unimplemented | budget-server | Agent |
| [ ] | P3-I1 | Generated budget-server client | budget-ingest | Agent |
| [ ] | — | Exit check passed | — | Human |

Later phases implement behind these definitions; they should not need a proto change.

## RPCs

| RPC | Caller | Implemented in |
|---|---|---|
| `LunchSyncService.CreateAccount` | budget-ingest, later the web UI | Phase 4 |
| `LunchSyncService.ListAccounts` | budget-ingest | Phase 4 |
| `LunchSyncService.CreateTransactions` | budget-ingest | Phase 5 |
| `LunchSyncService.UpdateTransaction` | budget-ingest | Phase 6 |
| `LunchSyncService.DeleteTransaction` | budget-ingest | Phase 6 |
| `LunchSyncService.SetAccountBalance` | budget-ingest | Phase 6 |
| `CategoryService.ListCategories` (existing) | budget-ingest | already works |

## Tasks

### P3-S1 `LunchSyncService` and new fields, unimplemented

Proto and wiring only: no migration, no storage change, no behaviour change for existing RPCs.

- New `proto/lunchsync/v1/lunchsync.proto`:
  ```proto
  syntax = "proto3";
  package lunchsync.v1;

  import "account/v1/account.proto";
  import "google/protobuf/timestamp.proto";

  service LunchSyncService {
    rpc CreateAccount(CreateAccountRequest) returns (CreateAccountResponse);
    rpc ListAccounts(ListAccountsRequest) returns (ListAccountsResponse);
    rpc CreateTransactions(CreateTransactionsRequest) returns (CreateTransactionsResponse);
    rpc UpdateTransaction(UpdateTransactionRequest) returns (UpdateTransactionResponse);
    rpc DeleteTransaction(DeleteTransactionRequest) returns (DeleteTransactionResponse);
    rpc SetAccountBalance(SetAccountBalanceRequest) returns (SetAccountBalanceResponse);
  }

  message CreateAccountRequest {
    string name = 1;
    account.v1.AccountType type = 2;
    string sub_type = 3;
    string starting_balance = 4;
  }
  message CreateAccountResponse { string account_id = 1; }

  message ListAccountsRequest {}
  message ListAccountsResponse { repeated account.v1.Account accounts = 1; }

  message NewTransaction {
    string id = 1;                     // client-supplied UUID; the idempotency key
    string amount = 2;                 // money out is negative
    string transaction_name = 3;
    optional string merchant_name = 4;
    google.protobuf.Timestamp transaction_date = 5;
    optional string category_id = 6;
    bool needs_review = 7;
    bool pending = 8;
  }
  message CreateTransactionsRequest {
    string account_id = 1;
    repeated NewTransaction transactions = 2;
  }
  message CreateTransactionsResponse {
    repeated string created_ids = 1;
    repeated string existing_ids = 2;  // already present; left untouched
  }

  // Omitted fields are left unchanged.
  message UpdateTransactionRequest {
    string id = 1;
    optional string amount = 2;
    optional string transaction_name = 3;
    optional string merchant_name = 4;
    optional google.protobuf.Timestamp transaction_date = 5;
    optional string category_id = 6;
    optional bool needs_review = 7;
    optional bool pending = 8;
  }
  message UpdateTransactionResponse {}

  message DeleteTransactionRequest { string id = 1; }
  message DeleteTransactionResponse {}

  message SetAccountBalanceRequest {
    string account_id = 1;
    string balance = 2;
  }
  message SetAccountBalanceResponse {}
  ```
- `proto/account/v1/account.proto`, on `Account`:
  ```proto
  enum AccountIntegration {
    ACCOUNT_INTEGRATION_UNSPECIFIED = 0;
    ACCOUNT_INTEGRATION_MANUAL = 1;
    ACCOUNT_INTEGRATION_PLAID = 2;
    ACCOUNT_INTEGRATION_LUNCHSYNC = 3;
  }
  AccountIntegration integration = 8;
  optional google.protobuf.Timestamp last_synced_at = 9;
  ```
- `proto/transaction/v1/transaction.proto`, on `Transaction`: `bool pending = 10;`
- `make proto-generate`.
- Handler package `internal/connecthandlers/v1/lunchsync/`, registered in `api/connect.go`. Every method returns `connect.CodeUnimplemented`.
- The new fields are left at their zero values. Do not populate them yet.

Acceptance:
- The Buf workflow's lint and breaking-change checks pass (the changes are additive).
- A test calls each of the six RPCs and gets `Unimplemented`.
- All existing tests pass untouched; `make start-local` in budget-stack boots and the web app behaves as before.

### P3-I1 Generated budget-server client

Needs a budget-server release containing P3-S1.

- `buf.gen.yaml` generating Go clients for `lunchsync.v1`, `account.v1`, `category.v1` and `transaction.v1` from the budget-server repo at that release tag, into `internal/budget/gen/`. Add a `proto-generate` make target and commit the generated code, as the web repo does.
- `internal/budget/client.go`: a small wrapper that constructs the `LunchSyncService` and `CategoryService` clients from a base URL with a request timeout. Define it behind an interface so handlers can be tested with a mock (add it to `.mockery.yaml`).
- Config: `BUDGET_SERVER_URL` in `ProcessEnvironmentVariables()`, defaulting to the compose service address.
- The wrapper is constructed at startup but no handler calls it yet.

Acceptance: a test using an in-process Connect server confirms the wrapper reaches each RPC and surfaces `Unimplemented` as an error; `make generate-mocks` produces the mock.

## Exit check

- Calling any `LunchSyncService` RPC on budget-server returns `Unimplemented`.
- In budget-stack, budget-ingest starts with the client configured and logs that budget-server is reachable.
- The web app and Plaid sync behave exactly as before.
