package asset

import "readbud/internal/domain"

// Asset represents the assets table per spec Section 11.8.
type Asset struct {
	domain.BaseModel
	AssetType         string  `gorm:"type:varchar(32);not null" json:"asset_type"`
	SourceKind        string  `gorm:"type:varchar(32);not null" json:"source_kind"`
	MimeType          string  `gorm:"type:varchar(64);not null" json:"mime_type"`
	StorageProvider   string  `gorm:"type:varchar(32);not null" json:"storage_provider"`
	Bucket            string  `gorm:"type:varchar(128);not null" json:"bucket"`
	ObjectKey         string  `gorm:"type:varchar(512);not null" json:"object_key"`
	LocalPath         *string `gorm:"type:varchar(512)" json:"local_path,omitempty"`
	Width             *int    `gorm:"type:int" json:"width,omitempty"`
	Height            *int    `gorm:"type:int" json:"height,omitempty"`
	SizeBytes         *int64  `gorm:"type:bigint" json:"size_bytes,omitempty"`
	SHA256            string  `gorm:"type:varchar(128);not null;index" json:"sha256"`
	SourceURL         *string `gorm:"type:text" json:"source_url,omitempty"`
	SourcePageURL     *string `gorm:"type:text" json:"source_page_url,omitempty"`
	SourceSite        *string `gorm:"type:varchar(128)" json:"source_site,omitempty"`
	SourceAuthor      *string `gorm:"type:varchar(128)" json:"source_author,omitempty"`
	LicenseType       *string `gorm:"type:varchar(64)" json:"license_type,omitempty"`
	AttributionText   *string `gorm:"type:varchar(255)" json:"attribution_text,omitempty"`
	PromptText        *string `gorm:"type:text" json:"prompt_text,omitempty"`
	IsAIGenerated     int16   `gorm:"type:smallint;not null;default:0" json:"is_ai_generated"`
	WechatURL         *string `gorm:"type:text" json:"wechat_url,omitempty"`
	WechatMediaID     *string `gorm:"type:varchar(128)" json:"wechat_media_id,omitempty"`
	WechatUploadStatus string `gorm:"type:varchar(32);not null;default:'pending'" json:"wechat_upload_status"`
}

// TableName overrides the default table name.
func (Asset) TableName() string {
	return "assets"
}

// Asset type constants.
const (
	AssetTypeContentImage   = "content_image"
	AssetTypeCoverImage     = "cover_image"
	AssetTypeChart          = "chart"
	AssetTypeGeneratedImage = "generated_image"
)

// Source kind constants.
const (
	SourceKindSearch     = "search"
	SourceKindGenerated  = "generated"
	SourceKindSourcePage = "source_page"
	SourceKindUploaded   = "uploaded"
)

// WeChat upload status constants.
const (
	WechatUploadPending = "pending"
	WechatUploadDone    = "done"
	WechatUploadFailed  = "failed"
)
