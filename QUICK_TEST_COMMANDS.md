# Comandos de Prueba Rápida - Endpoint de Venta

## Variables de Entorno

```bash
export TENANT_ID="00000000-0000-0000-0000-000000000001"
export VARIANT_SKU="TEST-VARIANT-001"
export BASE_URL="http://localhost:8001/stock/api/v1"
```

## 1. Crear Stock Inicial (10 unidades)

```bash
curl -X POST $BASE_URL/stock-entries \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "variant_sku": "'$VARIANT_SKU'",
    "entry_type": "initial_stock",
    "quantity": 10,
    "product_name": "Producto de Prueba"
  }' | jq '.'
```

## 2. Verificar Disponibilidad Inicial

```bash
curl -X GET "$BASE_URL/availability?sku=$VARIANT_SKU" \
  -H "X-Tenant-ID: $TENANT_ID" | jq '.'
```

**Esperado**: `available_quantity: 10`

## 3. Procesar Venta (3 unidades)

```bash
curl -X POST $BASE_URL/sale \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "variant_sku": "'$VARIANT_SKU'",
    "quantity": 3
  }' | jq '.'
```

**Esperado**:
```json
{
  "success": true,
  "message": "Sale processed successfully",
  "variant_sku": "TEST-VARIANT-001",
  "quantity_sold": 3,
  "remaining_stock": 7,
  "stock_entry_id": "..."
}
```

## 4. Verificar Stock Después de Venta

```bash
curl -X GET "$BASE_URL/availability?sku=$VARIANT_SKU" \
  -H "X-Tenant-ID: $TENANT_ID" | jq '.'
```

**Esperado**: `available_quantity: 7`

## 5. Intentar Venta Sin Stock Suficiente

```bash
curl -X POST $BASE_URL/sale \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "variant_sku": "'$VARIANT_SKU'",
    "quantity": 100
  }' | jq '.'
```

**Esperado**:
```json
{
  "success": false,
  "message": "Insufficient stock. Available: 7.00, Requested: 100.00",
  "variant_sku": "TEST-VARIANT-001",
  "remaining_stock": 7
}
```

## 6. Intentar Venta de Producto Inexistente

```bash
curl -X POST $BASE_URL/sale \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "variant_sku": "NO-EXISTE",
    "quantity": 1
  }' | jq '.'
```

**Esperado**:
```json
{
  "success": false,
  "message": "Product not found: NO-EXISTE",
  "variant_sku": "NO-EXISTE"
}
```

## Script Automatizado

Para ejecutar todas las pruebas automáticamente:

```bash
cd /Users/hornosg/MyProjects/saas-mt/services/stock-service
./test-sale-flow.sh
```

## Verificar Entradas de Stock en DB

```sql
-- Conectarse a la base de datos
docker exec -it stock-db psql -U postgres -d stock_db

-- Ver todas las entradas de stock
SELECT 
  id, 
  variant_sku, 
  entry_type, 
  quantity, 
  reference_number, 
  created_at 
FROM stock_entries 
WHERE tenant_id = '00000000-0000-0000-0000-000000000001'
  AND variant_sku = 'TEST-VARIANT-001'
ORDER BY created_at DESC;

-- Ver disponibilidad calculada
SELECT * FROM stock_availability 
WHERE tenant_id = '00000000-0000-0000-0000-000000000001'
  AND variant_sku = 'TEST-VARIANT-001';
```

## Limpiar Datos de Prueba

```bash
curl -X POST $BASE_URL/stock-entries \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "variant_sku": "'$VARIANT_SKU'",
    "entry_type": "adjustment",
    "quantity": -100,
    "notes": "Limpieza de datos de prueba"
  }'
```

O directamente en la base de datos:

```sql
DELETE FROM stock_entries 
WHERE tenant_id = '00000000-0000-0000-0000-000000000001'
  AND variant_sku = 'TEST-VARIANT-001';
```
