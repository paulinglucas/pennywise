#!/usr/bin/env bash
set -euo pipefail

IMAGE="${PENNYWISE_IMAGE:-pennywise:latest}"
CONTAINER="${PENNYWISE_CONTAINER:-pennywise}"
DATA_DIR="${PENNYWISE_DATA_DIR:-/opt/pennywise/data}"
PORT="${PENNYWISE_PORT:-80}"
HEALTH_URL="http://localhost:${PORT}/api/v1/health"
HEALTH_RETRIES=10
HEALTH_DELAY=3

echo "deploying $IMAGE as $CONTAINER"

PREVIOUS_IMAGE=""
if docker inspect "$CONTAINER" > /dev/null 2>&1; then
    PREVIOUS_IMAGE=$(docker inspect --format='{{.Config.Image}}' "$CONTAINER")
    echo "previous image: $PREVIOUS_IMAGE"
fi

echo "running pre-deploy backup"
PENNYWISE_DB_PATH="$DATA_DIR/pennywise.db" \
    "$(dirname "$0")/backup.sh"

if docker inspect "$CONTAINER" > /dev/null 2>&1; then
    echo "stopping previous container"
    docker stop "$CONTAINER" || true
    docker rename "$CONTAINER" "${CONTAINER}_rollback" 2>/dev/null || true
fi

echo "starting new container"
docker run -d \
    --name "$CONTAINER" \
    --restart unless-stopped \
    -p "$PORT:8081" \
    -v "$DATA_DIR:/opt/pennywise/data" \
    --env-file "${DATA_DIR}/../.env" \
    "$IMAGE"

echo "waiting for health check"
HEALTHY=false
for i in $(seq 1 $HEALTH_RETRIES); do
    sleep "$HEALTH_DELAY"
    STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$HEALTH_URL" 2>/dev/null || echo "000")
    if [ "$STATUS" = "200" ]; then
        HEALTHY=true
        echo "health check passed on attempt $i"
        break
    fi
    echo "attempt $i/$HEALTH_RETRIES: status $STATUS"
done

if [ "$HEALTHY" = true ]; then
    echo "deploy successful"
    docker rm -f "${CONTAINER}_rollback" 2>/dev/null || true
else
    echo "health check failed after $HEALTH_RETRIES attempts, rolling back"
    docker stop "$CONTAINER" 2>/dev/null || true
    docker rm "$CONTAINER" 2>/dev/null || true

    if docker inspect "${CONTAINER}_rollback" > /dev/null 2>&1; then
        docker rename "${CONTAINER}_rollback" "$CONTAINER"
        docker start "$CONTAINER"
        echo "rolled back to previous version"
    else
        echo "no previous version to roll back to"
        exit 1
    fi
    exit 1
fi
