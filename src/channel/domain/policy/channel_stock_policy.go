package policy

import "stock/src/channel/domain/entity"

// ChannelStockPolicy define las reglas de coordinación entre canales
// para garantizar consistencia del stock físico
type ChannelStockPolicy struct{}

// NewChannelStockPolicy crea una nueva instancia de la política
func NewChannelStockPolicy() *ChannelStockPolicy {
	return &ChannelStockPolicy{}
}

// MustManageStock determina si un canal DEBE validar stock físico
// 
// INVARIANTE DEL SISTEMA:
// Si un producto está habilitado para Marketplace, TODOS los canales
// deben respetar el stock físico para evitar sobreventa.
//
// Regla de negocio:
// - Si marketplace está habilitado → FORZAR stock management en todos los canales
// - Si marketplace NO está habilitado → respetar configuración original del canal
//
// Esta regla protege la integridad del stock cuando hay canales mixtos.
func (p *ChannelStockPolicy) MustManageStock(
	channelConfig *entity.ProductChannelConfig,
	marketplaceConfig *entity.ProductChannelConfig,
) bool {
	// Regla crítica: Si marketplace está habilitado para este producto,
	// ningún canal puede ignorar el stock físico
	if marketplaceConfig != nil &&
		marketplaceConfig.Enabled &&
		marketplaceConfig.Channel == entity.ChannelMarketplace {
		// FORZAR stock management para proteger marketplace
		return true
	}

	// Si marketplace no está habilitado, respetar configuración original
	return channelConfig.ManageStock
}

// CanSellWithoutStock verifica si un canal puede vender sin validar stock
// Es el inverso de MustManageStock, útil para lógica de negocio
func (p *ChannelStockPolicy) CanSellWithoutStock(
	channelConfig *entity.ProductChannelConfig,
	marketplaceConfig *entity.ProductChannelConfig,
) bool {
	return !p.MustManageStock(channelConfig, marketplaceConfig)
}

// GetMarketplaceAvailability calcula stock disponible para marketplace
// considerando quota y stock físico
//
// Regla: available_marketplace = min(physical_stock, marketplace_quota)
//
// Si quota es nil, usa todo el stock físico disponible.
func (p *ChannelStockPolicy) GetMarketplaceAvailability(
	marketplaceConfig *entity.ProductChannelConfig,
	physicalStock int,
) int {
	if marketplaceConfig == nil || !marketplaceConfig.Enabled {
		return 0
	}

	// Delegar a la entidad que ya tiene esta lógica
	return marketplaceConfig.AvailableForMarketplace(physicalStock)
}

// IsMarketplaceEnabled verifica si marketplace está habilitado para un producto
func (p *ChannelStockPolicy) IsMarketplaceEnabled(
	marketplaceConfig *entity.ProductChannelConfig,
) bool {
	return marketplaceConfig != nil &&
		marketplaceConfig.Enabled &&
		marketplaceConfig.Channel == entity.ChannelMarketplace
}

// ValidateChannelConsistency verifica que las configuraciones de canales
// sean consistentes entre sí
//
// Actualmente no hay reglas de inconsistencia, pero este método existe
// para extensibilidad futura (ej: si agregamos reglas como "B2B no puede
// coexistir con Marketplace" o similar)
func (p *ChannelStockPolicy) ValidateChannelConsistency(
	configs []*entity.ProductChannelConfig,
) error {
	// Por ahora, todas las combinaciones de canales son válidas
	// Este método está preparado para reglas futuras
	return nil
}
