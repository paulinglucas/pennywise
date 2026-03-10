#!/usr/bin/env bash
set -euo pipefail

DB_PATH="${PENNYWISE_DB_PATH:-/opt/pennywise/data/pennywise.db}"
BACKUP_DIR="${PENNYWISE_BACKUP_DIR:-/opt/pennywise/data/backups}"
MAX_BACKUPS="${PENNYWISE_MAX_BACKUPS:-30}"
B2_BUCKET="${PENNYWISE_B2_BUCKET:-}"

if [ ! -f "$DB_PATH" ]; then
    echo "database not found at $DB_PATH, nothing to back up"
    exit 0
fi

mkdir -p "$BACKUP_DIR"

STAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/pennywise_${STAMP}.db"

sqlite3 "$DB_PATH" "PRAGMA wal_checkpoint(TRUNCATE);"
cp "$DB_PATH" "$BACKUP_FILE"

echo "backed up to $BACKUP_FILE"

BACKUP_COUNT=$(find "$BACKUP_DIR" -name "pennywise_*.db" -type f | wc -l)
if [ "$BACKUP_COUNT" -gt "$MAX_BACKUPS" ]; then
    ls -1t "$BACKUP_DIR"/pennywise_*.db | tail -n +$((MAX_BACKUPS + 1)) | xargs -r rm -f
    REMOVED=$((BACKUP_COUNT - MAX_BACKUPS))
    echo "rotated $REMOVED old backup(s), keeping $MAX_BACKUPS"
fi

if [ -n "$B2_BUCKET" ]; then
    if command -v b2 > /dev/null 2>&1; then
        b2 upload-file "$B2_BUCKET" "$BACKUP_FILE" "backups/pennywise_${STAMP}.db"
        echo "uploaded to Backblaze B2 bucket $B2_BUCKET"
    else
        echo "b2 CLI not found, skipping cloud upload"
    fi
fi
