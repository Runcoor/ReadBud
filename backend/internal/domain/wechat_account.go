// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package domain

import "time"

// WechatAccount represents the wechat_accounts table per spec Section 11.3.
type WechatAccount struct {
	BaseModel
	Name              string     `gorm:"type:varchar(64);not null" json:"name"`
	AppID             string     `gorm:"type:varchar(64);not null" json:"app_id"`
	AppSecretEnc      string     `gorm:"type:text;not null" json:"-"`
	TokenMode         string     `gorm:"type:varchar(32);not null;default:'direct'" json:"token_mode"`
	DeliveryMode      string     `gorm:"type:varchar(32);not null;default:'extension'" json:"delivery_mode"`
	StableAccessToken *string    `gorm:"type:text" json:"-"`
	TokenExpireAt     *time.Time `gorm:"type:timestamptz" json:"token_expire_at,omitempty"`
	GatewayUserAPIID  *string    `gorm:"type:varchar(128)" json:"gateway_user_api_id,omitempty"`
	IsDefault         int16      `gorm:"type:smallint;not null;default:0" json:"is_default"`
	Status            int16      `gorm:"type:smallint;not null;default:1" json:"status"`
	Remark            string     `gorm:"type:varchar(255)" json:"remark"`
}

// TableName overrides the default table name.
func (WechatAccount) TableName() string {
	return "wechat_accounts"
}

// Token mode constants.
const (
	TokenModeDirect  = "direct"
	TokenModeStable  = "stable"
	TokenModeGateway = "gateway_v2"
)

// Delivery mode constants — describes HOW articles reach WeChat:
//   - api:       direct draft/add + freepublish/submit (requires verified service account)
//   - extension: browser plugin auto-fills the WeChat editor (works for any account)
//   - manual:    user copies + pastes into the editor by hand (last resort)
const (
	DeliveryModeAPI       = "api"
	DeliveryModeExtension = "extension"
	DeliveryModeManual    = "manual"
)
