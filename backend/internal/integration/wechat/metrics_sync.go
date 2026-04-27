// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package wechat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"go.uber.org/zap"

	"readbud/internal/adapter"
)

// WeChat datacube endpoints.
const (
	endpointArticleTotal = "https://api.weixin.qq.com/datacube/getarticletotal"
	endpointUserSummary  = "https://api.weixin.qq.com/datacube/getusersummary"
	endpointUserCumulate = "https://api.weixin.qq.com/datacube/getusercumulate"

	// articleMetricsLookbackDays is how many days back we aggregate when fetching
	// per-article metrics. WeChat caps getarticletotal range at 7 days.
	articleMetricsLookbackDays = 7

	defaultMetricsTimeout = 15 * time.Second
)

// RealMetricsSyncProvider is the production implementation of adapter.MetricsSyncProvider.
// It calls WeChat's datacube APIs (analytics) using the provided access_token.
type RealMetricsSyncProvider struct {
	http   *http.Client
	logger *zap.Logger
}

// NewRealMetricsSyncProvider builds a production metrics sync provider.
func NewRealMetricsSyncProvider(httpClient *http.Client, logger *zap.Logger) *RealMetricsSyncProvider {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultMetricsTimeout}
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &RealMetricsSyncProvider{http: httpClient, logger: logger}
}

// ----- GetArticleMetrics -----

type articleTotalReq struct {
	BeginDate string `json:"begin_date"`
	EndDate   string `json:"end_date"`
}

type articleTotalRespItem struct {
	RefDate string                  `json:"ref_date"`
	MsgID   string                  `json:"msgid"`
	Title   string                  `json:"title"`
	Details []articleTotalDayDetail `json:"details"`
}

type articleTotalDayDetail struct {
	StatDate          string `json:"stat_date"`
	TargetUser        int    `json:"target_user"`
	IntPageReadUser   int    `json:"int_page_read_user"`
	IntPageReadCount  int    `json:"int_page_read_count"`
	OriPageReadUser   int    `json:"ori_page_read_user"`
	OriPageReadCount  int    `json:"ori_page_read_count"`
	ShareUser         int    `json:"share_user"`
	ShareCount        int    `json:"share_count"`
	AddToFavUser      int    `json:"add_to_fav_user"`
	AddToFavCount     int    `json:"add_to_fav_count"`
}

type articleTotalResp struct {
	List    []articleTotalRespItem `json:"list"`
	Errcode int                    `json:"errcode"`
	Errmsg  string                 `json:"errmsg"`
}

