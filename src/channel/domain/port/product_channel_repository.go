package port

import (
	"context"

	"github.com/google/uuid"

	"stock/src/channel/domain/entity"
)

// ProductChannelRepository define las operaciones del repositorio de configuración de canales
type ProductChannelRepository interface {
	// Save guarda o actualiza una configuración de canal
	Save(ctx context.Context, config *entity.ProductChannelConfig) error

	// FindByTenantSKUAndChannel busca configuración específica por canal
	FindByTenantSKUAndChannel(ctx context.Context, tenantID uuid.UUID, variantSKU string, channel entity.Channel) (*entity.ProductChannelConfig, error)

	// FindByTenantAndSKU retorna todas las configuraciones de un producto (POS + Marketplace)
	FindByTenantAndSKU(ctx context.Context, tenantID uuid.UUID, variantSKU string) ([]*entity.ProductChannelConfig, error)

	// Delete elimina una configuración de canal
	Delete(ctx context.Context, tenantID uuid.UUID, variantSKU string, channel entity.Channel) error

	// ExistsByTenantSKUAndChannel verifica si existe configuración
	ExistsByTenantSKUAndChannel(ctx context.Context, tenantID uuid.UUID, variantSKU string, channel entity.Channel) (bool, error)
}
