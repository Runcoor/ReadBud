package dto

// CreateReviewRuleRequest is the DTO for POST /api/v1/review-rules.
type CreateReviewRuleRequest struct {
	RuleType    string `json:"rule_type" binding:"required,oneof=keyword_blacklist pattern_match content_policy"`
	RuleContent string `json:"rule_content" binding:"required"`
	RiskLevel   string `json:"risk_level" binding:"required,oneof=low medium high"`
	IsEnabled   *int16 `json:"is_enabled"`
}

// UpdateReviewRuleRequest is the DTO for PUT /api/v1/review-rules/:id.
type UpdateReviewRuleRequest struct {
	RuleType    *string `json:"rule_type" binding:"omitempty,oneof=keyword_blacklist pattern_match content_policy"`
	RuleContent *string `json:"rule_content"`
	RiskLevel   *string `json:"risk_level" binding:"omitempty,oneof=low medium high"`
}

// ToggleReviewRuleRequest is the DTO for POST /api/v1/review-rules/:id/toggle.
type ToggleReviewRuleRequest struct {
	IsEnabled int16 `json:"is_enabled" binding:"oneof=0 1"`
}

// EvaluateContentRequest is the DTO for POST /api/v1/review-rules/evaluate.
type EvaluateContentRequest struct {
	Content string `json:"content" binding:"required"`
}
