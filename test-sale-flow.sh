#!/bin/bash

# Script de prueba para el flujo de venta mínimo

# Configuración
BASE_URL="http://localhost:8001/stock/api/v1"
TENANT_ID="00000000-0000-0000-0000-000000000001"
VARIANT_SKU="TEST-VARIANT-001"

echo "=========================================="
echo "Test de Flujo de Venta Mínimo"
echo "=========================================="
echo ""

# Paso 1: Crear stock inicial
echo "1. Creando stock inicial de 10 unidades..."
CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/stock-entries" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "variant_sku": "'$VARIANT_SKU'",
    "entry_type": "initial_stock",
    "quantity": 10,
    "product_name": "Producto de Prueba"
  }')

echo "Response:"
echo "$CREATE_RESPONSE" | jq '.'
echo ""

# Paso 2: Verificar disponibilidad inicial
echo "2. Verificando disponibilidad inicial..."
AVAILABILITY_RESPONSE=$(curl -s -X GET "$BASE_URL/availability?sku=$VARIANT_SKU" \
  -H "X-Tenant-ID: $TENANT_ID")

echo "Response:"
echo "$AVAILABILITY_RESPONSE" | jq '.'
INITIAL_STOCK=$(echo "$AVAILABILITY_RESPONSE" | jq -r '.available_quantity')
echo "Stock disponible: $INITIAL_STOCK"
echo ""

# Paso 3: Procesar venta de 3 unidades
echo "3. Procesando venta de 3 unidades..."
SALE_RESPONSE=$(curl -s -X POST "$BASE_URL/sale" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "variant_sku": "'$VARIANT_SKU'",
    "quantity": 3
  }')

echo "Response:"
echo "$SALE_RESPONSE" | jq '.'
echo ""

# Paso 4: Verificar disponibilidad después de la venta
echo "4. Verificando disponibilidad después de la venta..."
AVAILABILITY_AFTER=$(curl -s -X GET "$BASE_URL/availability?sku=$VARIANT_SKU" \
  -H "X-Tenant-ID: $TENANT_ID")

echo "Response:"
echo "$AVAILABILITY_AFTER" | jq '.'
FINAL_STOCK=$(echo "$AVAILABILITY_AFTER" | jq -r '.available_quantity')
echo "Stock disponible: $FINAL_STOCK"
echo ""

# Validación final
echo "=========================================="
echo "RESUMEN"
echo "=========================================="
echo "Stock inicial: $INITIAL_STOCK"
echo "Cantidad vendida: 3"
echo "Stock final: $FINAL_STOCK"
echo ""

if [ "$FINAL_STOCK" == "7" ]; then
  echo "✅ PRUEBA EXITOSA: El stock se descontó correctamente (10 - 3 = 7)"
else
  echo "❌ PRUEBA FALLIDA: Se esperaba 7, pero se obtuvo $FINAL_STOCK"
fi
echo ""

# Paso 5: Intentar venta sin stock suficiente
echo "5. Intentando venta de 100 unidades (debe fallar)..."
SALE_FAIL_RESPONSE=$(curl -s -X POST "$BASE_URL/sale" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "variant_sku": "'$VARIANT_SKU'",
    "quantity": 100
  }')

echo "Response:"
echo "$SALE_FAIL_RESPONSE" | jq '.'
SUCCESS=$(echo "$SALE_FAIL_RESPONSE" | jq -r '.success')

if [ "$SUCCESS" == "false" ]; then
  echo "✅ Validación correcta: No permite venta sin stock suficiente"
else
  echo "❌ Error: Debería haber rechazado la venta"
fi
echo ""
