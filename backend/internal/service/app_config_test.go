package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ai-wms/ai-wms/backend/internal/domain"
)

// mockAppConfigRepo implements repository.AppConfigRepository for testing.
type mockAppConfigRepo struct {
	config *domain.AppConfigRow
}

func (m *mockAppConfigRepo) GetAppConfig(ctx context.Context) (*domain.AppConfigRow, error) {
	if m.config == nil {
		return &domain.AppConfigRow{Config: domain.DefaultAppConfig()}, nil
	}
	return m.config, nil
}

func (m *mockAppConfigRepo) UpdateAppConfig(ctx context.Context, config domain.AppConfig) error {
	m.config = &domain.AppConfigRow{Config: config, UpdatedAt: time.Now()}
	return nil
}

func TestAppConfigService_GetConfig_Defaults(t *testing.T) {
	repo := &mockAppConfigRepo{}
	svc := NewAppConfigService(repo)

	row, err := svc.GetConfig(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if row.Config.SiteName != "AI-WMS" {
		t.Errorf("expected default site name 'AI-WMS', got %q", row.Config.SiteName)
	}
	if row.Config.LowStockThreshold != 10 {
		t.Errorf("expected default low stock threshold 10, got %d", row.Config.LowStockThreshold)
	}
	if row.Config.DefaultPageSize != 20 {
		t.Errorf("expected default page size 20, got %d", row.Config.DefaultPageSize)
	}
	if row.Config.JWTAccessTTL != 3600 {
		t.Errorf("expected default JWT TTL 3600, got %d", row.Config.JWTAccessTTL)
	}
}

func TestAppConfigService_UpdateConfig_Partial(t *testing.T) {
	repo := &mockAppConfigRepo{
		config: &domain.AppConfigRow{
			Config: domain.AppConfig{
				SiteName:           "Initial",
				DefaultWarehouseID: uuid.New().String(),
				LowStockThreshold:  5,
				DefaultPageSize:    10,
				JWTAccessTTL:       1800,
			},
		},
	}
	svc := NewAppConfigService(repo)

	newName := "Updated Site"
	newThreshold := 20

	row, err := svc.UpdateConfig(context.Background(), UpdateConfigInput{
		SiteName:          &newName,
		LowStockThreshold: &newThreshold,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Updated fields.
	if row.Config.SiteName != newName {
		t.Errorf("expected site name %q, got %q", newName, row.Config.SiteName)
	}
	if row.Config.LowStockThreshold != newThreshold {
		t.Errorf("expected threshold %d, got %d", newThreshold, row.Config.LowStockThreshold)
	}

	// Unchanged fields should retain original values.
	if row.Config.DefaultPageSize != 10 {
		t.Errorf("expected default page size 10 (unchanged), got %d", row.Config.DefaultPageSize)
	}
	if row.Config.JWTAccessTTL != 1800 {
		t.Errorf("expected JWT TTL 1800 (unchanged), got %d", row.Config.JWTAccessTTL)
	}
}

func TestAppConfigService_UpdateConfig_Validation(t *testing.T) {
	repo := &mockAppConfigRepo{}
	svc := NewAppConfigService(repo)

	tests := []struct {
		name  string
		input UpdateConfigInput
	}{
		{
			name:  "empty site name",
			input: UpdateConfigInput{SiteName: strPtr("")},
		},
		{
			name:  "negative low stock threshold",
			input: UpdateConfigInput{LowStockThreshold: intPtr(-1)},
		},
		{
			name:  "page size too small",
			input: UpdateConfigInput{DefaultPageSize: intPtr(0)},
		},
		{
			name:  "page size too large",
			input: UpdateConfigInput{DefaultPageSize: intPtr(101)},
		},
		{
			name:  "jwt ttl too short",
			input: UpdateConfigInput{JWTAccessTTL: intPtr(30)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.UpdateConfig(context.Background(), tt.input)
			if err == nil {
				t.Errorf("expected validation error for %q, got nil", tt.name)
			}
		})
	}
}

func TestAppConfigService_UpdateConfig_JSONRoundTrip(t *testing.T) {
	// Verify that config values round-trip correctly through JSON marshaling.
	repo := &mockAppConfigRepo{}
	svc := NewAppConfigService(repo)

	cfg := domain.AppConfig{
		SiteName:           "My WMS",
		DefaultWarehouseID: uuid.New().String(),
		LowStockThreshold:  15,
		DefaultPageSize:    50,
		JWTAccessTTL:       7200,
	}

	// Marshal to JSON and back to verify no field name mismatches.
	raw, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded domain.AppConfig
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded != cfg {
		t.Errorf("round-trip mismatch: got %+v, want %+v", decoded, cfg)
	}

	// Also verify through the service.
	siteName := "My WMS"
	threshold := 15
	row, err := svc.UpdateConfig(context.Background(), UpdateConfigInput{
		SiteName:          &siteName,
		LowStockThreshold: &threshold,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if row.Config.SiteName != siteName {
		t.Errorf("expected %q, got %q", siteName, row.Config.SiteName)
	}
}

// Helpers.

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }
