package main

import (
	"testing"
	"time"

	"github.com/xieburoucoco/spider-hub/platforms/coinglass"
)

// 🧪 TestIncrementalFlowLimit 增量笔数依停机时间自动放大，并夹在上下限之间
func TestIncrementalFlowLimit(t *testing.T) {
	now := time.Date(2026, 10, 8, 5, 11, 0, 0, time.UTC)
	cases := []struct {
		name   string
		latest time.Time
		dur    time.Duration
		want   int
	}{
		{"刚跑过一小时", now.Add(-time.Hour), time.Hour, flowIncrementalLimit},
		{"停机约47小时（本次实际情况）", time.Date(2026, 10, 6, 6, 0, 0, 0, time.UTC), time.Hour, 51},
		{"停机10天", now.Add(-240 * time.Hour), time.Hour, 244},
		{"停机一年，夹到服务端上限", now.Add(-365 * 24 * time.Hour), time.Hour, coinglass.FlowMaxLimit},
		{"日线停2天", now.Add(-48 * time.Hour), 24 * time.Hour, flowIncrementalLimit},
	}
	for _, c := range cases {
		if got := incrementalFlowLimit(c.latest, c.dur, now); got != c.want {
			t.Errorf("❌ %s: 应为 %d，实际 %d", c.name, c.want, got)
		}
	}
}
