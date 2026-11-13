# budget-ingest

A small bridge that answers as the [Lunch Money](https://lunchmoney.dev) v2 API so the unmodified LunchSync iOS app can push Apple Card, Apple Cash and Apple Savings transactions into budget-server. It speaks Lunch Money's REST API to the phone and ConnectRPC to budget-server.

Right now it is a skeleton: it serves `GET /healthz` on port `9448` and nothing else.

The plan, phases and API contract live in [docs/plan/README.md](docs/plan/README.md). Development conventions are in [docs/dev-rules.md](docs/dev-rules.md).

## Running locally

Go is installed with [mise](https://mise.jdx.dev) (see `mise.toml`):

```bash
mise install
go test ./...
```

To run this service alone in Docker from your local changes (listens on `localhost:9448`):

```bash
make serve-local
```

`make serve-local-rebuild` rebuilds and restarts it; `make help` lists every target.

The whole system is run and tested together through [budget-stack](https://github.com/carson-networks/budget-stack), not service by service. See [docs/dev-rules.md](docs/dev-rules.md).
