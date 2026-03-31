package service

import (
	"math"
	"testing"
	"time"
)

func TestComputeHistoricalScore(t *testing.T) {
	tests := []struct {
		name     string
		feedback PerformanceFeedback
		wantMin  float64
		wantMax  float64
	}{
		{
			name:     "zero metrics",
			feedback: PerformanceFeedback{ReadCount: 0, ShareCount: 0, FansGained: 0},
			wantMin:  0,
			wantMax:  0.01,
		},
		{
			name:     "moderate performance",
			feedback: PerformanceFeedback{ReadCount: 5000, ShareCount: 500, FansGained: 200},
			wantMin:  20,
			wantMax:  80,
		},
		{
			name:     "high performance",
			feedback: PerformanceFeedback{ReadCount: 50000, ShareCount: 5000, FansGained: 2000},
			wantMin:  80,
			wantMax:  100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeHistoricalScore(tt.feedback)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("computeHistoricalScore(%+v) = %v, want [%v, %v]",
					tt.feedback, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestNormalise(t *testing.T) {
	// normalise(0, x) should be 0
	if got := normalise(0, 10000); got != 0 {
		t.Errorf("normalise(0, 10000) = %v, want 0", got)
	}

	// normalise should be monotonically increasing
	prev := normalise(100, 10000)
	for _, v := range []float64{500, 1000, 5000, 10000, 50000} {
		cur := normalise(v, 10000)
		if cur <= prev {
			t.Errorf("normalise not monotonic: normalise(%v) = %v <= %v", v, cur, prev)
		}
		prev = cur
	}

	// normalise should approach 100 for large values
	big := normalise(1000000, 10000)
	if big > 100 {
		t.Errorf("normalise(1000000, 10000) = %v, should be <= 100", big)
	}
	if big < 99 {
		t.Errorf("normalise(1000000, 10000) = %v, should be close to 100", big)
	}
}

func TestComputeRecencyFactor(t *testing.T) {
	// nil lastUsed → minimum factor
	got := computeRecencyFactor(nil)
	if got != 0.1 {
		t.Errorf("computeRecencyFactor(nil) = %v, want 0.1", got)
	}

	// just now → factor near 1.0
	now := time.Now()
	got = computeRecencyFactor(&now)
	if math.Abs(got-1.0) > 0.01 {
		t.Errorf("computeRecencyFactor(now) = %v, want ~1.0", got)
	}

	// 30 days ago → factor near 0.5
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	got = computeRecencyFactor(&thirtyDaysAgo)
	if math.Abs(got-0.5) > 0.05 {
		t.Errorf("computeRecencyFactor(30d ago) = %v, want ~0.5", got)
	}

	// 180 days ago → factor near minimum
	halfYearAgo := time.Now().AddDate(0, -6, 0)
	got = computeRecencyFactor(&halfYearAgo)
	if got > 0.15 {
		t.Errorf("computeRecencyFactor(180d ago) = %v, want <= 0.15", got)
	}
}
