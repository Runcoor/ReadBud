package service

import (
	"context"
	"fmt"

	"readbud/internal/api/dto"
	"readbud/internal/domain"
	"readbud/internal/pkg/crypto"
	"readbud/internal/repository/postgres"
)

// WechatAccountService handles WeChat account business logic.
type WechatAccountService struct {
	repo   postgres.WechatAccountRepository
	encKey []byte
}

// NewWechatAccountService creates a new WechatAccountService.
func NewWechatAccountService(repo postgres.WechatAccountRepository, encSecret string) *WechatAccountService {
	return &WechatAccountService{
		repo:   repo,
		encKey: crypto.DeriveKey(encSecret),
	}
}

// List returns all WeChat accounts with secrets masked.
func (s *WechatAccountService) List(ctx context.Context) ([]dto.WechatAccountVO, error) {
	accounts, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("wechatAccountService.List: %w", err)
	}

	vos := make([]dto.WechatAccountVO, 0, len(accounts))
	for _, acct := range accounts {
		vos = append(vos, toWechatVO(acct))
	}
	return vos, nil
}

// Create creates a new WeChat account.
func (s *WechatAccountService) Create(ctx context.Context, req dto.WechatAccountRequest) (*dto.WechatAccountVO, error) {
	acct := domain.WechatAccount{
		Name:      req.Name,
		AppID:     req.AppID,
		TokenMode: req.TokenMode,
		Status:    domain.StatusActive,
		Remark:    req.Remark,
	}

	if req.AppSecret != "" {
		encrypted, err := crypto.Encrypt(s.encKey, req.AppSecret)
		if err != nil {
			return nil, fmt.Errorf("wechatAccountService.Create: encrypt: %w", err)
		}
		acct.AppSecretEnc = encrypted
	}

	// Handle default flag — clear other defaults in a transaction-safe way
	if req.IsDefault {
		if err := s.repo.ClearDefault(ctx); err != nil {
			return nil, fmt.Errorf("wechatAccountService.Create: clear default: %w", err)
		}
		acct.IsDefault = 1
	}

	if err := s.repo.Create(ctx, &acct); err != nil {
		return nil, fmt.Errorf("wechatAccountService.Create: %w", err)
	}

	vo := toWechatVO(acct)
	return &vo, nil
}

// Update updates an existing WeChat account.
func (s *WechatAccountService) Update(ctx context.Context, publicID string, req dto.WechatAccountRequest) (*dto.WechatAccountVO, error) {
	acct, err := s.repo.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("wechatAccountService.Update: %w", err)
	}
	if acct == nil {
		return nil, ErrNotFound
	}

	acct.Name = req.Name
	acct.AppID = req.AppID
	acct.TokenMode = req.TokenMode
	acct.Remark = req.Remark

	if req.AppSecret != "" {
		encrypted, err := crypto.Encrypt(s.encKey, req.AppSecret)
		if err != nil {
			return nil, fmt.Errorf("wechatAccountService.Update: encrypt: %w", err)
		}
		acct.AppSecretEnc = encrypted
	}

	if req.IsDefault {
		if err := s.repo.ClearDefault(ctx); err != nil {
			return nil, fmt.Errorf("wechatAccountService.Update: clear default: %w", err)
		}
		acct.IsDefault = 1
	} else {
		acct.IsDefault = 0
	}

	if err := s.repo.Update(ctx, acct); err != nil {
		return nil, fmt.Errorf("wechatAccountService.Update: %w", err)
	}

	vo := toWechatVO(*acct)
	return &vo, nil
}

func toWechatVO(acct domain.WechatAccount) dto.WechatAccountVO {
	return dto.WechatAccountVO{
		ID:        acct.PublicID,
		Name:      acct.Name,
		AppID:     acct.AppID,
		TokenMode: acct.TokenMode,
		IsDefault: acct.IsDefault == 1,
		Status:    acct.Status,
		Remark:    acct.Remark,
	}
}
