FROM golang:1.25-alpine AS backend-build
WORKDIR /build
COPY backend/go.mod backend/go.sum* ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 go build -o /pennywise cmd/server/main.go

FROM node:22-alpine AS frontend-build
WORKDIR /build
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

FROM alpine:3.21
RUN apk add --no-cache nginx sqlite
RUN adduser -D -H pennywise

COPY --from=backend-build /pennywise /opt/pennywise/server
COPY --from=frontend-build /build/dist /opt/pennywise/frontend
COPY deploy/nginx.conf /etc/nginx/http.d/default.conf

RUN mkdir -p /opt/pennywise/data \
    && chown pennywise:pennywise /opt/pennywise/data \
    && sed -i 's/^user nginx;/user pennywise;/' /etc/nginx/nginx.conf \
    && mkdir -p /var/lib/nginx/tmp /var/lib/nginx/logs /var/log/nginx /run/nginx \
    && chown -R pennywise:pennywise /var/lib/nginx /var/log/nginx /run/nginx

EXPOSE 8081

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -q --spider http://localhost:8080/api/v1/health || exit 1

USER pennywise
WORKDIR /opt/pennywise

CMD ["sh", "-c", "nginx -g 'daemon on; master_process off;' && exec ./server"]
