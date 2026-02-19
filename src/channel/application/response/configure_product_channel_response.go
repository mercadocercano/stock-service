package response

import "time"

// ConfigureProductChannelResponse representa la respuesta de configuración
type ConfigureProductChannelResponse struct {
	TenantID   string `json:"tenant_id"`
	VariantSKU string `json:"variant_sku"`
	Channel    string `json:"channel"`

	Enabled     bool `json:"enabled"`
	ManageStock bool `json:"manage_stock"`

	MarketplaceQuota *int `json:"marketplace_quota,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
