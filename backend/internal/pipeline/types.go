// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

// Package pipeline defines shared task types and payloads for the content pipeline.
// This package has NO dependencies on service or worker to avoid import cycles.
package pipeline

import (
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

// Task type constants for the content pipeline stages.
const (
	TypeKeywordExpand = "pipeline:keyword_expand"
	TypeSourceSearch  = "pipeline:source_search"
	TypeContentCrawl  = "pipeline:content_crawl"
	TypeHotScore      = "pipeline:hot_score"
	TypeArticleWrite  = "pipeline:article_write"
	TypeImageMatch    = "pipeline:image_match"
	TypeChartGen      = "pipeline:chart_gen"
	TypeHTMLCompile   = "pipeline:html_compile"
	TypePublish       = "pipeline:publish"
)

// Payload carries data through all pipeline stages.
type Payload struct {
	TaskID         int64    `json:"task_id"`
	PublicID       string   `json:"public_id"`
	Queries        []string `json:"queries,omitempty"`
	SourceURLs     []string `json:"source_urls,omitempty"`
	CrawledContent string   `json:"crawled_content,omitempty"`
}

// NewTask creates an Asynq task for the given pipeline stage.
func NewTask(taskType string, payload Payload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("pipeline.NewTask: marshal: %w", err)
	}
	return asynq.NewTask(taskType, data), nil
}

// ParsePayload extracts the pipeline payload from an Asynq task.
func ParsePayload(task *asynq.Task) (*Payload, error) {
	var p Payload
	if err := json.Unmarshal(task.Payload(), &p); err != nil {
		return nil, fmt.Errorf("pipeline.ParsePayload: %w", err)
	}
	return &p, nil
}
