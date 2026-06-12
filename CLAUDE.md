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

## Memoria persistente (Engram)

Tenés acceso a memoria persistente entre sesiones vía las herramientas MCP de Engram (`mem_save`, `mem_search`, `mem_context`, etc.). Proyecto: **`mercado-cercano`** (memoria compartida con el resto del ecosistema).

**Cuándo guardar** — sin esperar que te lo pidan:
- Al resolver un bug no trivial: síntoma, causa raíz, fix aplicado.
- Al tomar una decisión de diseño: qué se decidió y por qué.
- Al descubrir un patrón o convención del proyecto que no está documentada.
- Al completar una feature o refactor significativo: qué cambió y dónde.

**Cuándo buscar** — antes de empezar cualquier tarea:
- `mem_context` al inicio de sesión o tras una compaction para recuperar el estado anterior.
- `mem_search` cuando el usuario menciona algo que puede tener historial ("el bug de stock atómico", "la migración de la semana pasada").

**Al cerrar sesión**: llamar `mem_session_summary` para dejar un resumen recuperable.