// GetArticleMetrics fetches per-article read/share metrics from /datacube/getarticletotal.
// The articleID is matched against the WeChat msgid field. Stats are summed across the
// last N days (capped by WeChat's 7-day limit). If the article isn't found in the window
// (likely too old or just-published), this returns zero counts rather than an error so
// the caller can record "no data yet" without failing the sync job.
func (p *RealMetricsSyncProvider) GetArticleMetrics(
	ctx context.Context, accessToken string, articleID string,
) (*adapter.ArticleMetrics, error) {
	if accessToken == "" {
		return nil, errors.New("GetArticleMetrics: access token is required")
	}
	if articleID == "" {
		return nil, errors.New("GetArticleMetrics: article_id is required")
	}

	// WeChat data is typically delayed a day; query yesterday-N..yesterday window.
	end := time.Now().AddDate(0, 0, -1)
	begin := end.AddDate(0, 0, -(articleMetricsLookbackDays - 1))

	body, err := json.Marshal(articleTotalReq{
		BeginDate: begin.Format("2006-01-02"),
		EndDate:   end.Format("2006-01-02"),
	})
	if err != nil {
		return nil, fmt.Errorf("GetArticleMetrics: marshal: %w", err)
	}

	endpoint := endpointArticleTotal + "?access_token=" + url.QueryEscape(accessToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("GetArticleMetrics: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	raw, err := p.doRaw(req)
	if err != nil {
		return nil, fmt.Errorf("GetArticleMetrics: %w", err)
	}
	var resp articleTotalResp
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("GetArticleMetrics: decode: %w", err)
	}
	if resp.Errcode != 0 {
		return nil, &APIError{Code: resp.Errcode, Message: resp.Errmsg, Op: "getarticletotal"}
	}

	out := &adapter.ArticleMetrics{
		ArticleID: articleID,
		RawJSON:   raw,
	}
	for _, item := range resp.List {
		if item.MsgID != articleID {
			continue
		}
		for _, d := range item.Details {
			out.ReadCount += d.IntPageReadCount + d.OriPageReadCount
			out.ReadUserCount += d.IntPageReadUser + d.OriPageReadUser
			out.ShareCount += d.ShareCount
			out.ShareUserCount += d.ShareUser
		}
	}

	p.logger.Info("fetched article metrics from WeChat",
		zap.String("article_id", articleID),
		zap.String("begin", begin.Format("2006-01-02")),
		zap.String("end", end.Format("2006-01-02")),
		zap.Int("read", out.ReadCount),
		zap.Int("share", out.ShareCount),
	)
	return out, nil
}

// ----- GetFansMetrics -----

type userSummaryReq struct {
	BeginDate string `json:"begin_date"`
	EndDate   string `json:"end_date"`
}

type userSummaryItem struct {
	RefDate    string `json:"ref_date"`
	UserSource int    `json:"user_source"`
	NewUser    int    `json:"new_user"`
	CancelUser int    `json:"cancel_user"`
}

type userSummaryResp struct {
	List    []userSummaryItem `json:"list"`
	Errcode int               `json:"errcode"`
	Errmsg  string            `json:"errmsg"`
}

type userCumulateItem struct {
	RefDate      string `json:"ref_date"`
	CumulateUser int    `json:"cumulate_user"`
}

type userCumulateResp struct {
	List    []userCumulateItem `json:"list"`
	Errcode int                `json:"errcode"`
	Errmsg  string             `json:"errmsg"`
}

// GetFansMetrics fetches daily fan growth via /datacube/getusersummary and pairs it with
// /datacube/getusercumulate for the cumulative total. date format: YYYY-MM-DD.
func (p *RealMetricsSyncProvider) GetFansMetrics(
	ctx context.Context, accessToken string, date string,
) (*adapter.FansMetrics, error) {
	if accessToken == "" {
		return nil, errors.New("GetFansMetrics: access token is required")
	}
	if date == "" {
		return nil, errors.New("GetFansMetrics: date is required")
	}
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, fmt.Errorf("GetFansMetrics: invalid date %q: %w", date, err)
	}

	summary, summaryRaw, err := p.fetchUserSummary(ctx, accessToken, date)
	if err != nil {
		return nil, err
	}
	cumulative, _, err := p.fetchUserCumulate(ctx, accessToken, date)
	if err != nil {
		// Cumulative is non-critical; log and continue with summary-only data.
		p.logger.Warn("fetch user cumulate failed", zap.String("date", date), zap.Error(err))
	}

	addFans := 0
	cancelFans := 0
	for _, item := range summary {
		if item.RefDate != date {
			continue
		}
		addFans += item.NewUser
		cancelFans += item.CancelUser
	}

	out := &adapter.FansMetrics{
		Date:            date,
		AddFansCount:    addFans,
		CancelFansCount: cancelFans,
		NetFansCount:    addFans - cancelFans,
		RawJSON:         summaryRaw,
	}
	if cumulative > 0 {
		// Embed cumulative count in raw JSON envelope so callers can access it without
		// extending the adapter contract.
		envelope := map[string]any{
			"summary":       json.RawMessage(summaryRaw),
			"cumulate_user": cumulative,
		}
		if b, err := json.Marshal(envelope); err == nil {
			out.RawJSON = b
		}
	}

	p.logger.Info("fetched fans metrics from WeChat",
		zap.String("date", date),
		zap.Int("new", addFans),
		zap.Int("cancel", cancelFans),
		zap.Int("cumulate", cumulative),
	)
	return out, nil
}

func (p *RealMetricsSyncProvider) fetchUserSummary(ctx context.Context, accessToken, date string) ([]userSummaryItem, []byte, error) {
	body, err := json.Marshal(userSummaryReq{BeginDate: date, EndDate: date})
	if err != nil {
		return nil, nil, fmt.Errorf("marshal: %w", err)
	}
	endpoint := endpointUserSummary + "?access_token=" + url.QueryEscape(accessToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	raw, err := p.doRaw(req)
	if err != nil {
		return nil, nil, fmt.Errorf("getusersummary: %w", err)
	}
	var resp userSummaryResp
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, nil, fmt.Errorf("decode usersummary: %w", err)
	}
	if resp.Errcode != 0 {
		return nil, nil, &APIError{Code: resp.Errcode, Message: resp.Errmsg, Op: "getusersummary"}
	}
	return resp.List, raw, nil
}

