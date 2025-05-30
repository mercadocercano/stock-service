package repository

import (
	"context"

	"stock/src/location/domain/entity"
)

// LocationRepository define la interfaz del repositorio de ubicaciones
type LocationRepository interface {
	Create(ctx context.Context, location *entity.Location) error
	Update(ctx context.Context, location *entity.Location) error
	Delete(ctx context.Context, id string, tenantID string) error
	GetByID(ctx context.Context, id string, tenantID string) (*entity.Location, error)
	FindByCriteria(ctx context.Context, criteria LocationCriteria) ([]*entity.Location, error)
}

// LocationCriteria define los criterios de búsqueda para ubicaciones
type LocationCriteria struct {
	TenantID string  `json:"tenant_id"`
	Name     *string `json:"name,omitempty"`
	Code     *string `json:"code,omitempty"`
	Active   *bool   `json:"active,omitempty"`
}
