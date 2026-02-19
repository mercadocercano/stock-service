package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"stock/src/channel/application/request"
	"stock/src/channel/application/response"
	"stock/src/channel/domain/entity"
	"stock/src/channel/domain/port"
)

// ConfigureProductChannelUseCase configura cómo se maneja stock por canal de venta
type ConfigureProductChannelUseCase struct {
	repo port.ProductChannelRepository
}

// NewConfigureProductChannelUseCase crea una nueva instancia del caso de uso
func NewConfigureProductChannelUseCase(repo port.ProductChannelRepository) *ConfigureProductChannelUseCase {
	return &ConfigureProductChannelUseCase{
		repo: repo,
	}
}

// Execute ejecuta la configuración del canal
func (uc *ConfigureProductChannelUseCase) Execute(
	ctx context.Context,
	tenantID string,
	req *request.ConfigureProductChannelRequest,
) (*response.ConfigureProductChannelResponse, error) {
	// Validar request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Parsear tenant ID
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}

	// Parsear channel
	channel := entity.Channel(req.Channel)

	// Crear entidad de dominio con reglas de negocio
	config, err := entity.NewProductChannelConfig(
		tenantUUID,
		req.VariantSKU,
		channel,
		req.Enabled,
		req.ManageStock,
		req.MarketplaceQuota,
	)
	if err != nil {
		return nil, fmt.Errorf("domain validation failed: %w", err)
	}

	// Persistir configuración
	if err := uc.repo.Save(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to save configuration: %w", err)
	}

	// Construir respuesta
	return &response.ConfigureProductChannelResponse{
		TenantID:         config.TenantID.String(),
		VariantSKU:       config.VariantSKU,
		Channel:          string(config.Channel),
		Enabled:          config.Enabled,
		ManageStock:      config.ManageStock,
		MarketplaceQuota: config.MarketplaceQuota,
		CreatedAt:        config.CreatedAt,
		UpdatedAt:        config.UpdatedAt,
	}, nil
}
