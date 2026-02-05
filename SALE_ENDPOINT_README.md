# Endpoint de Venta Mínimo (Mock)

## Descripción

Endpoint mínimo para procesar ventas que descuentan stock automáticamente. Esta es una implementación mockeada que **NO incluye**:
- Gestión de órdenes persistentes
- Pagos
- Clientes
- Estados complejos de orden
- Eventos o colas de mensajería
- Consistencia distribuida

## Endpoint

```
POST /api/v1/sale
```

## Headers Requeridos

```
X-Tenant-ID: <tenant-uuid>
Content-Type: application/json
```

## Request Body

```json
{
  "variant_sku": "PRODUCT-SKU-001",
  "quantity": 3
}
```

### Campos

| Campo | Tipo | Requerido | Descripción |
|-------|------|-----------|-------------|
| `variant_sku` | string | Sí | SKU de la variante del producto |
| `quantity` | float | Sí | Cantidad a vender (debe ser > 0) |

## Responses

### Éxito (200 OK)

```json
{
  "success": true,
  "message": "Sale processed successfully",
  "variant_sku": "PRODUCT-SKU-001",
  "quantity_sold": 3,
  "remaining_stock": 7,
  "stock_entry_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Stock Insuficiente (400 Bad Request)

```json
{
  "success": false,
  "message": "Insufficient stock. Available: 7.00, Requested: 10.00",
  "variant_sku": "PRODUCT-SKU-001",
  "remaining_stock": 7
}
```

### Producto No Encontrado (400 Bad Request)

```json
{
  "success": false,
  "message": "Product not found: PRODUCT-SKU-001",
  "variant_sku": "PRODUCT-SKU-001"
}
```

### Error de Validación (400 Bad Request)

```json
{
  "error": "Invalid request data",
  "details": "quantity must be greater than zero"
}
```

## Flujo de Procesamiento

1. **Validación de Request**: Verifica que `variant_sku` y `quantity` sean válidos
2. **Verificación de Disponibilidad**: Consulta el stock disponible via endpoint `/availability`
3. **Validación de Stock**: Verifica que haya stock suficiente
4. **Registro de Venta**: Crea una entrada de stock de tipo `sale` con cantidad negativa
5. **Cálculo de Stock Restante**: Retorna el stock disponible después de la venta

## Comportamiento del Stock

- El endpoint crea una entrada de tipo `EntryTypeSale` en la tabla `stock_entries`
- La cantidad se registra como positiva en la entrada, pero se interpreta como negativa al calcular disponibilidad
- La vista materializada o el cálculo de disponibilidad resta automáticamente las ventas del stock total

## Ejemplo de Uso con cURL

### Via Kong Gateway (Recomendado)

```bash
curl -X POST http://localhost:8001/stock/api/v1/sale \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001" \
  -d '{
    "variant_sku": "PRODUCT-001",
    "quantity": 3
  }'
```

### Directo al Servicio (Solo desarrollo)

```bash
curl -X POST http://localhost:8100/api/v1/sale \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001" \
  -d '{
    "variant_sku": "PRODUCT-001",
    "quantity": 3
  }'
```

## Script de Prueba Completo

Ejecuta el script de prueba incluido:

```bash
cd services/stock-service
./test-sale-flow.sh
```

Este script ejecuta un flujo completo:
1. Crea stock inicial de 10 unidades
2. Verifica disponibilidad inicial
3. Procesa venta de 3 unidades
4. Verifica que el stock resultante sea 7
5. Intenta venta sin stock suficiente (debe fallar)

## Criterio de Éxito

✅ **Flujo exitoso**:
- Stock inicial: 10 unidades
- Venta procesada: 3 unidades
- Stock resultante: 7 unidades
- `GET /availability?sku=<sku>` retorna `available_quantity: 7`

## Limitaciones Conocidas

1. **No hay persistencia de órdenes**: La venta solo se registra como una entrada de stock
2. **No hay rollback automático**: Si necesitas revertir, usa un ajuste manual
3. **No hay validación de negocio adicional**: No valida precios, clientes, métodos de pago, etc.
4. **No hay eventos**: No se emiten eventos de venta para otros servicios
5. **No hay transaccionalidad distribuida**: Solo garantiza consistencia dentro de stock-service

## Integración con Order Service (Futuro)

Cuando se implemente un `order-service` completo, este endpoint puede:
- Ser llamado desde order-service tras confirmar el pago
- Ser reemplazado por el flujo reserve → consume
- Mantenerse como opción para ventas POS sin orden previa

## Notas Técnicas

- **Arquitectura**: Sigue arquitectura hexagonal del proyecto
- **Multi-tenant**: Respeta isolación por tenant_id
- **Idempotencia**: No implementada - cada llamada crea una nueva entrada
- **Concurrencia**: No hay locks - posible condición de carrera en alta concurrencia
- **Auditoría**: Cada venta se registra con timestamp y reference_number único

## Próximos Pasos (No Incluidos)

- [ ] Implementar order-service completo
- [ ] Agregar flujo de pagos
- [ ] Implementar eventos de dominio
- [ ] Agregar locks optimistas para concurrencia
- [ ] Implementar idempotencia con idempotency-key
- [ ] Agregar integración con facturación
