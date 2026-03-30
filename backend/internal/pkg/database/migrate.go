package database

import (
	"fmt"

	"gorm.io/gorm"

	"readbud/internal/domain"
	"readbud/internal/domain/asset"
	"readbud/internal/domain/draft"
	"readbud/internal/domain/metrics"
	"readbud/internal/domain/publish"
	"readbud/internal/domain/source"
	"readbud/internal/domain/task"
)

// AutoMigrate runs GORM auto-migration for all domain models.
// This is used in development. Production uses SQL migration files.
func AutoMigrate(db *gorm.DB) error {
	models := []interface{}{
		// Core tables
		&domain.User{},
		&domain.ProviderConfig{},
		&domain.WechatAccount{},
		&task.ContentTask{},
		&source.SourceDocument{},
		&draft.ArticleDraft{},
		&draft.ArticleBlock{},
		&asset.Asset{},
		&publish.PublishJob{},
		&publish.PublishRecord{},
		&metrics.MetricsSnapshot{},
		// Auxiliary tables
		&domain.BrandProfile{},
		&domain.StyleProfile{},
		&domain.DraftVersion{},
		&domain.ContentCitation{},
		&domain.ReviewRule{},
		&domain.DistributionPackage{},
		&domain.TopicLibrary{},
	}

	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("database.AutoMigrate: %w", err)
	}

	// Create additional indexes that GORM tags cannot express.
	if err := createCustomIndexes(db); err != nil {
		return fmt.Errorf("database.AutoMigrate: custom indexes: %w", err)
	}

	return nil
}

// createCustomIndexes creates indexes that need raw SQL (e.g. gin_trgm_ops).
func createCustomIndexes(db *gorm.DB) error {
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_topic_keyword_trgm
		 ON topic_library USING gin (keyword gin_trgm_ops)`,
	}

	for _, idx := range indexes {
		if err := db.Exec(idx).Error; err != nil {
			return fmt.Errorf("create index: %w", err)
		}
	}

	return nil
}
