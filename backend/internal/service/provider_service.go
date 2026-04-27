// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

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
	// Auto-set is_default if this is the first provider of this type
	existing, err := s.repo.FindByType(ctx, req.ProviderType)
	if err != nil {
		return nil, fmt.Errorf("providerConfigService.Create: %w", err)
	}

	cfg := domain.ProviderConfig{
		ProviderType: req.ProviderType,
		ProviderName: req.ProviderName,
		ConfigJSON:   datatypes.JSON(req.ConfigJSON),
		Status:       domain.StatusActive,
		IsDefault:    len(existing) == 0,
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

// GetActiveByType returns the first active provider config of the given type.
// Returns nil (without error) if none is found.
func (s *ProviderConfigService) GetActiveByType(ctx context.Context, providerType string) (*domain.ProviderConfig, error) {
	configs, err := s.repo.FindByType(ctx, providerType)
	if err != nil {
		return nil, fmt.Errorf("providerConfigService.GetActiveByType: %w", err)
	}
	if len(configs) == 0 {
		return nil, nil
	}
	return &configs[0], nil
}

// DecryptSecret decrypts the secret_json_enc field of a provider config.
// Returns empty string if no secret is stored.
func (s *ProviderConfigService) DecryptSecret(_ context.Context, cfg *domain.ProviderConfig) (string, error) {
	if cfg.SecretJSONEnc == "" {
		return "", nil
	}
	decrypted, err := crypto.Decrypt(s.encKey, cfg.SecretJSONEnc)
	if err != nil {
		return "", fmt.Errorf("providerConfigService.DecryptSecret: %w", err)
	}
	return decrypted, nil
}

// SetDefault marks a provider as the default for its type, clearing the flag on others.
func (s *ProviderConfigService) SetDefault(ctx context.Context, publicID string) (*dto.ProviderConfigVO, error) {
	cfg, err := s.repo.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("providerConfigService.SetDefault: %w", err)
	}
	if cfg == nil {
		return nil, ErrNotFound
	}

	// Clear is_default for all providers of the same type
	peers, err := s.repo.FindByType(ctx, cfg.ProviderType)
	if err != nil {
		return nil, fmt.Errorf("providerConfigService.SetDefault: %w", err)
	}
	for _, p := range peers {
		if p.IsDefault {
			p.IsDefault = false
			if err := s.repo.Update(ctx, &p); err != nil {
				return nil, fmt.Errorf("providerConfigService.SetDefault: clear old default: %w", err)
			}
		}
	}

	cfg.IsDefault = true
	if err := s.repo.Update(ctx, cfg); err != nil {
		return nil, fmt.Errorf("providerConfigService.SetDefault: %w", err)
	}

	vo := s.toVO(*cfg)
	return &vo, nil
}

// Delete soft-deletes a provider config by public ID.
func (s *ProviderConfigService) Delete(ctx context.Context, publicID string) error {
	cfg, err := s.repo.FindByPublicID(ctx, publicID)
	if err != nil {
		return fmt.Errorf("providerConfigService.Delete: %w", err)
	}
	if cfg == nil {
		return ErrNotFound
	}
	if err := s.repo.Delete(ctx, cfg.ID); err != nil {
		return fmt.Errorf("providerConfigService.Delete: %w", err)
	}
	return nil
}

// FindByPublicID returns a provider config by its public ID.
func (s *ProviderConfigService) FindByPublicID(ctx context.Context, publicID string) (*domain.ProviderConfig, error) {
	cfg, err := s.repo.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("providerConfigService.FindByPublicID: %w", err)
	}
	return cfg, nil
}

func (s *ProviderConfigService) toVO(cfg domain.ProviderConfig) dto.ProviderConfigVO {
	return dto.ProviderConfigVO{
		ID:           cfg.PublicID,
		ProviderType: cfg.ProviderType,
		ProviderName: cfg.ProviderName,
		ConfigJSON:   json.RawMessage(cfg.ConfigJSON),
		HasSecret:    cfg.SecretJSONEnc != "",
		Status:       cfg.Status,
		IsDefault:    cfg.IsDefault,
	}
}
