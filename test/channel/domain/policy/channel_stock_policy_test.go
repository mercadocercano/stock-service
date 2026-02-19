package policy_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"stock/src/channel/domain/entity"
	"stock/src/channel/domain/policy"
)

// ========== Tests de MustManageStock ==========

func TestMustManageStock_MarketplaceEnabled_ForcesStockManagement(t *testing.T) {
	p := policy.NewChannelStockPolicy()
	tenantID := uuid.New()

	// POS configurado para NO manejar stock
	posConfig, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelPOS,
		true,
		false, // manage_stock = false
		nil,
	)
	require.NoError(t, err)

	// Marketplace HABILITADO
	quota := 5
	marketplaceConfig, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelMarketplace,
		true,  // enabled
		true,
		&quota,
	)
	require.NoError(t, err)

	// CRÍTICO: POS DEBE manejar stock (forzado por marketplace)
	mustManage := p.MustManageStock(posConfig, marketplaceConfig)
	assert.True(t, mustManage, "POS must manage stock when marketplace is enabled")

	// Verificación inversa
	canSellWithoutStock := p.CanSellWithoutStock(posConfig, marketplaceConfig)
	assert.False(t, canSellWithoutStock, "POS cannot sell without stock when marketplace is enabled")
}

func TestMustManageStock_MarketplaceDisabled_POSCanIgnore(t *testing.T) {
	p := policy.NewChannelStockPolicy()
	tenantID := uuid.New()

	// POS configurado para NO manejar stock
	posConfig, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelPOS,
		true,
		false, // manage_stock = false
		nil,
	)
	require.NoError(t, err)

	// Marketplace DESHABILITADO
	quota := 5
	marketplaceConfig, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelMarketplace,
		false, // disabled
		true,
		&quota,
	)
	require.NoError(t, err)

	// POS puede ignorar stock (marketplace deshabilitado)
	mustManage := p.MustManageStock(posConfig, marketplaceConfig)
	assert.False(t, mustManage, "POS can ignore stock when marketplace is disabled")

	canSellWithoutStock := p.CanSellWithoutStock(posConfig, marketplaceConfig)
	assert.True(t, canSellWithoutStock, "POS can sell without stock when marketplace is disabled")
}

func TestMustManageStock_NoMarketplaceConfig_RespectsOriginal(t *testing.T) {
	p := policy.NewChannelStockPolicy()
	tenantID := uuid.New()

	// POS configurado para NO manejar stock
	posConfig, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelPOS,
		true,
		false, // manage_stock = false
		nil,
	)
	require.NoError(t, err)

	// NO hay configuración de marketplace
	mustManage := p.MustManageStock(posConfig, nil)
	assert.False(t, mustManage, "Should respect original config when no marketplace config exists")
}

func TestMustManageStock_POSWithStockManagement_AlwaysTrue(t *testing.T) {
	p := policy.NewChannelStockPolicy()
	tenantID := uuid.New()

	// POS configurado PARA manejar stock
	posConfig, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelPOS,
		true,
		true, // manage_stock = true
		nil,
	)
	require.NoError(t, err)

	// Con o sin marketplace, POS debe manejar stock (configuración original)
	mustManage := p.MustManageStock(posConfig, nil)
	assert.True(t, mustManage, "POS with manage_stock=true always manages stock")
}

func TestMustManageStock_MarketplaceEnabledWithZeroQuota_StillForcesStockManagement(t *testing.T) {
	p := policy.NewChannelStockPolicy()
	tenantID := uuid.New()

	// POS sin stock management
	posConfig, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelPOS,
		true,
		false,
		nil,
	)
	require.NoError(t, err)

	// Marketplace habilitado con quota = 0
	// (no puede vender, pero está habilitado)
	quota := 0
	marketplaceConfig, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelMarketplace,
		true, // enabled
		true,
		&quota,
	)
	require.NoError(t, err)

	// Debe FORZAR stock management porque marketplace está ENABLED
	// (la quota no importa para esta regla)
	mustManage := p.MustManageStock(posConfig, marketplaceConfig)
	assert.True(t, mustManage, "Must force stock management even with zero quota if marketplace is enabled")
}

// ========== Tests de GetMarketplaceAvailability ==========

func TestGetMarketplaceAvailability_WithQuota(t *testing.T) {
	p := policy.NewChannelStockPolicy()
	tenantID := uuid.New()

	quota := 3
	marketplaceConfig, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelMarketplace,
		true,
		true,
		&quota,
	)
	require.NoError(t, err)

	// Stock físico = 10, quota = 3
	available := p.GetMarketplaceAvailability(marketplaceConfig, 10)
	assert.Equal(t, 3, available, "Should limit to quota when quota < physical stock")
}

