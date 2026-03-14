package usecase

import (
	"context"
	"fmt"
	"math"

	"github.com/google/uuid"

	"stock/src/stock_entry/application/response"
	"stock/src/stock_entry/domain/port"
)

type ListAvailabilityUseCase struct {
	availabilityRepo port.StockAvailabilityRepository
}

func NewListAvailabilityUseCase(availabilityRepo port.StockAvailabilityRepository) *ListAvailabilityUseCase {
	return &ListAvailabilityUseCase{
		availabilityRepo: availabilityRepo,
	}
}

type ListAvailabilityResult struct {
	Items      []response.StockAvailabilityResponse `json:"items"`
	TotalCount int                                  `json:"total_count"`
	Page       int                                  `json:"page"`
	PageSize   int                                  `json:"page_size"`
	TotalPages int                                  `json:"total_pages"`
}

func (uc *ListAvailabilityUseCase) Execute(ctx context.Context, tenantID string, page, pageSize int) (*ListAvailabilityResult, error) {
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant_id: %w", err)
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 500 {
		pageSize = 20
	}

	totalCount, err := uc.availabilityRepo.CountByTenant(ctx, tenantUUID)
	if err != nil {
		return nil, fmt.Errorf("error counting availability: %w", err)
	}

	offset := (page - 1) * pageSize
	availabilities, err := uc.availabilityRepo.FindByTenant(ctx, tenantUUID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing availability: %w", err)
	}

	items := make([]response.StockAvailabilityResponse, 0, len(availabilities))
	for _, a := range availabilities {
		items = append(items, response.FromStockAvailability(a))
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))

	return &ListAvailabilityResult{
		Items:      items,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}
