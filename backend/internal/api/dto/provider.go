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
	Name      string `json:"name" binding:"required,min=1,max=64"`
	AppID     string `json:"app_id" binding:"required,min=1,max=64"`
	AppSecret string `json:"app_secret,omitempty"`
	TokenMode string `json:"token_mode" binding:"required,oneof=direct stable gateway_v2"`
	IsDefault bool   `json:"is_default"`
	Remark    string `json:"remark" binding:"omitempty,max=500"`
}

// WechatAccountVO is the view object for WeChat account (secrets masked).
type WechatAccountVO struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AppID     string `json:"app_id"`
	TokenMode string `json:"token_mode"`
	IsDefault bool   `json:"is_default"`
	Status    int16  `json:"status"`
	Remark    string `json:"remark"`
}
