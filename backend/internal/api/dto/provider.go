// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package dto

import "encoding/json"

// ProviderConfigRequest is the DTO for creating/updating a provider config.
type ProviderConfigRequest struct {
	ProviderType string          `json:"provider_type" binding:"required,oneof=llm image_search image_gen search storage crawler"`
	ProviderName string          `json:"provider_name" binding:"required,min=1,max=64"`
	ConfigJSON   json.RawMessage `json:"config_json" binding:"required"`
	SecretJSON   string          `json:"secret_json,omitempty"`
	Status       *int16          `json:"status,omitempty" binding:"omitempty,min=0,max=1"`
}

// ProviderConfigVO is the view object for provider config (secrets masked).
type ProviderConfigVO struct {
	ID           string          `json:"id"`
	ProviderType string          `json:"provider_type"`
	ProviderName string          `json:"provider_name"`
	ConfigJSON   json.RawMessage `json:"config_json"`
	HasSecret    bool            `json:"has_secret"`
	Status       int16           `json:"status"`
	IsDefault    bool            `json:"is_default"`
}

// WechatAccountRequest is the DTO for creating/updating a WeChat account.
type WechatAccountRequest struct {
	Name         string `json:"name" binding:"required,min=1,max=64"`
	AppID        string `json:"app_id" binding:"required,min=1,max=64"`
	AppSecret    string `json:"app_secret,omitempty"`
	TokenMode    string `json:"token_mode" binding:"required,oneof=direct stable gateway_v2"`
	DeliveryMode string `json:"delivery_mode" binding:"omitempty,oneof=api extension manual"`
	IsDefault    bool   `json:"is_default"`
	Remark       string `json:"remark" binding:"omitempty,max=500"`
}

// ExtensionTokenIssueRequest is the body for POST /extension-tokens.
type ExtensionTokenIssueRequest struct {
	Name      string `json:"name" binding:"omitempty,max=64"`
	TTLHours  int    `json:"ttl_hours,omitempty" binding:"omitempty,min=0,max=8760"`
}

// ExtensionTokenVO is the view object for an extension token (token plaintext NEVER included).
type ExtensionTokenVO struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	TokenPrefix string  `json:"token_prefix"`
	LastUsedAt  *string `json:"last_used_at,omitempty"`
	ExpiresAt   *string `json:"expires_at,omitempty"`
	RevokedAt   *string `json:"revoked_at,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

// ExtensionTokenIssueResponse is returned after a successful token issuance.
// The plaintext token is shown to the user EXACTLY ONCE here.
type ExtensionTokenIssueResponse struct {
	Token string           `json:"token"`
	Info  ExtensionTokenVO `json:"info"`
}

// WechatAccountVO is the view object for WeChat account (secrets masked).
type WechatAccountVO struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	AppID        string `json:"app_id"`
	TokenMode    string `json:"token_mode"`
	DeliveryMode string `json:"delivery_mode"`
	IsDefault    bool   `json:"is_default"`
	Status       int16  `json:"status"`
	Remark       string `json:"remark"`
}
