# ADR-001: Transacciones atómicas de stock con lock row-level

**Estado**: Aceptado
**Fecha**: 2026-06-10
**Contexto**: Existía una race condition entre la validación de disponibilidad (`CheckAvailability`) y el descuento de stock (`ProcessSale`): dos transacciones concurrentes podían leer el mismo valor sin lock y sobrevender. Había que garantizar atomicidad sin mover lógica de negocio a la base de datos.

## Decisión

Adoptamos una operación atómica `ProcessSaleAtomic` que reemplaza el par `CheckAvailability` + `ProcessSale`. La atomicidad se logra con una transacción SQL que usa `SELECT FOR UPDATE` sobre la fila específica de `stock_availability`. La validación de negocio permanece en Go; PostgreSQL solo garantiza el aislamiento (lock row-level), no valida. Los triggers no cambian: solo recalculan agregados, no contienen lógica de negocio. El rollback de ventas fallidas se hace por compensación explícita (`compensate/:id`), no por saga automática. No se usa UPSERT defensivo, por lo que no se crean filas fantasma en `stock_availability`.

## Alternativas consideradas

| Opción | Por qué no |
|--------|-----------|
| UPSERT para garantizar la fila | Crea estado artificial (filas fantasma) |
| Validación dentro del trigger | Mezcla persistencia con lógica de negocio |
| Saga pattern completa | Complejidad innecesaria para un flujo de solo 2 pasos |
| Event sourcing | Over-engineering para este caso |
| Lock distribuido (Redis) | Los locks de DB son suficientes |

## Consecuencias

**Positivas**: Elimina la sobreventa por concurrencia; la lógica de negocio queda testeable en Go fuera de la DB; el aislamiento es de fila, no de tabla.
**Negativas / trade-offs**: El rollback es manual (compensación explícita), no automático; `SELECT FOR UPDATE` introduce contención sobre filas calientes.
**Neutral**: No requiere migraciones nuevas (usa el schema existente); el endpoint HTTP no cambia de firma salvo el nuevo campo `reference`.