func (p *RealMetricsSyncProvider) fetchUserCumulate(ctx context.Context, accessToken, date string) (int, []byte, error) {
	body, err := json.Marshal(userSummaryReq{BeginDate: date, EndDate: date})
	if err != nil {
		return 0, nil, fmt.Errorf("marshal: %w", err)
	}
	endpoint := endpointUserCumulate + "?access_token=" + url.QueryEscape(accessToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return 0, nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	raw, err := p.doRaw(req)
	if err != nil {
		return 0, nil, fmt.Errorf("getusercumulate: %w", err)
	}
	var resp userCumulateResp
	if err := json.Unmarshal(raw, &resp); err != nil {
		return 0, nil, fmt.Errorf("decode usercumulate: %w", err)
	}
	if resp.Errcode != 0 {
		return 0, nil, &APIError{Code: resp.Errcode, Message: resp.Errmsg, Op: "getusercumulate"}
	}
	for _, item := range resp.List {
		if item.RefDate == date {
			return item.CumulateUser, raw, nil
		}
	}
	return 0, raw, nil
}

// doRaw executes the request and returns the raw body for callers that want both the
// decoded shape and the original bytes (for storage in *.RawJSON columns).
func (p *RealMetricsSyncProvider) doRaw(req *http.Request) ([]byte, error) {
	res, err := p.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http: %w", err)
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("http %d: %s", res.StatusCode, string(raw))
	}
	return raw, nil
}

// ----- Stub (kept for tests / fallback) -----

// StubMetricsSyncProvider is a placeholder implementation of adapter.MetricsSyncProvider.
// Returns synthetic-looking data; do NOT use in production.
type StubMetricsSyncProvider struct {
	logger *zap.Logger
}

// NewStubMetricsSyncProvider creates a new stub metrics sync provider.
func NewStubMetricsSyncProvider(logger *zap.Logger) *StubMetricsSyncProvider {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &StubMetricsSyncProvider{logger: logger}
}

func (p *StubMetricsSyncProvider) GetArticleMetrics(_ context.Context, accessToken string, articleID string) (*adapter.ArticleMetrics, error) {
	if accessToken == "" {
		return nil, errors.New("StubMetricsSyncProvider.GetArticleMetrics: access token is required")
	}
	if articleID == "" {
		return nil, errors.New("StubMetricsSyncProvider.GetArticleMetrics: article_id is required")
	}
	readCount := 100 + rand.Intn(9900)
	readUserCount := readCount - rand.Intn(readCount/5+1)
	shareCount := rand.Intn(readCount / 10)
	shareUserCount := shareCount - rand.Intn(shareCount/3+1)
	if shareUserCount < 0 {
		shareUserCount = 0
	}
	rawData := map[string]any{
		"article_id":       articleID,
		"read_num":         readCount,
		"read_unique_num":  readUserCount,
		"share_num":        shareCount,
		"share_unique_num": shareUserCount,
		"stub":             true,
		"ts":               strconv.FormatInt(time.Now().Unix(), 10),
	}
	rawJSON, _ := json.Marshal(rawData)
	p.logger.Info("stub: fetched article metrics", zap.String("article_id", articleID), zap.Int("read", readCount))
	return &adapter.ArticleMetrics{
		ArticleID:      articleID,
		ReadCount:      readCount,
		ReadUserCount:  readUserCount,
		ShareCount:     shareCount,
		ShareUserCount: shareUserCount,
		RawJSON:        rawJSON,
	}, nil
}

func (p *StubMetricsSyncProvider) GetFansMetrics(_ context.Context, accessToken string, date string) (*adapter.FansMetrics, error) {
	if accessToken == "" {
		return nil, errors.New("StubMetricsSyncProvider.GetFansMetrics: access token is required")
	}
	if date == "" {
		return nil, errors.New("StubMetricsSyncProvider.GetFansMetrics: date is required")
	}
	addFans := 5 + rand.Intn(50)
	cancelFans := rand.Intn(addFans/2 + 1)
	rawData := map[string]any{
		"ref_date":      date,
		"new_user":      addFans,
		"cancel_user":   cancelFans,
		"cumulate_user": 10000 + rand.Intn(5000),
		"stub":          true,
	}
	rawJSON, _ := json.Marshal(rawData)
	p.logger.Info("stub: fetched fans metrics", zap.String("date", date), zap.Int("new", addFans))
	return &adapter.FansMetrics{
		Date:            date,
		AddFansCount:    addFans,
		CancelFansCount: cancelFans,
		NetFansCount:    addFans - cancelFans,
		RawJSON:         rawJSON,
	}, nil
}

// Compile-time interface checks.
var (
	_ adapter.MetricsSyncProvider = (*RealMetricsSyncProvider)(nil)
	_ adapter.MetricsSyncProvider = (*StubMetricsSyncProvider)(nil)
)
