package aggregator

import (
	"math"
	"sort"
)

// Mean は float64 スライスの算術平均を返す。空スライスの場合は 0 を返す。
func Mean(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	s := 0.0
	for _, x := range v {
		s += x
	}
	return s / float64(len(v))
}

// Median は float64 スライスの中央値を返す。
// 要素数が偶数の場合は中央 2 値の平均を返す。空スライスの場合は 0 を返す。
func Median(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	// 元スライスを破壊しないようにコピーしてソート
	s := make([]float64, len(v))
	copy(s, v)
	sort.Float64s(s)
	n := len(s)
	if n%2 == 0 {
		return (s[n/2-1] + s[n/2]) / 2
	}
	return s[n/2]
}

// Max64 は float64 スライスの最大値を返す。空スライスの場合は 0 を返す。
func Max64(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	m := v[0]
	for _, x := range v[1:] {
		if x > m {
			m = x
		}
	}
	return m
}

// -------------------------------------------------------------------
// KPI 集計関数
// -------------------------------------------------------------------

// Sum は float64 スライスの合計を返す。空スライスの場合は 0 を返す。
func Sum(v []float64) float64 {
	s := 0.0
	for _, x := range v {
		s += x
	}
	return s
}

// Avg は float64 スライスの平均を小数点2桁で返す。
// 空スライスの場合は 0 を返す。
// 丸めなしの平均が必要な内部計算には Mean を使うこと。
func Avg(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	raw := Sum(v) / float64(len(v))
	return math.Round(raw*100) / 100
}

// Count は任意のスライスの要素数を返す。
func Count[T any](v []T) int {
	return len(v)
}

// Rate は part/total*100 を小数点2桁で返す（単位: %）。
// total が 0 の場合は 0 を返す。
func Rate(part, total float64) float64 {
	if total == 0 {
		return 0
	}
	raw := part / total * 100
	return math.Round(raw*100) / 100
}

// Distribution は comparable なスライスを受け取り、
// 値ごとの出現回数を map[T]int で返す。
// 空スライスの場合は空の（nil ではない）マップを返す。
func Distribution[T comparable](v []T) map[T]int {
	// 最悪ケース（全要素が異なる）を想定して初期容量を確保する
	m := make(map[T]int, len(v))
	for _, x := range v {
		m[x]++
	}
	return m
}

// GroupBy は任意のキー関数でレコードスライスをグループ化する。
// K はマップのキー型（comparable であれば何でもよい）、
// T はレコードの要素型。
func GroupBy[K comparable, T any](items []T, keyFn func(T) K) map[K][]T {
	m := make(map[K][]T)
	for _, item := range items {
		k := keyFn(item)
		m[k] = append(m[k], item)
	}
	return m
}