package draft

import (
	"gorm.io/datatypes"

	"readbud/internal/domain"
)

// ArticleDraft represents the article_drafts table per spec Section 11.6.
type ArticleDraft struct {
	domain.BaseModel
	TaskID           int64          `gorm:"type:bigint;not null;index:idx_draft_task_version,priority:1" json:"task_id"`
	WechatAccountID  *int64         `gorm:"type:bigint" json:"wechat_account_id,omitempty"`
	Title            string         `gorm:"type:varchar(255);not null" json:"title"`
	Subtitle         *string        `gorm:"type:varchar(255)" json:"subtitle,omitempty"`
	Digest           string         `gorm:"type:varchar(512)" json:"digest"`
	AuthorName       string         `gorm:"type:varchar(64)" json:"author_name"`
	ContentSourceURL *string        `gorm:"type:text" json:"content_source_url,omitempty"`
	CoverAssetID     *int64         `gorm:"type:bigint" json:"cover_asset_id,omitempty"`
	CompiledHTML     string         `gorm:"type:text" json:"-"`
	OutlineJSON      datatypes.JSON `gorm:"type:jsonb" json:"outline_json,omitempty"`
	QualityScore     float64        `gorm:"type:decimal(10,2);not null;default:0" json:"quality_score"`
	SimilarityScore  float64        `gorm:"type:decimal(10,2);not null;default:0" json:"similarity_score"`
	RiskLevel        string         `gorm:"type:varchar(32);not null;default:'low'" json:"risk_level"`
	ReviewStatus     string         `gorm:"type:varchar(32);not null;default:'pending'" json:"review_status"`
	Version          int            `gorm:"type:int;not null;default:1;index:idx_draft_task_version,priority:2,sort:desc" json:"version"`
	StyleUsed        string         `gorm:"type:varchar(32)" json:"style_used"`
	OpeningType      string         `gorm:"type:varchar(32)" json:"opening_type"`
	TitlePattern     string         `gorm:"type:varchar(32)" json:"title_pattern"`
	CTAType          string         `gorm:"type:varchar(32)" json:"cta_type"`
}

// TableName overrides the default table name.
func (ArticleDraft) TableName() string {
	return "article_drafts"
}

// Review status constants.
const (
	ReviewPending = "pending"
	ReviewPass    = "pass"
	ReviewReject  = "reject"
)

// Risk level constants.
const (
	RiskLow    = "low"
	RiskMedium = "medium"
	RiskHigh   = "high"
)
