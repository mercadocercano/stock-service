package entity_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/src/channel/domain/entity"
)

// ========== Tests de Creación y Validación ==========

func TestNewProductChannelConfig_Success(t *testing.T) {
	tenantID := uuid.New()
	quota := 10

	config, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelMarketplace,
		true,
		true,
		&quota,
	)

	require.NoError(t, err)
	assert.Equal(t, tenantID, config.TenantID)
	assert.Equal(t, "SKU-001", config.VariantSKU)
	assert.Equal(t, entity.ChannelMarketplace, config.Channel)
	assert.True(t, config.Enabled)
	assert.True(t, config.ManageStock)
	assert.NotNil(t, config.MarketplaceQuota)
	assert.Equal(t, 10, *config.MarketplaceQuota)
}

func TestNewProductChannelConfig_MarketplaceMustManageStock(t *testing.T) {
	tenantID := uuid.New()

	// Intentar crear marketplace con manageStock = false (inválido)
	_, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelMarketplace,
		true,
		false, // ❌ Marketplace SIEMPRE debe manejar stock
		nil,
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "marketplace must manage stock")
}

func TestNewProductChannelConfig_QuotaCannotBeNegative(t *testing.T) {
	tenantID := uuid.New()
	negativeQuota := -5

	_, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelMarketplace,
		true,
		true,
		&negativeQuota,
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be negative")
}

func TestNewProductChannelConfig_QuotaOnlyForMarketplace(t *testing.T) {
	tenantID := uuid.New()
	quota := 10

	// Intentar poner quota en canal POS (inválido)
	_, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelPOS,
		true,
		true,
		&quota, // ❌ Quota solo aplica a marketplace
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "only applies to marketplace")
}

func TestNewProductChannelConfig_InvalidChannel(t *testing.T) {
	tenantID := uuid.New()

	_, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.Channel("INVALID"),
		true,
		true,
		nil,
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid channel")
}

func TestNewProductChannelConfig_RequiredFields(t *testing.T) {
	tests := []struct {
		name      string
		tenantID  uuid.UUID
		sku       string
		expectErr string
	}{
		{
			name:      "missing tenant_id",
			tenantID:  uuid.Nil,
			sku:       "SKU-001",
			expectErr: "tenant_id is required",
		},
		{
			name:      "missing variant_sku",
			tenantID:  uuid.New(),
			sku:       "",
			expectErr: "variant_sku is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := entity.NewProductChannelConfig(
				tt.tenantID,
				tt.sku,
				entity.ChannelPOS,
				true,
				false,
				nil,
			)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectErr)
		})
	}
}

// ========== Tests de Lógica de Negocio ==========

func TestAvailableForMarketplace_WithoutQuota_UsesPhysicalStock(t *testing.T) {
	tenantID := uuid.New()

	config, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelMarketplace,
		true,
		true,
		nil, // Sin quota → usa todo el stock físico
	)

	require.NoError(t, err)

	// Stock físico = 10, sin quota
	available := config.AvailableForMarketplace(10)
	assert.Equal(t, 10, available, "Should use all physical stock when quota is nil")
}

func TestAvailableForMarketplace_QuotaLimitsSales(t *testing.T) {
	tenantID := uuid.New()
	quota := 3

	config, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelMarketplace,
		true,
		true,
		&quota,
	)

	require.NoError(t, err)

	// Stock físico = 10, quota = 3
	// Debe retornar min(10, 3) = 3
	available := config.AvailableForMarketplace(10)
	assert.Equal(t, 3, available, "Should limit to quota when quota < physical stock")
}

func TestAvailableForMarketplace_PhysicalStockLimits(t *testing.T) {
	tenantID := uuid.New()
	quota := 10

	config, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelMarketplace,
		true,
		true,
		&quota,
	)

	require.NoError(t, err)

	// Stock físico = 3, quota = 10
	// Debe retornar min(3, 10) = 3
	available := config.AvailableForMarketplace(3)
	assert.Equal(t, 3, available, "Should limit to physical stock when physical < quota")
}

func TestAvailableForMarketplace_ZeroStock(t *testing.T) {
	tenantID := uuid.New()
	quota := 10

	config, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelMarketplace,
		true,
		true,
		&quota,
	)

	require.NoError(t, err)

	available := config.AvailableForMarketplace(0)
	assert.Equal(t, 0, available, "Should return 0 when physical stock is 0")
}

func TestAvailableForMarketplace_DisabledChannel(t *testing.T) {
	tenantID := uuid.New()
	quota := 10

	config, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelMarketplace,
		false, // Disabled
		true,
		&quota,
	)

	require.NoError(t, err)

	available := config.AvailableForMarketplace(100)
	assert.Equal(t, 0, available, "Should return 0 when channel is disabled")
}

func TestAvailableForMarketplace_POSChannel(t *testing.T) {
	tenantID := uuid.New()

	config, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelPOS,
		true,
		false,
		nil,
	)

	require.NoError(t, err)

	// POS no usa AvailableForMarketplace
	available := config.AvailableForMarketplace(10)
	assert.Equal(t, 0, available, "POS channel should return 0 for marketplace availability")
}

