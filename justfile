set dotenv-load

default:
    @just --list

# ---------- Setup ----------

setup:
    mise install
    cd backend && go mod download
    cd frontend && npm install

# ---------- Development ----------

dev:
    docker compose up

dev-backend:
    cd backend && PENNYWISE_DB_PATH=../{{db_path}} go run cmd/server/main.go

dev-frontend:
    cd frontend && npm run dev

# ---------- Code Generation ----------

generate:
    oapi-codegen -package api -generate types,chi-server api/openapi.yaml > backend/internal/api/generated.go
    cd frontend && npx openapi-typescript ../api/openapi.yaml -o src/api/generated.ts

check-generated:
    just generate
    git diff --exit-code backend/internal/api/generated.go frontend/src/api/generated.ts

# ---------- Database ----------

db_path := "data/pennywise.db"
backup_dir := "data/backups"
max_backups := "5"

backup-db:
    #!/usr/bin/env sh
    if [ ! -f {{db_path}} ]; then exit 0; fi
    mkdir -p {{backup_dir}}
    stamp=$(date +%Y%m%d_%H%M%S)
    sqlite3 {{db_path}} "PRAGMA wal_checkpoint(TRUNCATE);"
    cp {{db_path}} {{backup_dir}}/pennywise_${stamp}.db
    ls -1t {{backup_dir}}/pennywise_*.db | tail -n +$(({{max_backups}} + 1)) | xargs -r rm -f
    echo "backed up to {{backup_dir}}/pennywise_${stamp}.db"

migrate: backup-db
    cd backend && PENNYWISE_DB_PATH=../{{db_path}} go run cmd/server/main.go migrate

reset-db:
    rm -f data/pennywise.db
    just migrate

seed:
    just reset-db
    sqlite3 data/pennywise.db < scripts/seed_data.sql

# ---------- Testing ----------

test: test-backend test-frontend

test-backend:
    cd backend && go test ./... -race -coverprofile=coverage.out
    cd backend && go tool cover -func=coverage.out

test-frontend:
    cd frontend && npx vitest run --coverage

# ---------- Linting ----------

lint: lint-backend lint-frontend

lint-backend:
    cd backend && golangci-lint run ./...

lint-frontend:
    cd frontend && npx eslint . --max-warnings 0
    cd frontend && npx prettier --check .
    cd frontend && npx tsc --noEmit

# ---------- Build ----------

build: build-backend build-frontend

build-backend:
    cd backend && CGO_ENABLED=0 go build -o ../bin/pennywise cmd/server/main.go

build-frontend:
    cd frontend && npm run build

build-docker:
    docker build --network=host -t pennywise:latest .

# ---------- Utilities ----------

scan-secrets:
    gitleaks detect --source . --verbose

clean:
    rm -rf bin/ data/pennywise.db
    rm -f backend/internal/api/generated.go
    rm -f frontend/src/api/generated.ts
    cd backend && go clean
    cd frontend && rm -rf dist/ node_modules/.cache coverage/

ci: lint test build
