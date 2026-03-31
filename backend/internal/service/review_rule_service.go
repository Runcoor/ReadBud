package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"readbud/internal/domain"
	"readbud/internal/repository/postgres"
)

// RuleViolation represents a single rule violation found during content evaluation.
type RuleViolation struct {
	RulePublicID string `json:"rule_id"`
	RuleType     string `json:"rule_type"`
	RiskLevel    string `json:"risk_level"`
	Detail       string `json:"detail"`
}

// ReviewRuleService handles review rule business logic.
type ReviewRuleService struct {
	repo postgres.ReviewRuleRepository
}

// NewReviewRuleService creates a new ReviewRuleService.
func NewReviewRuleService(repo postgres.ReviewRuleRepository) *ReviewRuleService {
	return &ReviewRuleService{repo: repo}
}

// Create creates a new review rule.
func (s *ReviewRuleService) Create(ctx context.Context, rule *domain.ReviewRule) error {
	if err := s.repo.Create(ctx, rule); err != nil {
		return fmt.Errorf("reviewRuleService.Create: %w", err)
	}
	return nil
}

// Get returns a review rule by its public ID.
func (s *ReviewRuleService) Get(ctx context.Context, publicID string) (*domain.ReviewRule, error) {
	rule, err := s.repo.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("reviewRuleService.Get: %w", err)
	}
	if rule == nil {
		return nil, ErrNotFound
	}
	return rule, nil
}

// Update updates an existing review rule identified by public ID.
func (s *ReviewRuleService) Update(ctx context.Context, publicID string, ruleType *string, ruleContent *string, riskLevel *string) (*domain.ReviewRule, error) {
	rule, err := s.repo.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("reviewRuleService.Update: %w", err)
	}
	if rule == nil {
		return nil, ErrNotFound
	}

	if ruleType != nil {
		rule.RuleType = *ruleType
	}
	if ruleContent != nil {
		rule.RuleContent = *ruleContent
	}
	if riskLevel != nil {
		rule.RiskLevel = *riskLevel
	}

	if err := s.repo.Update(ctx, rule); err != nil {
		return nil, fmt.Errorf("reviewRuleService.Update: %w", err)
	}
	return rule, nil
}

// Delete deletes a review rule by its public ID.
func (s *ReviewRuleService) Delete(ctx context.Context, publicID string) error {
	rule, err := s.repo.FindByPublicID(ctx, publicID)
	if err != nil {
		return fmt.Errorf("reviewRuleService.Delete: %w", err)
	}
	if rule == nil {
		return ErrNotFound
	}
	if err := s.repo.Delete(ctx, rule.ID); err != nil {
		return fmt.Errorf("reviewRuleService.Delete: %w", err)
	}
	return nil
}

// List returns all review rules.
func (s *ReviewRuleService) List(ctx context.Context) ([]domain.ReviewRule, error) {
	rules, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("reviewRuleService.List: %w", err)
	}
	return rules, nil
}

// Toggle enables or disables a review rule by its public ID.
func (s *ReviewRuleService) Toggle(ctx context.Context, publicID string, enabled bool) (*domain.ReviewRule, error) {
	rule, err := s.repo.FindByPublicID(ctx, publicID)
	if err != nil {
		return nil, fmt.Errorf("reviewRuleService.Toggle: %w", err)
	}
	if rule == nil {
		return nil, ErrNotFound
	}

	if enabled {
		rule.IsEnabled = 1
	} else {
		rule.IsEnabled = 0
	}

	if err := s.repo.Update(ctx, rule); err != nil {
		return nil, fmt.Errorf("reviewRuleService.Toggle: %w", err)
	}
	return rule, nil
}

// EvaluateContent checks content against all enabled rules and returns violations.
func (s *ReviewRuleService) EvaluateContent(ctx context.Context, content string) ([]RuleViolation, error) {
	rules, err := s.repo.ListEnabled(ctx)
	if err != nil {
		return nil, fmt.Errorf("reviewRuleService.EvaluateContent: %w", err)
	}

	var violations []RuleViolation
	contentLower := strings.ToLower(content)

	for _, rule := range rules {
		switch rule.RuleType {
		case "keyword_blacklist":
			violations = append(violations, evaluateKeywordBlacklist(rule, contentLower)...)
		case "pattern_match":
			violations = append(violations, evaluatePatternMatch(rule, content)...)
		case "content_policy":
			violations = append(violations, evaluateContentPolicy(rule, contentLower)...)
		}
	}

	if violations == nil {
		violations = []RuleViolation{}
	}

	return violations, nil
}

// evaluateKeywordBlacklist checks if content contains any blacklisted keywords.
func evaluateKeywordBlacklist(rule domain.ReviewRule, contentLower string) []RuleViolation {
	keywords := strings.Split(rule.RuleContent, ",")
	var matched []string
	for _, kw := range keywords {
		kw = strings.TrimSpace(kw)
		if kw == "" {
			continue
		}
		if strings.Contains(contentLower, strings.ToLower(kw)) {
			matched = append(matched, kw)
		}
	}
	if len(matched) > 0 {
		return []RuleViolation{
			{
				RulePublicID: rule.PublicID,
				RuleType:     rule.RuleType,
				RiskLevel:    rule.RiskLevel,
				Detail:       fmt.Sprintf("content contains blacklisted keywords: %s", strings.Join(matched, ", ")),
			},
		}
	}
	return nil
}

// evaluatePatternMatch checks if content matches a regex pattern.
func evaluatePatternMatch(rule domain.ReviewRule, content string) []RuleViolation {
	re, err := regexp.Compile(rule.RuleContent)
	if err != nil {
		return []RuleViolation{
			{
				RulePublicID: rule.PublicID,
				RuleType:     rule.RuleType,
				RiskLevel:    rule.RiskLevel,
				Detail:       fmt.Sprintf("invalid regex pattern: %s", err.Error()),
			},
		}
	}
	matches := re.FindAllString(content, 5)
	if len(matches) > 0 {
		return []RuleViolation{
			{
				RulePublicID: rule.PublicID,
				RuleType:     rule.RuleType,
				RiskLevel:    rule.RiskLevel,
				Detail:       fmt.Sprintf("content matches pattern: %s", strings.Join(matches, ", ")),
			},
		}
	}
	return nil
}

// evaluateContentPolicy performs a heuristic check based on policy description keywords.
func evaluateContentPolicy(rule domain.ReviewRule, contentLower string) []RuleViolation {
	// Heuristic: extract key terms from the policy description and check for their presence.
	policyTerms := strings.Fields(strings.ToLower(rule.RuleContent))
	matchCount := 0
	for _, term := range policyTerms {
		term = strings.TrimSpace(term)
		if len(term) < 3 {
			continue
		}
		if strings.Contains(contentLower, term) {
			matchCount++
		}
	}
	// If more than half the policy terms appear in the content, flag it.
	threshold := len(policyTerms) / 2
	if threshold < 1 {
		threshold = 1
	}
	if matchCount >= threshold {
		return []RuleViolation{
			{
				RulePublicID: rule.PublicID,
				RuleType:     rule.RuleType,
				RiskLevel:    rule.RiskLevel,
				Detail:       fmt.Sprintf("content may violate policy: %s", rule.RuleContent),
			},
		}
	}
	return nil
}
