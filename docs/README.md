# Documentación — stock-service

Servicio de stock e inventario del ecosistema SaaS Marketplace (multi-tenant).

## Architecture Decision Records

| ADR | Título | Estado | Fecha |
|-----|--------|--------|-------|
| [ADR-001](adr/ADR-001-transacciones-atomicas-stock.md) | Transacciones atómicas de stock con lock row-level | Aceptado | 2026-06-10 |
| [ADR-002](adr/ADR-002-coordinacion-stock-multi-canal.md) | Coordinación de stock multi-canal con dominio puro y stock físico único | Aceptado | 2026-06-10 |

## Arquitectura

- [Transacciones atómicas (HITO D)](architecture/transacciones-atomicas.md) — eliminación de la race condition de stock con `SELECT FOR UPDATE`.
- [Módulo multi-canal](architecture/multi-canal.md) — coordinación de stock entre POS y Marketplace (dominio puro).
- [Flujo de venta mínimo](architecture/flujo-venta-minimo.md) — implementación del flujo de venta que descuenta stock.

## Guías

- [Endpoint de venta](guides/endpoint-venta.md) — contrato y uso del `POST /api/v1/sale`.
- [Comandos de prueba rápida](guides/comandos-prueba-rapida.md) — curls para probar el flujo de venta end-to-end.

## API

El contrato de API se documenta vía OpenAPI: [`../api-docs/openapi.yml`](../api-docs/openapi.yml).
