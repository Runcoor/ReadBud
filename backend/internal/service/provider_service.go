package service

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/datatypes"

	"readbud/internal/api/dto"
	"readbud/internal/domain"
	"readbud/internal/pkg/crypto"
	"readbud/internal/repository/postgres"
)

// ProviderConfigService handles provider configuration business logic.
type ProviderConfigService struct {
	repo      postgres.ProviderConfigRepository
	encKey    []byte
}

// NewProviderConfigService creates a new ProviderConfigService.
func NewProviderConfigService(repo postgres.ProviderConfigRepository, encSecret string) *ProviderConfigService {
	return &ProviderConfigService{
		repo:   repo,
		encKey: crypto.DeriveKey(encSecret),
	}
}

// List returns all provider configs with secrets masked.
func (s *ProviderConfigService) List(ctx context.Context) ([]dto.ProviderConfigVO, error) {
	configs, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("providerConfigService.List: %w", err)
	}

	vos := make([]dto.ProviderConfigVO, 0, len(configs))
	for _, cfg := range configs {
		vos = append(vos, s.toVO(cfg))
	}
	return vos, nil
}

// Create creates a new provider config, encrypting the secret if provided.
func (s *ProviderConfigService) Create(ctx context.Context, req dto.ProviderConfigRequest) (*dto.ProviderConfigVO, error) {
	cfg := domain.ProviderConfig{
		ProviderType: req.ProviderType,
		ProviderName: req.ProviderName,
		ConfigJSON:   datatypes.JSON(req.ConfigJSON),
		Status:       domain.StatusActive,
	}

	if req.SecretJSON != "" {
		encrypted, err := crypto.Encrypt(s.encKey, req.SecretJSON)
		if err != nil {
			return nil, fmt.Errorf("providerConfigService.Create: encrypt secret: %w", err)
		}
		cfg.SecretJSONEnc = encrypted
	}

	if err := s.repo.Create(ctx, &cfg); err != nil {
		return nil, fmt.Errorf("providerConfigService.Create: %w", err)
	}

	vo := s.toVO(cfg)
	return &vo, nil
}

// Update updates an existing provider config.
func (s *ProviderConfigService) Update(ctx context.Context, publicID string, req dto.ProviderConfigRequest) (*dto.ProviderConfigVO, error) {
	cfg, err := s.repo.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("providerConfigService.Update: %w", err)
	}
	if cfg == nil {
		return nil, ErrNotFound
	}

	cfg.ProviderType = req.ProviderType
	cfg.ProviderName = req.ProviderName
	cfg.ConfigJSON = datatypes.JSON(req.ConfigJSON)

	if req.Status != nil {
		cfg.Status = *req.Status
	}

	if req.SecretJSON != "" {
		encrypted, err := crypto.Encrypt(s.encKey, req.SecretJSON)
		if err != nil {
			return nil, fmt.Errorf("providerConfigService.Update: encrypt secret: %w", err)
		}
		cfg.SecretJSONEnc = encrypted
	}

	if err := s.repo.Update(ctx, cfg); err != nil {
		return nil, fmt.Errorf("providerConfigService.Update: %w", err)
	}

	vo := s.toVO(*cfg)
	return &vo, nil
}

// GetDecryptedSecret returns the decrypted secret for a provider (used internally).
func (s *ProviderConfigService) GetDecryptedSecret(ctx context.Context, providerType string) (string, error) {
	configs, err := s.repo.FindByType(ctx, providerType)
	if err != nil {
		return "", fmt.Errorf("providerConfigService.GetDecryptedSecret: %w", err)
	}
	if len(configs) == 0 {
		return "", fmt.Errorf("providerConfigService.GetDecryptedSecret: no active %s provider", providerType)
	}

	if configs[0].SecretJSONEnc == "" {
		return "", nil
	}

	decrypted, err := crypto.Decrypt(s.encKey, configs[0].SecretJSONEnc)
	if err != nil {
		return "", fmt.Errorf("providerConfigService.GetDecryptedSecret: decrypt: %w", err)
	}
	return decrypted, nil
}

func (s *ProviderConfigService) toVO(cfg domain.ProviderConfig) dto.ProviderConfigVO {
	return dto.ProviderConfigVO{
		ID:           cfg.PublicID,
		ProviderType: cfg.ProviderType,
		ProviderName: cfg.ProviderName,
		ConfigJSON:   json.RawMessage(cfg.ConfigJSON),
		HasSecret:    cfg.SecretJSONEnc != "",
		Status:       cfg.Status,
	}
}
