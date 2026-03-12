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

## Frontend

The React frontend provides a single-page app with:

- **Auth flow:** Login page with cookie-based session, automatic redirect when unauthenticated
- **App shell:** Responsive layout with sidebar navigation (desktop) and bottom tab bar (mobile)
- **Routing:** Dashboard, Transactions, Assets, Goals, Projections pages
- **API client:** Typed fetch wrapper using OpenAPI-generated types, with `ApiError` class for structured error handling
- **State management:** TanStack Query for server state with auth-aware query caching

## Production Deployment (Raspberry Pi)

### Build the Docker image

```bash
docker build -t pennywise:latest .
```

Or pull a tagged release from GitHub Container Registry:

```bash
docker pull ghcr.io/paulinglucas/pennywise:latest
```

### Run with Docker

```bash
mkdir -p /opt/pennywise/data

docker run -d \
  --name pennywise \
  --restart unless-stopped \
  -p 80:8081 \
  -v /opt/pennywise/data:/opt/pennywise/data \
  --env-file /opt/pennywise/.env \
  pennywise:latest
```

Create `/opt/pennywise/.env` with:

```bash
PENNYWISE_JWT_SECRET=<generate-a-random-64-char-string>
PENNYWISE_DB_PATH=/opt/pennywise/data/pennywise.db
```

### Run with systemd (bare metal)

```bash
# Build the binary
cd backend && CGO_ENABLED=0 go build -o /opt/pennywise/server cmd/server/main.go
cd frontend && npm ci && npm run build
cp -r dist /opt/pennywise/frontend

# Create service user
useradd -r -s /usr/sbin/nologin pennywise
mkdir -p /opt/pennywise/data
chown pennywise:pennywise /opt/pennywise/data

# Install systemd service
cp deploy/pennywise.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable --now pennywise
```

### Deploy with automatic rollback

The deploy script backs up the database, launches the new container, verifies the health check, and rolls back automatically if it fails:

```bash
PENNYWISE_IMAGE=pennywise:v1.2.0 scripts/deploy.sh
```

### Database backups

Run manually or via cron:

```bash
scripts/backup.sh
```

Configure via environment variables:

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `PENNYWISE_DB_PATH` | `/opt/pennywise/data/pennywise.db` | Path to SQLite database |
| `PENNYWISE_BACKUP_DIR` | `/opt/pennywise/data/backups` | Backup output directory |
| `PENNYWISE_MAX_BACKUPS` | `30` | Number of backups to retain |
| `PENNYWISE_B2_BUCKET` | (empty) | Backblaze B2 bucket for cloud backup |

Automate daily backups with cron:

```cron
0 2 * * * /opt/pennywise/scripts/backup.sh >> /var/log/pennywise-backup.log 2>&1
```

### Releasing new versions

Tag a release to trigger the GitHub Actions release workflow:

```bash
git tag v1.0.0
git push origin v1.0.0
```

This builds and pushes a Docker image to GitHub Container Registry. Then deploy on the Pi:

```bash
docker pull ghcr.io/paulinglucas/pennywise:1.0.0
PENNYWISE_IMAGE=ghcr.io/paulinglucas/pennywise:1.0.0 scripts/deploy.sh
```

## Account Sync (SimpleFIN)

Pennywise syncs account balances from your bank via [SimpleFIN](https://www.simplefin.org/), a read-only financial data protocol. SimpleFIN cannot move money or initiate transactions — it only reads balances.

### How it works

1. Sign up at [bridge.simplefin.org](https://bridge.simplefin.org/) ($1.50/mo or $15/yr) and connect your bank accounts
2. Generate a Setup Token from the Bridge dashboard
3. In Pennywise, go to **Settings** and paste the token
4. Map each SimpleFIN account to a Pennywise account using the link UI
5. Balances sync automatically once daily (default 6:00 AM local time)

### Security model

- **Read-only access.** SimpleFIN has no write endpoints. No one in the chain can move your money.
- **Your bank credentials never touch Pennywise.** Authentication with your bank happens through the SimpleFIN Bridge (via the MX aggregator). Pennywise only receives an API access URL.
- **The access URL is encrypted at rest.** Stored in SQLite using AES-256-GCM, with the key derived from your JWT secret via PBKDF2.
- **Setup tokens are single-use.** Once claimed, a token cannot be reused. If a second claim is attempted, SimpleFIN returns 403.
- **You control revocation.** Disconnect from Pennywise Settings, or revoke from the SimpleFIN Bridge dashboard at any time.

### Data flow

```text
Your Bank -> MX (aggregator) -> SimpleFIN Bridge -> Pennywise (your hardware)
```

Bank credentials are stored by MX, not by SimpleFIN or Pennywise. All data after the Bridge lives exclusively in your local SQLite database.

### Balance history

Each sync that detects a balance change (threshold: >0.5 cents) updates the asset's current value and writes a record to the `asset_history` table. This provides a time series of your account balances over time.

### Configuration

| Variable | Default | Description |
| -------- | ------- | ----------- |
| `PENNYWISE_SYNC_HOUR` | `6` | Hour of day (0-23) for automatic daily sync |

You can also trigger a manual sync from the Settings page at any time.

## Observability

The backend exposes a `/metrics` endpoint (Prometheus format, localhost-only) and OpenTelemetry tracing. A pre-built Grafana dashboard is included at `deploy/grafana-dashboard.json`.

To set up monitoring:

1. Configure Prometheus with `deploy/prometheus.yml`
2. Import `deploy/grafana-dashboard.json` into Grafana
3. Panels include: request rate, error rate, response time percentiles, DB query performance, failed request trends

Web Vitals from the frontend are ingested via `POST /api/v1/telemetry/vitals` (no auth required).

## License

MIT with additional restrictions. See [LICENSE](LICENSE).
