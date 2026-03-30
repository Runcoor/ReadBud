package domain

import (
	"gorm.io/datatypes"
)

// ProviderConfig represents the provider_configs table per spec Section 11.2.
type ProviderConfig struct {
	BaseModel
	ProviderType string         `gorm:"type:varchar(32);not null;index" json:"provider_type"`
	ProviderName string         `gorm:"type:varchar(64);not null" json:"provider_name"`
	ConfigJSON   datatypes.JSON `gorm:"type:jsonb" json:"config_json"`
	SecretJSONEnc string        `gorm:"type:text" json:"-"`
	Status       int16          `gorm:"type:smallint;not null;default:1" json:"status"`
}

// TableName overrides the default table name.
func (ProviderConfig) TableName() string {
	return "provider_configs"
}

// Provider type constants.
const (
	ProviderTypeLLM         = "llm"
	ProviderTypeImageSearch = "image_search"
	ProviderTypeImageGen    = "image_gen"
	ProviderTypeSearch      = "search"
	ProviderTypeStorage     = "storage"
	ProviderTypeCrawler     = "crawler"
)
