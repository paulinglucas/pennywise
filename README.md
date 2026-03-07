# Pennywise

Self-hosted personal finance dashboard. Dark mode. No ads. No data harvesting.

Track spending, assets, debts, and goals across every financial account you own. Project your path to financial independence.

Two users. One Raspberry Pi. Zero ongoing cost.

## Tech Stack

- **Backend:** Go 1.25, Chi router, SQLite (WAL mode, pure Go driver)
- **Frontend:** React 19, TypeScript 5, Tailwind CSS v4, TanStack Query, Recharts
- **Contract:** OpenAPI spec with code generation (oapi-codegen + openapi-typescript)
- **Tooling:** just, mise, Docker Compose, GitHub Actions CI

## Prerequisites

- [mise](https://mise.jdx.dev) for tool version management
- [Docker](https://docs.docker.com/get-docker/) and Docker Compose

## Getting Started

```bash
mise install          # Install Go 1.25, Node 22, just
just setup            # Install Go and Node dependencies
just dev              # Start full dev environment via Docker Compose
```

## Available Commands

Run `just --list` to see all available recipes:

| Command | Description |
|---------|-------------|
| `just setup` | First-time project setup |
| `just dev` | Start full dev environment |
| `just dev-backend` | Start backend only |
| `just dev-frontend` | Start frontend only |
| `just generate` | Regenerate types from OpenAPI spec |
| `just migrate` | Run database migrations |
| `just reset-db` | Drop and recreate database |
| `just seed` | Reset database and load sample data |
| `just backup-db` | Snapshot database before migrations |
| `just lint` | Run all linters |
| `just test` | Run all tests |
| `just build` | Build production binaries |
| `just ci` | Run full CI check locally |

## Observability

The backend exposes a `/metrics` endpoint (Prometheus format, localhost-only) and OpenTelemetry tracing. A pre-built Grafana dashboard is included at `deploy/grafana-dashboard.json`.

To set up monitoring:

1. Configure Prometheus with `deploy/prometheus.yml`
2. Import `deploy/grafana-dashboard.json` into Grafana
3. Panels include: request rate, error rate, response time percentiles, DB query performance, failed request trends

Web Vitals from the frontend are ingested via `POST /api/v1/telemetry/vitals` (no auth required).

## License

MIT with additional restrictions. See [LICENSE](LICENSE).
