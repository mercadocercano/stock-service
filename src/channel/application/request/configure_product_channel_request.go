package request

import "errors"

// ConfigureProductChannelRequest representa la petición para configurar un canal
type ConfigureProductChannelRequest struct {
	VariantSKU string `json:"variant_sku" binding:"required"`
	Channel    string `json:"channel" binding:"required"` // "POS" | "MARKETPLACE"

	Enabled     bool `json:"enabled"`
	ManageStock bool `json:"manage_stock"`

	// MarketplaceQuota es opcional
	// Si es nil, marketplace usa todo el stock físico disponible
	MarketplaceQuota *int `json:"marketplace_quota,omitempty"`
}

// Validate valida la petición
func (r *ConfigureProductChannelRequest) Validate() error {
	if r.VariantSKU == "" {
		return errors.New("variant_sku is required")
	}

	if r.Channel == "" {
		return errors.New("channel is required")
	}

	if r.Channel != "POS" && r.Channel != "MARKETPLACE" {
		return errors.New("channel must be POS or MARKETPLACE")
	}

	// Validación adicional: quota solo para marketplace
	if r.Channel != "MARKETPLACE" && r.MarketplaceQuota != nil {
		return errors.New("marketplace_quota only applies to MARKETPLACE channel")
	}

	// Validación: quota no puede ser negativa
	if r.MarketplaceQuota != nil && *r.MarketplaceQuota < 0 {
		return errors.New("marketplace_quota cannot be negative")
	}

	// Validación crítica: marketplace debe manejar stock
	if r.Channel == "MARKETPLACE" && !r.ManageStock {
		return errors.New("marketplace channel must manage stock")
	}

	return nil
}
