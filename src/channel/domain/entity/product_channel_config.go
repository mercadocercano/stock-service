package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Channel representa los canales de venta disponibles
type Channel string

const (
	ChannelPOS         Channel = "POS"
	ChannelMarketplace Channel = "MARKETPLACE"
)

// ProductChannelConfig configura cómo se maneja el stock por canal de venta
// Dominio puro: sin dependencias de infraestructura
type ProductChannelConfig struct {
	TenantID   uuid.UUID
	VariantSKU string
	Channel    Channel

	// Enabled indica si el producto está habilitado para este canal
	Enabled bool

	// ManageStock indica si este canal debe validar stock físico
	// POS: puede ser false (venta sin stock)
	// Marketplace: SIEMPRE true
	ManageStock bool

	// MarketplaceQuota es el techo de venta para marketplace
	// Si es nil, usa todo el stock físico disponible
	// Si tiene valor, limita a ese número (min con stock físico)
	MarketplaceQuota *int

	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewProductChannelConfig crea una nueva configuración con validaciones de dominio
func NewProductChannelConfig(
	tenantID uuid.UUID,
	variantSKU string,
	channel Channel,
	enabled bool,
	manageStock bool,
	marketplaceQuota *int,
) (*ProductChannelConfig, error) {
	// Validaciones de dominio
	if tenantID == uuid.Nil {
		return nil, errors.New("tenant_id is required")
	}

	if variantSKU == "" {
		return nil, errors.New("variant_sku is required")
	}

	if channel != ChannelPOS && channel != ChannelMarketplace {
		return nil, errors.New("invalid channel: must be POS or MARKETPLACE")
	}

	// Regla de negocio crítica:
	// Marketplace SIEMPRE debe manejar stock
	if channel == ChannelMarketplace && !manageStock {
		return nil, errors.New("marketplace must manage stock")
	}

	// Regla: quota solo aplica a marketplace
	if channel != ChannelMarketplace && marketplaceQuota != nil {
		return nil, errors.New("marketplace_quota only applies to marketplace channel")
	}

	// Regla: quota no puede ser negativa
	if marketplaceQuota != nil && *marketplaceQuota < 0 {
		return nil, errors.New("marketplace_quota cannot be negative")
	}

	now := time.Now().UTC()

	return &ProductChannelConfig{
		TenantID:         tenantID,
		VariantSKU:       variantSKU,
		Channel:          channel,
		Enabled:          enabled,
		ManageStock:      manageStock,
		MarketplaceQuota: marketplaceQuota,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

// AvailableForMarketplace calcula cuánto stock puede vender marketplace
// Regla: min(stock_fisico, marketplace_quota)
// Dominio puro: sin llamadas a DB
func (c *ProductChannelConfig) AvailableForMarketplace(physicalStock int) int {
	if !c.Enabled {
		return 0
	}

	if c.Channel != ChannelMarketplace {
		return 0
	}

	if physicalStock <= 0 {
		return 0
	}

	// Si no hay quota configurada, usa todo el stock físico
	if c.MarketplaceQuota == nil {
		return physicalStock
	}

	// Retornar el mínimo entre stock físico y quota
	quota := *c.MarketplaceQuota
	if quota < physicalStock {
		return quota
	}

	return physicalStock
}

// RequiresStockValidation indica si este canal debe validar stock antes de vender
func (c *ProductChannelConfig) RequiresStockValidation() bool {
	return c.ManageStock
}

// CanSell indica si se puede vender en este canal
func (c *ProductChannelConfig) CanSell() bool {
	return c.Enabled
}

// UpdateQuota actualiza la quota de marketplace
// Solo válido para canal marketplace
func (c *ProductChannelConfig) UpdateQuota(newQuota *int) error {
	if c.Channel != ChannelMarketplace {
		return errors.New("can only update quota for marketplace channel")
	}

	if newQuota != nil && *newQuota < 0 {
		return errors.New("quota cannot be negative")
	}

	c.MarketplaceQuota = newQuota
	c.UpdatedAt = time.Now().UTC()

	return nil
}

// Enable habilita el producto para este canal
func (c *ProductChannelConfig) Enable() {
	c.Enabled = true
	c.UpdatedAt = time.Now().UTC()
}

// Disable deshabilita el producto para este canal
func (c *ProductChannelConfig) Disable() {
	c.Enabled = false
	c.UpdatedAt = time.Now().UTC()
}

// SetManageStock configura si debe manejar stock
// Para marketplace, siempre debe ser true (validación en constructor/update)
func (c *ProductChannelConfig) SetManageStock(manageStock bool) error {
	if c.Channel == ChannelMarketplace && !manageStock {
		return errors.New("marketplace must manage stock")
	}

	c.ManageStock = manageStock
	c.UpdatedAt = time.Now().UTC()

	return nil
}

// IsValid valida la consistencia del aggregate
func (c *ProductChannelConfig) IsValid() error {
	if c.TenantID == uuid.Nil {
		return errors.New("tenant_id is required")
	}

	if c.VariantSKU == "" {
		return errors.New("variant_sku is required")
	}

	if c.Channel != ChannelPOS && c.Channel != ChannelMarketplace {
		return errors.New("invalid channel")
	}

	if c.Channel == ChannelMarketplace && !c.ManageStock {
		return errors.New("marketplace must manage stock")
	}

	if c.Channel != ChannelMarketplace && c.MarketplaceQuota != nil {
		return errors.New("quota only applies to marketplace")
	}

	if c.MarketplaceQuota != nil && *c.MarketplaceQuota < 0 {
		return errors.New("quota cannot be negative")
	}

	return nil
}
