package aggregator

import "sort"

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