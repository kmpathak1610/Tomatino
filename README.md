# Food Delivery Platform

Monorepo: Next.js apps + Go backend + PostgreSQL/PostGIS.

## Quick start
```
docker compose up -d
cd backend && go run ./cmd/migrate
cd backend && go run ./cmd/api
```

Frontend: pnpm install && pnpm dev