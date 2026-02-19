package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"stock/src/channel/application/request"
	"stock/src/channel/application/usecase"
	"stock/src/channel/domain/entity"
)

// ========== Mock Repository ==========

type MockProductChannelRepository struct {
	mock.Mock
}

func (m *MockProductChannelRepository) Save(ctx context.Context, config *entity.ProductChannelConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockProductChannelRepository) FindByTenantSKUAndChannel(ctx context.Context, tenantID uuid.UUID, variantSKU string, channel entity.Channel) (*entity.ProductChannelConfig, error) {
	args := m.Called(ctx, tenantID, variantSKU, channel)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ProductChannelConfig), args.Error(1)
}

func (m *MockProductChannelRepository) FindByTenantAndSKU(ctx context.Context, tenantID uuid.UUID, variantSKU string) ([]*entity.ProductChannelConfig, error) {
	args := m.Called(ctx, tenantID, variantSKU)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.ProductChannelConfig), args.Error(1)
}

func (m *MockProductChannelRepository) Delete(ctx context.Context, tenantID uuid.UUID, variantSKU string, channel entity.Channel) error {
	args := m.Called(ctx, tenantID, variantSKU, channel)
	return args.Error(0)
}

func (m *MockProductChannelRepository) ExistsByTenantSKUAndChannel(ctx context.Context, tenantID uuid.UUID, variantSKU string, channel entity.Channel) (bool, error) {
	args := m.Called(ctx, tenantID, variantSKU, channel)
	return args.Bool(0), args.Error(1)
}

// ========== Tests ==========

func TestConfigureProductChannelUseCase_Success_Marketplace(t *testing.T) {
	mockRepo := new(MockProductChannelRepository)
	uc := usecase.NewConfigureProductChannelUseCase(mockRepo)

	tenantID := uuid.New()
	quota := 10

	req := &request.ConfigureProductChannelRequest{
		VariantSKU:       "SKU-001",
		Channel:          "MARKETPLACE",
		Enabled:          true,
		ManageStock:      true,
		MarketplaceQuota: &quota,
	}

	// Configurar mock: Save debe tener éxito
	mockRepo.On("Save", mock.Anything, mock.MatchedBy(func(config *entity.ProductChannelConfig) bool {
		return config.VariantSKU == "SKU-001" &&
			config.Channel == entity.ChannelMarketplace &&
			config.Enabled &&
			config.ManageStock &&
			*config.MarketplaceQuota == 10
	})).Return(nil)

	// Ejecutar
	resp, err := uc.Execute(context.Background(), tenantID.String(), req)

	// Verificaciones
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "SKU-001", resp.VariantSKU)
	assert.Equal(t, "MARKETPLACE", resp.Channel)
	assert.True(t, resp.Enabled)
	assert.True(t, resp.ManageStock)
	assert.Equal(t, 10, *resp.MarketplaceQuota)

	mockRepo.AssertExpectations(t)
}

func TestConfigureProductChannelUseCase_Success_POS_WithoutStockManagement(t *testing.T) {
	mockRepo := new(MockProductChannelRepository)
	uc := usecase.NewConfigureProductChannelUseCase(mockRepo)

	tenantID := uuid.New()

	req := &request.ConfigureProductChannelRequest{
		VariantSKU:  "SKU-001",
		Channel:     "POS",
		Enabled:     true,
		ManageStock: false, // POS puede NO manejar stock
	}

	mockRepo.On("Save", mock.Anything, mock.MatchedBy(func(config *entity.ProductChannelConfig) bool {
		return config.Channel == entity.ChannelPOS &&
			!config.ManageStock &&
			config.MarketplaceQuota == nil
	})).Return(nil)

	resp, err := uc.Execute(context.Background(), tenantID.String(), req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "POS", resp.Channel)
	assert.False(t, resp.ManageStock)
	assert.Nil(t, resp.MarketplaceQuota)

	mockRepo.AssertExpectations(t)
}

func TestConfigureProductChannelUseCase_Failure_MarketplaceMustManageStock(t *testing.T) {
	mockRepo := new(MockProductChannelRepository)
	uc := usecase.NewConfigureProductChannelUseCase(mockRepo)

	tenantID := uuid.New()

	req := &request.ConfigureProductChannelRequest{
		VariantSKU:  "SKU-001",
		Channel:     "MARKETPLACE",
		Enabled:     true,
		ManageStock: false, // ❌ Inválido para marketplace
	}

	// No debería llamar a Save porque la validación falla antes
	// mockRepo.On("Save", ...) NO se configura

	_, err := uc.Execute(context.Background(), tenantID.String(), req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "must manage stock")

	// Verify que Save NUNCA fue llamado
	mockRepo.AssertNotCalled(t, "Save")
}

