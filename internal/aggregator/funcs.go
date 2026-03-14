// このファイルは engine 内部から呼ばれる集計ユーティリティ。
// 実装の実体は pkg/mathutil に置き、こちらはラッパーとして委譲する。
// 外部モジュール（company repo 等）は pkg/mathutil を直接 import すること。
package aggregator

import "github.com/tomo1015/personal-kpi-tool/pkg/mathutil"

// Mean は float64 スライスの算術平均を返す。空スライスの場合は 0 を返す。
func Mean(v []float64) float64 { return mathutil.Mean(v) }

// Median は float64 スライスの中央値を返す。空スライスの場合は 0 を返す。
func Median(v []float64) float64 { return mathutil.Median(v) }

// Max64 は float64 スライスの最大値を返す。空スライスの場合は 0 を返す。
func Max64(v []float64) float64 { return mathutil.Max64(v) }

// Sum は float64 スライスの合計を返す。空スライスの場合は 0 を返す。
func Sum(v []float64) float64 { return mathutil.Sum(v) }

// Avg は float64 スライスの平均を小数点2桁で返す。空スライスの場合は 0 を返す。
func Avg(v []float64) float64 { return mathutil.Avg(v) }

// Count は任意のスライスの要素数を返す。
func Count[T any](v []T) int { return mathutil.Count(v) }

// Rate は part/total*100 を小数点2桁で返す（単位: %）。total=0 のとき 0 を返す。
func Rate(part, total float64) float64 { return mathutil.Rate(part, total) }

// Distribution は値ごとの出現回数を map[T]int で返す。
func Distribution[T comparable](v []T) map[T]int { return mathutil.Distribution(v) }

// GroupBy は任意のキー関数でレコードスライスをグループ化する。
func GroupBy[K comparable, T any](items []T, keyFn func(T) K) map[K][]T {
	return mathutil.GroupBy(items, keyFn)
}