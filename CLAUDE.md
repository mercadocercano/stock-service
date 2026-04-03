# CLAUDE.md — stock-service

Guía breve para asistentes de código. Hablar siempre en español.

## Identidad

Servicio de stock e inventario del ecosistema SaaS Marketplace (multi-tenant).

## Stack y puertos

Go, Gin, PostgreSQL. **8100** (host) → **8080** (Docker). Base **stock_db**.

## Comandos

`go run main.go` · `go test ./...` · `go build -o stock-service .` (raíz con `main.go`, no `cmd/api`).

## Despliegue

`k8s/stock/`. Kong: **`/stock/`**.

## Endpoints de referencia

`POST /api/v1/sale`, `POST /api/v1/stock-entries`, `GET /api/v1/availability`, `POST /api/v1/stock/process-sale-atomic`, `POST /api/v1/stock/compensate/:id`, `POST /api/v1/stock-entries/bulk`.

## Referencias

`ai-tools/rules/architecture.md`, `ai-tools/rules/multi-tenant.md`, `ai-tools/rules/api-gateway.md`.