func TestRequiresStockValidation_POS_WithoutStock(t *testing.T) {
	tenantID := uuid.New()

	config, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelPOS,
		true,
		false, // POS puede vender sin validar stock
		nil,
	)

	require.NoError(t, err)

	assert.False(t, config.RequiresStockValidation(), "POS with manageStock=false should not require validation")
}

func TestRequiresStockValidation_POS_WithStock(t *testing.T) {
	tenantID := uuid.New()

	config, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelPOS,
		true,
		true, // POS puede optar por validar stock
		nil,
	)

	require.NoError(t, err)

	assert.True(t, config.RequiresStockValidation(), "POS with manageStock=true should require validation")
}

func TestRequiresStockValidation_Marketplace(t *testing.T) {
	tenantID := uuid.New()

	config, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelMarketplace,
		true,
		true, // Marketplace SIEMPRE valida
		nil,
	)

	require.NoError(t, err)

	assert.True(t, config.RequiresStockValidation(), "Marketplace always requires stock validation")
}

// ========== Tests de Mutaciones ==========

func TestUpdateQuota_Success(t *testing.T) {
	tenantID := uuid.New()
	initialQuota := 10

	config, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelMarketplace,
		true,
		true,
		&initialQuota,
	)

	require.NoError(t, err)

	// Actualizar quota
	newQuota := 20
	err = config.UpdateQuota(&newQuota)

	require.NoError(t, err)
	assert.Equal(t, 20, *config.MarketplaceQuota)
}

func TestUpdateQuota_OnlyMarketplace(t *testing.T) {
	tenantID := uuid.New()

	config, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelPOS,
		true,
		false,
		nil,
	)

	require.NoError(t, err)

	quota := 10
	err = config.UpdateQuota(&quota)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "only update quota for marketplace")
}

func TestUpdateQuota_CannotBeNegative(t *testing.T) {
	tenantID := uuid.New()
	initialQuota := 10

	config, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelMarketplace,
		true,
		true,
		&initialQuota,
	)

	require.NoError(t, err)

	negativeQuota := -5
	err = config.UpdateQuota(&negativeQuota)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be negative")
}

func TestEnableDisable(t *testing.T) {
	tenantID := uuid.New()

	config, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelPOS,
		false, // Inicialmente disabled
		false,
		nil,
	)

	require.NoError(t, err)
	assert.False(t, config.CanSell())

	// Habilitar
	config.Enable()
	assert.True(t, config.CanSell())

	// Deshabilitar
	config.Disable()
	assert.False(t, config.CanSell())
}

func TestSetManageStock_POSCanToggle(t *testing.T) {
	tenantID := uuid.New()

	config, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelPOS,
		true,
		false,
		nil,
	)

	require.NoError(t, err)

	// POS puede activar manageStock
	err = config.SetManageStock(true)
	require.NoError(t, err)
	assert.True(t, config.ManageStock)

	// POS puede desactivar manageStock
	err = config.SetManageStock(false)
	require.NoError(t, err)
	assert.False(t, config.ManageStock)
}

func TestSetManageStock_MarketplaceCannotDisable(t *testing.T) {
	tenantID := uuid.New()

	config, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelMarketplace,
		true,
		true,
		nil,
	)

	require.NoError(t, err)

	// Marketplace NO puede desactivar manageStock
	err = config.SetManageStock(false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "marketplace must manage stock")
}

// ========== Test de Escenarios Reales ==========

func TestRealScenario_MarketplaceQuotaLimitsButPOSCanSellMore(t *testing.T) {
	tenantID := uuid.New()
	marketplaceQuota := 2

	// Configuración Marketplace
	marketplaceConfig, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelMarketplace,
		true,
		true,
		&marketplaceQuota,
	)
	require.NoError(t, err)

	// Configuración POS
	posConfig, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelPOS,
		true,
		true, // POS también maneja stock
		nil,
	)
	require.NoError(t, err)

	// Stock físico = 5
	physicalStock := 5

	// Marketplace: puede vender min(5, 2) = 2
	marketplaceAvailable := marketplaceConfig.AvailableForMarketplace(physicalStock)
	assert.Equal(t, 2, marketplaceAvailable)

	// POS: puede vender las 5 (usa stock físico completo)
	assert.True(t, posConfig.RequiresStockValidation())
	// POS no tiene límite de quota, usa ProcessSaleAtomic directo contra stock físico
}

func TestRealScenario_POSWithoutStockManagement(t *testing.T) {
	tenantID := uuid.New()

	// POS configurado para NO manejar stock
	posConfig, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelPOS,
		true,
		false, // No valida stock
		nil,
	)
	require.NoError(t, err)

	// POS no requiere validación de stock
	assert.False(t, posConfig.RequiresStockValidation())
	assert.True(t, posConfig.CanSell())

	// Puede vender incluso con stock físico = 0
	// (el UseCase no llamará a stock-service)
}