func TestConfigureProductChannelUseCase_Failure_NegativeQuota(t *testing.T) {
	mockRepo := new(MockProductChannelRepository)
	uc := usecase.NewConfigureProductChannelUseCase(mockRepo)

	tenantID := uuid.New()
	negativeQuota := -5

	req := &request.ConfigureProductChannelRequest{
		VariantSKU:       "SKU-001",
		Channel:          "MARKETPLACE",
		Enabled:          true,
		ManageStock:      true,
		MarketplaceQuota: &negativeQuota,
	}

	_, err := uc.Execute(context.Background(), tenantID.String(), req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be negative")

	mockRepo.AssertNotCalled(t, "Save")
}

func TestConfigureProductChannelUseCase_Failure_QuotaOnlyForMarketplace(t *testing.T) {
	mockRepo := new(MockProductChannelRepository)
	uc := usecase.NewConfigureProductChannelUseCase(mockRepo)

	tenantID := uuid.New()
	quota := 10

	req := &request.ConfigureProductChannelRequest{
		VariantSKU:       "SKU-001",
		Channel:          "POS",
		Enabled:          true,
		ManageStock:      false,
		MarketplaceQuota: &quota, // ❌ Quota solo aplica a marketplace
	}

	_, err := uc.Execute(context.Background(), tenantID.String(), req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "only applies to")

	mockRepo.AssertNotCalled(t, "Save")
}

func TestConfigureProductChannelUseCase_Failure_InvalidTenantID(t *testing.T) {
	mockRepo := new(MockProductChannelRepository)
	uc := usecase.NewConfigureProductChannelUseCase(mockRepo)

	req := &request.ConfigureProductChannelRequest{
		VariantSKU:  "SKU-001",
		Channel:     "POS",
		Enabled:     true,
		ManageStock: false,
	}

	_, err := uc.Execute(context.Background(), "invalid-uuid", req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tenant_id")

	mockRepo.AssertNotCalled(t, "Save")
}

func TestConfigureProductChannelUseCase_Failure_RepositoryError(t *testing.T) {
	mockRepo := new(MockProductChannelRepository)
	uc := usecase.NewConfigureProductChannelUseCase(mockRepo)

	tenantID := uuid.New()

	req := &request.ConfigureProductChannelRequest{
		VariantSKU:  "SKU-001",
		Channel:     "POS",
		Enabled:     true,
		ManageStock: false,
	}

	// Simular error de persistencia
	mockRepo.On("Save", mock.Anything, mock.Anything).Return(errors.New("database connection error"))

	_, err := uc.Execute(context.Background(), tenantID.String(), req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save")

	mockRepo.AssertExpectations(t)
}

func TestConfigureProductChannelRequest_Validate_Success(t *testing.T) {
	quota := 5

	req := &request.ConfigureProductChannelRequest{
		VariantSKU:       "SKU-001",
		Channel:          "MARKETPLACE",
		Enabled:          true,
		ManageStock:      true,
		MarketplaceQuota: &quota,
	}

	err := req.Validate()
	assert.NoError(t, err)
}

func TestConfigureProductChannelRequest_Validate_MissingSKU(t *testing.T) {
	req := &request.ConfigureProductChannelRequest{
		Channel:     "POS",
		Enabled:     true,
		ManageStock: false,
	}

	err := req.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "variant_sku is required")
}

func TestConfigureProductChannelRequest_Validate_MissingChannel(t *testing.T) {
	req := &request.ConfigureProductChannelRequest{
		VariantSKU:  "SKU-001",
		Enabled:     true,
		ManageStock: false,
	}

	err := req.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "channel is required")
}

func TestConfigureProductChannelRequest_Validate_InvalidChannel(t *testing.T) {
	req := &request.ConfigureProductChannelRequest{
		VariantSKU:  "SKU-001",
		Channel:     "INVALID",
		Enabled:     true,
		ManageStock: false,
	}

	err := req.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be POS or MARKETPLACE")
}
