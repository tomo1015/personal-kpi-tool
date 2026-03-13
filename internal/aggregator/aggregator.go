// Package aggregator は KPIDefinition を受け取り集計処理を実行する。
// ゲーム固有ロジックはすべて KPIDefinition.Compute に委譲するため、
// このパッケージ自体はゲーム仕様を一切知らない。
package aggregator

import (
	"fmt"

	"github.com/tomo1015/personal-kpi-tool/pkg/kpidef"
)

// Compute は def.Compute(rows, filterInfo) を呼び出し、
// RenderData を返す薄いラッパー。
// 将来的にミドルウェア（タイミング計測・ロギング等）を
// ここに挟めるよう関数として切り出している。
func Compute(def kpidef.KPIDefinition, rows any, filterInfo string) (*kpidef.RenderData, error) {
	if def == nil {
		return nil, fmt.Errorf("KPIDefinition が nil です")
	}
	rd, err := def.Compute(rows, filterInfo)
	if err != nil {
		return nil, fmt.Errorf("集計エラー: %w", err)
	}
	if rd == nil {
		return nil, fmt.Errorf("Compute が nil の RenderData を返しました")
	}
	// タイトルが未設定の場合は KPIDefinition のタイトルで補完する
	if rd.Title == "" {
		rd.Title = def.DashboardTitle()
	}
	return rd, nil
}