func TestGetMarketplaceAvailability_WithoutQuota(t *testing.T) {
	p := policy.NewChannelStockPolicy()
	tenantID := uuid.New()

	marketplaceConfig, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelMarketplace,
		true,
		true,
		nil, // Sin quota
	)
	require.NoError(t, err)

	// Stock físico = 10, sin quota
	available := p.GetMarketplaceAvailability(marketplaceConfig, 10)
	assert.Equal(t, 10, available, "Should use all physical stock when quota is nil")
}

func TestGetMarketplaceAvailability_DisabledMarketplace(t *testing.T) {
	p := policy.NewChannelStockPolicy()
	tenantID := uuid.New()

	quota := 10
	marketplaceConfig, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelMarketplace,
		false, // disabled
		true,
		&quota,
	)
	require.NoError(t, err)

	available := p.GetMarketplaceAvailability(marketplaceConfig, 10)
	assert.Equal(t, 0, available, "Disabled marketplace should return 0")
}

func TestGetMarketplaceAvailability_NilConfig(t *testing.T) {
	p := policy.NewChannelStockPolicy()

	available := p.GetMarketplaceAvailability(nil, 10)
	assert.Equal(t, 0, available, "Nil marketplace config should return 0")
}

// ========== Tests de IsMarketplaceEnabled ==========

func TestIsMarketplaceEnabled_Enabled(t *testing.T) {
	p := policy.NewChannelStockPolicy()
	tenantID := uuid.New()

	quota := 5
	marketplaceConfig, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelMarketplace,
		true, // enabled
		true,
		&quota,
	)
	require.NoError(t, err)

	assert.True(t, p.IsMarketplaceEnabled(marketplaceConfig))
}

func TestIsMarketplaceEnabled_Disabled(t *testing.T) {
	p := policy.NewChannelStockPolicy()
	tenantID := uuid.New()

	quota := 5
	marketplaceConfig, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelMarketplace,
		false, // disabled
		true,
		&quota,
	)
	require.NoError(t, err)

	assert.False(t, p.IsMarketplaceEnabled(marketplaceConfig))
}

func TestIsMarketplaceEnabled_NilConfig(t *testing.T) {
	p := policy.NewChannelStockPolicy()

	assert.False(t, p.IsMarketplaceEnabled(nil))
}

func TestIsMarketplaceEnabled_POSChannel(t *testing.T) {
	p := policy.NewChannelStockPolicy()
	tenantID := uuid.New()

	// Configuración POS (no marketplace)
	posConfig, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelPOS,
		true,
		false,
		nil,
	)
	require.NoError(t, err)

	assert.False(t, p.IsMarketplaceEnabled(posConfig), "POS config should return false")
}

// ========== Test de Escenario Real Completo ==========

func TestRealScenario_MultiChannelCoordination(t *testing.T) {
	p := policy.NewChannelStockPolicy()
	tenantID := uuid.New()

	// Producto habilitado en AMBOS canales
	// Marketplace con quota = 2
	// POS quiere NO manejar stock
	quota := 2

	posConfig, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelPOS,
		true,
		false, // POS quiere ignorar stock
		nil,
	)
	require.NoError(t, err)

	marketplaceConfig, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelMarketplace,
		true,
		true,
		&quota,
	)
	require.NoError(t, err)

	// Stock físico actual = 5

	// 1. Marketplace disponible = min(5, 2) = 2
	marketplaceAvailable := p.GetMarketplaceAvailability(marketplaceConfig, 5)
	assert.Equal(t, 2, marketplaceAvailable)

	// 2. POS DEBE manejar stock (forzado por marketplace)
	posMustManage := p.MustManageStock(posConfig, marketplaceConfig)
	assert.True(t, posMustManage, "POS cannot ignore stock when marketplace is enabled")

	// 3. Verificar que marketplace está habilitado
	assert.True(t, p.IsMarketplaceEnabled(marketplaceConfig))

	// Conclusión del escenario:
	// - Marketplace puede vender hasta 2 unidades
	// - POS debe validar contra stock físico (forzado)
	// - Ambos canales compiten por el mismo stock físico de 5 unidades
	// - Sin sobreventa posible
}

func TestRealScenario_POSOnly_CanIgnoreStock(t *testing.T) {
	p := policy.NewChannelStockPolicy()
	tenantID := uuid.New()

	// Producto SOLO en POS (marketplace no habilitado)
	posConfig, err := entity.NewProductChannelConfig(
		tenantID,
		"SKU-001",
		entity.ChannelPOS,
		true,
		false, // No manejar stock
		nil,
	)
	require.NoError(t, err)

	// Marketplace NO configurado
	var marketplaceConfig *entity.ProductChannelConfig = nil

	// POS puede ignorar stock
	posMustManage := p.MustManageStock(posConfig, marketplaceConfig)
	assert.False(t, posMustManage, "POS can ignore stock when marketplace is not configured")

	canSellWithoutStock := p.CanSellWithoutStock(posConfig, marketplaceConfig)
	assert.True(t, canSellWithoutStock)

	// Conclusión:
	// - Producto solo POS
	// - POS puede vender sin validar stock (comportamiento libre)
}
