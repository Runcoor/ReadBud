package domain

import "time"

// WechatAccount represents the wechat_accounts table per spec Section 11.3.
type WechatAccount struct {
	BaseModel
	Name              string     `gorm:"type:varchar(64);not null" json:"name"`
	AppID             string     `gorm:"type:varchar(64);not null" json:"app_id"`
	AppSecretEnc      string     `gorm:"type:text;not null" json:"-"`
	TokenMode         string     `gorm:"type:varchar(32);not null;default:'direct'" json:"token_mode"`
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
	TokenModeDirect   = "direct"
	TokenModeStable   = "stable"
	TokenModeGateway  = "gateway_v2"
)
