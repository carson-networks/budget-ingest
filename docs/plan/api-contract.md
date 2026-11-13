# LunchSync API contract (Lunch Money v2 subset)

What budget-ingest must serve at `https://api.lunchmoney.dev/v2`. Shapes are taken from the `V2*` structs in LunchSync's `Api/LunchMoneyAPIV2.swift`: request fields are what it encodes, response fields are what it decodes. Fields marked **required** are non-optional in the Swift decoder, so omitting them fails the whole call.

## Common

- Header on every request: `Authorization: Bearer <token>`, `Content-Type: application/json`.
- Any 2xx is success. Never return `429`; LunchSync sleeps and retries on it.
- Errors: non-2xx with `{"message": "...", "errors": [{"errMsg": "..."}]}`.
- Unknown request fields are ignored. `custom_metadata` in particular is large and optional.
- Dates are `yyyy-MM-dd`. Amounts in transaction bodies are strings with two decimals. Currency is always `"usd"`.

## Endpoints

### `GET /me`

Called when the token is entered, and to show the account name.

```json
{
  "id": 1,
  "name": "Budget",
  "email": "budget@localhost",
  "account_id": 1,
  "budget_name": "Budget",
  "primary_currency": "usd",
  "api_key_label": "<token name>"
}
```

All fields required except `api_key_label`.

### `GET /manual_accounts`

Lists accounts the user can pair a Wallet account with: LunchSync-type accounts only, from `LunchSyncService.ListAccounts`.

```json
{ "manual_accounts": [
  { "id": 3, "name": "Apple Card", "type": "credit", "subtype": "credit card",
    "balance": "-120.5000", "currency": "usd", "institution_name": "Apple",
    "balance_as_of": "2026-11-11T12:00:00Z", "created_at": "2026-11-01T00:00:00Z" }
] }
```

Required: `id`, `name`, `type`, `balance` (string).

### `POST /manual_accounts`

Sent when the user creates an account from the app instead of pairing an existing one. Creates a LunchSync-type account through `LunchSyncService.CreateAccount`.

Request: `name` (required), `type`, `subtype?`, `balance` (number, used as the starting balance), `currency?`, `institution_name?`.

The app sends one of these, and the web UI's account form mirrors them:

| Wallet account | `type` | `subtype` | budget-server |
|---|---|---|---|
| Apple Card | `credit` | `credit card` | `ACCOUNT_TYPE_CREDIT_CARDS` |
| Apple Savings | `cash` | `savings` | `ACCOUNT_TYPE_CASH` |
| Apple Cash | `cash` | `digital wallet` | `ACCOUNT_TYPE_CASH` |

`subtype` is stored as budget-server's `sub_type`. Response: the created manual account object (same shape as in the list).

### `PUT /manual_accounts/{id}`

Balance updates after each sync. Request fields, all optional: `balance` (number), `balance_as_of`, `name`, `display_name`, `institution_name`, `currency`, `exclude_from_transactions`. Only `balance` and `balance_as_of` are acted on. `balance` is passed through unchanged.

Response: the updated manual account object.

### `GET /categories?format=flattened`

Used for the app's category mapping screen.

```json
{ "categories": [
  { "id": 7, "name": "Groceries", "is_income": false, "is_group": false, "group_id": 2, "archived": false }
] }
```

Required: `id`, `name`. Parent categories are returned as `is_group: true`; children carry `group_id`.

### `POST /transactions`

The main sync call. Up to 500 items per request.

Request:

```json
{
  "transactions": [
    { "date": "2026-11-10", "amount": "12.34", "currency": "usd", "payee": "Blue Bottle",
      "category_id": 7, "notes": null, "manual_account_id": 3,
      "status": "unreviewed", "external_id": "<FinanceKit transaction id>",
      "custom_metadata": { "source": "...", "wallet": { "merchantName": "...", "status": "booked", "isPending": false } } }
  ],
  "apply_rules": true, "skip_duplicates": true, "skip_balance_update": false
}
```

- `amount`: positive is money out, negative is money in. Negate it to match budget-server (the Plaid provider does the same).
- `manual_account_id` is omitted when the Wallet account isn't paired; reject those items.
- `custom_metadata` is present only when the app's "send FinanceKit metadata" setting is on. `wallet.status` is one of `booked`, `pending`, `authorized`, `memo`, `rejected`.
- `category_id` is omitted when no mapping exists.
- `apply_rules`, `skip_duplicates`, `skip_balance_update` are ignored. Dedup on `external_id` always applies.

Response:

```json
{
  "transactions": [ { "id": 101, "date": "2026-11-10", "amount": "12.34", "payee": "Blue Bottle",
                      "external_id": "...", "manual_account_id": 3, "status": "unreviewed" } ],
  "skipped_duplicates": [
    { "reason": "duplicate_external_id", "request_transactions_index": 4,
      "existing_transaction_id": 88, "request_transaction": { "external_id": "..." } }
  ]
}
```

- `transactions` holds only newly inserted rows. Required per row: `id`, `date`, `amount`. LunchSync matches rows back to its request by `external_id`, so always echo it.
- Every item whose `(manual_account_id, external_id)` already exists is reported in `skipped_duplicates` with `reason: "duplicate_external_id"` and the existing ID. This is how LunchSync recovers IDs after a retry.
- Amounts in responses use LunchSync's sign (positive is money out).

### `PUT /transactions/{id}`

Sent when a Wallet transaction changes (pending settles, amount or payee changes) or the user edits it in the app. All fields optional: `date`, `payee`, `amount`, `currency`, `category_id`, `manual_account_id`, `notes`, `status`, `custom_metadata`. `external_id` is never sent.

Response: the full transaction object. LunchSync only checks for 2xx.

### `GET /transactions` and `GET /transactions/{id}`

Query: `start_date`, `end_date`, `manual_account_id?`, `limit?`, `offset?`, plus `include_group_children` and `include_split_parents` (ignored).

```json
{ "transactions": [ ... ], "has_more": false }
```

Returns only transactions budget-ingest has an ID mapping for. The v2 sync path should not need the list call, but the client implements it and a correct answer is cheap.

## Status mapping

| LunchSync `status` | budget-server |
|---|---|
| `unreviewed` or omitted | `needs_review = true` |
| `reviewed` | `needs_review = false` |

A transaction with no `category_id` is always `needs_review = true`.

## Pending and rejected

Driven by `custom_metadata.wallet.status`, on both insert and update:

| `wallet.status` | Action |
|---|---|
| `booked` | `pending = false` |
| `pending`, `authorized`, `memo` | `pending = true` |
| `rejected` on insert | Not created in budget-server. An ID is still allocated and returned as inserted, so the app stops retrying it. |
| `rejected` on update | Deleted in budget-server. The mapping is kept as a tombstone so a re-push is reported as a duplicate, not recreated. |
| metadata absent | `pending = false` on insert; unchanged on update |
