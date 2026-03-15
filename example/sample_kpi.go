// Package example は personal-kpi-tool の KPIDefinition 実装サンプルです。
// company repo が独自の KPIDefinition を実装する際の参考として使用してください。
package example

import (
	"fmt"
	"time"

	"github.com/tomo1015/personal-kpi-tool/pkg/kpidef"
)

// SampleDef は engine の動作確認用の最小 KPIDefinition 実装。
// 渡された []map[string]string を集計し、件数・合計・平均値を返す。
type SampleDef struct{}

func (s *SampleDef) DashboardTitle() string { return "サンプル KPI レポート" }
func (s *SampleDef) TemplatePath() string   { return "" }

// Compute は rows（[]map[string]string）を受け取り、value カラムを集計する。
func (s *SampleDef) Compute(rows any, filterInfo string) (*kpidef.RenderData, error) {
	recs, ok := rows.([]map[string]string)
	if !ok {
		return nil, fmt.Errorf("rows の型が想定外です: %T", rows)
	}

	// value カラムの合計・平均
	sum := 0.0
	for _, rec := range recs {
		v := 0.0
		if _, err := fmt.Sscanf(rec["value"], "%f", &v); err != nil {
			continue // パース失敗行はスキップ
		}
		sum += v
	}
	avg := 0.0
	if len(recs) > 0 {
		avg = sum / float64(len(recs))
	}

	// ユーザー数（重複なし）
	users := make(map[string]bool)
	for _, rec := range recs {
		if uid := rec["user_id"]; uid != "" {
			users[uid] = true
		}
	}

	return &kpidef.RenderData{
		Title:       s.DashboardTitle(),
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
		FilterInfo:  filterInfo,
		NUsers:      len(users),
		Payload: map[string]any{
			"count": len(recs),
			"sum":   sum,
			"avg":   avg,
		},
	}, nil
}
