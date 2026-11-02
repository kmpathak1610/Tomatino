# Architecture

Modular monolith Go backend exposing REST /api/v1. Next.js apps (customer/admin/restaurant/delivery) consume APIs. PostgreSQL + PostGIS source of truth. Redis for cache/jobs. MinIO for storage. Provider interfaces for payments/SMS/push/maps.

See monorepo structure in README.