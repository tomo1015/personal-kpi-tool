// Package aggregator は internal/aggregator の公開ラッパーです。
package aggregator

import (
	"github.com/tomo1015/personal-kpi-tool/internal/aggregator"
	"github.com/tomo1015/personal-kpi-tool/pkg/kpidef"
)

// Compute は KPIDefinition を受け取り集計を実行します。
func Compute(def kpidef.KPIDefinition, rows any, filterInfo string) (*kpidef.RenderData, error) {
	return aggregator.Compute(def, rows, filterInfo)
}
