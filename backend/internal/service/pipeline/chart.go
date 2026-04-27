// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"readbud/internal/domain/source"
	"readbud/internal/repository/postgres"
)

// ChartDetectorService analyzes source document data points and detects chart opportunities.
type ChartDetectorService struct {
	sourceRepo postgres.SourceDocumentRepository
}

// NewChartDetectorService creates a new ChartDetectorService.
func NewChartDetectorService(sourceRepo postgres.SourceDocumentRepository) *ChartDetectorService {
	return &ChartDetectorService{sourceRepo: sourceRepo}
}

// DetectCharts analyzes data points from source documents and returns chart candidates.
func (s *ChartDetectorService) DetectCharts(ctx context.Context, taskID int64) ([]ChartCandidate, error) {
	docs, err := s.sourceRepo.FindByTaskIDOrderByScore(ctx, taskID, 10)
	if err != nil {
		return nil, fmt.Errorf("chartDetectorService.DetectCharts: %w", err)
	}

	var candidates []ChartCandidate

	for _, doc := range docs {
		docCandidates := detectFromDocument(doc)
		candidates = append(candidates, docCandidates...)
	}

	return candidates, nil
}

// detectFromDocument extracts chart candidates from a single source document's DataPointsJSON.
func detectFromDocument(doc source.SourceDocument) []ChartCandidate {
	if len(doc.DataPointsJSON) == 0 {
		return nil
	}

	var dataPoints []DataPoint
	if err := json.Unmarshal(doc.DataPointsJSON, &dataPoints); err != nil {
		return nil
	}

	if len(dataPoints) < 3 {
		return nil
	}

	var candidates []ChartCandidate

	// Detect time series data (dimension contains time-related keywords)
	if candidate, ok := detectTimeSeries(dataPoints, doc.Title); ok {
		candidates = append(candidates, candidate)
	}

	// Detect ranking data (values sorted or rankable)
	if candidate, ok := detectRanking(dataPoints, doc.Title); ok {
		candidates = append(candidates, candidate)
	}

	// Detect category comparison data
	if candidate, ok := detectCategoryComparison(dataPoints, doc.Title); ok {
		candidates = append(candidates, candidate)
	}

	// Detect composition data (values that could represent parts of a whole)
	if candidate, ok := detectComposition(dataPoints, doc.Title); ok {
		candidates = append(candidates, candidate)
	}

	// If no specific pattern detected but enough data points, default to bar chart
	if len(candidates) == 0 {
		candidates = append(candidates, buildCandidate("bar", dataPoints, doc.Title))
	}

	return candidates
}

// isTimeDimension checks if a dimension string indicates time-based data.
func isTimeDimension(dim string) bool {
	timeKeywords := []string{"年", "月", "日", "季度", "week", "month", "year", "quarter", "date", "time", "period"}
	lower := strings.ToLower(dim)
	for _, kw := range timeKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// detectTimeSeries checks if data points represent a time series.
func detectTimeSeries(points []DataPoint, docTitle string) (ChartCandidate, bool) {
	timeCount := 0
	for _, p := range points {
		if isTimeDimension(p.Dimension) {
			timeCount++
		}
	}
	// If majority of points have time dimensions, treat as time series
	if timeCount < len(points)/2 {
		return ChartCandidate{}, false
	}

	candidate := buildCandidate("line", points, docTitle)
	candidate.Title = fmt.Sprintf("%s - 趋势变化", shortenTitle(docTitle))
	return candidate, true
}

// detectRanking checks if data points represent a ranking.
func detectRanking(points []DataPoint, docTitle string) (ChartCandidate, bool) {
	if len(points) < 3 {
		return ChartCandidate{}, false
	}

	// Check if values are in descending order (or nearly so)
	descCount := 0
	for i := 1; i < len(points); i++ {
		if points[i].Value <= points[i-1].Value {
			descCount++
		}
	}

	if descCount < (len(points)-1)*2/3 {
		return ChartCandidate{}, false
	}

	candidate := buildCandidate("horizontal_bar", points, docTitle)
	candidate.Title = fmt.Sprintf("%s - 排名对比", shortenTitle(docTitle))
	return candidate, true
}

// detectCategoryComparison checks if data points represent category comparisons.
func detectCategoryComparison(points []DataPoint, docTitle string) (ChartCandidate, bool) {
	// Category comparison: no time dimensions, distinct labels, 3-10 items
	if len(points) < 3 || len(points) > 10 {
		return ChartCandidate{}, false
	}

	for _, p := range points {
		if isTimeDimension(p.Dimension) {
			return ChartCandidate{}, false
		}
	}

	candidate := buildCandidate("bar", points, docTitle)
	candidate.Title = fmt.Sprintf("%s - 对比分析", shortenTitle(docTitle))
	return candidate, true
}

// detectComposition checks if data points could represent parts of a whole.
func detectComposition(points []DataPoint, docTitle string) (ChartCandidate, bool) {
	if len(points) < 2 || len(points) > 8 {
		return ChartCandidate{}, false
	}

	// Check if all values are positive and could be percentages or proportions
	total := 0.0
	allPositive := true
	for _, p := range points {
		if p.Value < 0 {
			allPositive = false
			break
		}
		total += p.Value
	}

	if !allPositive {
		return ChartCandidate{}, false
	}

	// If total is close to 100 or unit indicates percentages, treat as composition
	unit := ""
	if len(points) > 0 {
		unit = points[0].Unit
	}

	isPercentage := strings.Contains(unit, "%") || strings.Contains(unit, "百分") ||
		(total >= 95 && total <= 105)

	if !isPercentage {
		return ChartCandidate{}, false
	}

	candidate := buildCandidate("donut", points, docTitle)
	candidate.Title = fmt.Sprintf("%s - 构成分析", shortenTitle(docTitle))
	return candidate, true
}

// buildCandidate creates a ChartCandidate from data points.
func buildCandidate(chartType string, points []DataPoint, docTitle string) ChartCandidate {
	labels := make([]string, 0, len(points))
	values := make([]float64, 0, len(points))
	unit := ""
	src := ""

	for _, p := range points {
		labels = append(labels, p.Label)
		values = append(values, p.Value)
		if unit == "" && p.Unit != "" {
			unit = p.Unit
		}
		if src == "" && p.Source != "" {
			src = p.Source
		}
	}

	if src == "" {
		src = docTitle
	}

	return ChartCandidate{
		ChartType: chartType,
		Title:     shortenTitle(docTitle),
		Labels:    labels,
		Values:    values,
		Unit:      unit,
		Source:    src,
	}
}

// shortenTitle truncates a title to a reasonable length for chart display.
func shortenTitle(title string) string {
	runes := []rune(title)
	if len(runes) > 20 {
		return string(runes[:20])
	}
	return title
}
