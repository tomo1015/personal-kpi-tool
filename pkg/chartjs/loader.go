// Package chartjs は internal/chartjs の公開ラッパーです。
package chartjs

import "github.com/tomo1015/personal-kpi-tool/internal/chartjs"

// Load は Chart.js のソースコードを返します。
func Load() (string, error) {
	return chartjs.Load()
}
