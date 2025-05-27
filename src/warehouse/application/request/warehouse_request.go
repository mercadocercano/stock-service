package request

// CreateWarehouseRequest representa la solicitud para crear un nuevo almacén
type CreateWarehouseRequest struct {
	TenantID    string `json:"tenant_id" binding:"required"`
	LocationID  string `json:"location_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Type        string `json:"type" binding:"required,oneof=regular special virtual"`
	Description string `json:"description"`
	Priority    int    `json:"priority" binding:"gte=0"`
}

// UpdateWarehouseRequest representa la solicitud para actualizar un almacén
type UpdateWarehouseRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Type        string `json:"type" binding:"required,oneof=regular special virtual"`
	Description string `json:"description"`
	Priority    int    `json:"priority" binding:"gte=0"`
}
