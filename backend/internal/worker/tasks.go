// Package worker defines Asynq task types and payloads for the content pipeline.
package worker

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

// PipelinePayload carries data through all pipeline stages.
type PipelinePayload struct {
	TaskID   int64    `json:"task_id"`
	PublicID string   `json:"public_id"`
	Queries  []string `json:"queries,omitempty"` // From keyword expand
}

// NewPipelineTask creates an Asynq task for the given pipeline stage.
func NewPipelineTask(taskType string, payload PipelinePayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("worker.NewPipelineTask: marshal payload: %w", err)
	}
	return asynq.NewTask(taskType, data), nil
}

// ParsePipelinePayload extracts the pipeline payload from an Asynq task.
func ParsePipelinePayload(task *asynq.Task) (*PipelinePayload, error) {
	var p PipelinePayload
	if err := json.Unmarshal(task.Payload(), &p); err != nil {
		return nil, fmt.Errorf("worker.ParsePipelinePayload: %w", err)
	}
	return &p, nil
}
