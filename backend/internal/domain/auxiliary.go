// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package domain

import (
	"time"

	"gorm.io/datatypes"
)

// BrandProfile represents the brand_profiles table per spec Section 11.12.
type BrandProfile struct {
	BaseModel
	Name            string         `gorm:"type:varchar(64);not null" json:"name"`
	BrandTone       string         `gorm:"type:text" json:"brand_tone"`
	ForbiddenWords  datatypes.JSON `gorm:"type:jsonb" json:"forbidden_words"`
	PreferredWords  datatypes.JSON `gorm:"type:jsonb" json:"preferred_words"`
	CTARules        datatypes.JSON `gorm:"type:jsonb" json:"cta_rules"`
	CoverStyleRules datatypes.JSON `gorm:"type:jsonb" json:"cover_style_rules"`
	ImageStyleRules datatypes.JSON `gorm:"type:jsonb" json:"image_style_rules"`
}

// TableName overrides the default table name.
func (BrandProfile) TableName() string {
	return "brand_profiles"
}

// StyleProfile represents the style_profiles table per spec Section 11.13.
type StyleProfile struct {
	BaseModel
	Name               string         `gorm:"type:varchar(64);not null" json:"name"`
	ApplicableScene    string         `gorm:"type:varchar(255)" json:"applicable_scene"`
	OpeningTemplate    string         `gorm:"type:text" json:"opening_template"`
	StructureTemplate  datatypes.JSON `gorm:"type:jsonb" json:"structure_template"`
	ClosingTemplate    string         `gorm:"type:text" json:"closing_template"`
	SentencePreference datatypes.JSON `gorm:"type:jsonb" json:"sentence_preference"`
	TitlePreference    datatypes.JSON `gorm:"type:jsonb" json:"title_preference"`
}

// TableName overrides the default table name.
func (StyleProfile) TableName() string {
	return "style_profiles"
}

// DraftVersion represents the draft_versions table per spec Section 11.14.
type DraftVersion struct {
	BaseModel
	DraftID      int64          `gorm:"type:bigint;not null;index" json:"draft_id"`
	VersionNo    int            `gorm:"type:int;not null" json:"version_no"`
	Title        string         `gorm:"type:varchar(255)" json:"title"`
	Digest       string         `gorm:"type:varchar(512)" json:"digest"`
	BlocksJSON   datatypes.JSON `gorm:"type:jsonb" json:"blocks_json"`
	HTMLSnapshot string         `gorm:"type:text" json:"-"`
	OperatorID   *int64         `gorm:"type:bigint" json:"operator_id,omitempty"`
	ChangeReason string         `gorm:"type:varchar(255)" json:"change_reason"`
}

// TableName overrides the default table name.
func (DraftVersion) TableName() string {
	return "draft_versions"
}

// ContentCitation represents the content_citations table per spec Section 11.15.
type ContentCitation struct {
	BaseModel
	DraftID          int64  `gorm:"type:bigint;not null;index" json:"draft_id"`
	BlockID          int64  `gorm:"type:bigint;not null" json:"block_id"`
	SourceDocumentID int64  `gorm:"type:bigint;not null" json:"source_document_id"`
	CitationType     string `gorm:"type:varchar(32);not null" json:"citation_type"`
	CitationText     string `gorm:"type:text" json:"citation_text"`
	SourceLink       string `gorm:"type:text" json:"source_link"`
	SourceNote       string `gorm:"type:varchar(255)" json:"source_note"`
}

// TableName overrides the default table name.
func (ContentCitation) TableName() string {
	return "content_citations"
}

// ReviewRule represents the review_rules table per spec Section 11.16.
type ReviewRule struct {
	BaseModel
	RuleType    string `gorm:"type:varchar(32);not null" json:"rule_type"`
	RuleContent string `gorm:"type:text;not null" json:"rule_content"`
	RiskLevel   string `gorm:"type:varchar(32);not null" json:"risk_level"`
	IsEnabled   int16  `gorm:"type:smallint;not null;default:1" json:"is_enabled"`
}

// TableName overrides the default table name.
func (ReviewRule) TableName() string {
	return "review_rules"
}

// DistributionPackage represents the distribution_packages table per spec Section 11.17.
type DistributionPackage struct {
	BaseModel
	DraftID             int64  `gorm:"type:bigint;not null;uniqueIndex" json:"draft_id"`
	CommunityCopy       string `gorm:"type:text" json:"community_copy"`
	MomentsCopy         string `gorm:"type:text" json:"moments_copy"`
	SummaryCard         string `gorm:"type:text" json:"summary_card"`
	CommentGuide        string `gorm:"type:text" json:"comment_guide"`
	NextTopicSuggestion string `gorm:"type:text" json:"next_topic_suggestion"`
}

// TableName overrides the default table name.
func (DistributionPackage) TableName() string {
	return "distribution_packages"
}

// TopicLibrary represents the topic_library table per spec Section 11.18.
type TopicLibrary struct {
	BaseModel
	Keyword          string     `gorm:"type:varchar(255);not null" json:"keyword"`
	Audience         string     `gorm:"type:varchar(255)" json:"audience"`
	ArticleGoal      string     `gorm:"type:varchar(64)" json:"article_goal"`
	HistoricalScore  float64    `gorm:"type:decimal(10,2);not null;default:0" json:"historical_score"`
	LastUsedAt       *time.Time `gorm:"type:timestamptz" json:"last_used_at,omitempty"`
	RecommendWeight  float64    `gorm:"type:decimal(10,2);not null;default:0" json:"recommend_weight"`
}

// TableName overrides the default table name.
func (TopicLibrary) TableName() string {
	return "topic_library"
}
