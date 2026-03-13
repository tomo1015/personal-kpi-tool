package aggregator

import (
	"testing"
)

// -------------------------------------------------------------------
// Sum
// -------------------------------------------------------------------

func TestSum(t *testing.T) {
	tests := []struct {
		name string
		in   []float64
		want float64
	}{
		{
			name: "正常系: 複数要素",
			in:   []float64{1, 2, 3, 4, 5},
			want: 15,
		},
		{
			name: "正常系: 小数を含む",
			in:   []float64{1.1, 2.2, 3.3},
			want: 6.6,
		},
		{
			name: "正常系: 負の値を含む",
			in:   []float64{-1, 2, -3},
			want: -2,
		},
		{
			name: "境界値: 空スライス",
			in:   []float64{},
			want: 0,
		},
		{
			name: "境界値: 要素1件",
			in:   []float64{42},
			want: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Sum(tt.in)
			// 小数の丸め誤差を考慮して 1e-9 の許容差で比較する
			if !almostEqual(got, tt.want, 1e-9) {
				t.Errorf("Sum(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

// -------------------------------------------------------------------
// Avg
// -------------------------------------------------------------------

func TestAvg(t *testing.T) {
	tests := []struct {
		name string
		in   []float64
		want float64
	}{
		{
			name: "正常系: 整数値",
			in:   []float64{1, 2, 3, 4, 5},
			want: 3.0,
		},
		{
			name: "正常系: 小数点2桁に丸められる",
			in:   []float64{1, 2},
			want: 1.5,
		},
		{
			name: "正常系: 小数点3桁以下が切り捨てられる",
			// 1/3 = 0.3333... → 0.33 に丸め
			in:   []float64{1, 0, 0},
			want: 0.33,
		},
		{
			// 1.005 は浮動小数点で 1.00499999... として格納されるため
			// math.Round(100.4999...) = 100 となり 1.00 になる。
			// これは実装バグではなく IEEE 754 の仕様どおりの動作。
			name: "正常系: x.xx5 は浮動小数点誤差で切り捨てになる",
			in:   []float64{1.005, 1.005},
			want: 1.0,
		},
		{
			// 2/3 = 0.6666... → 0.67 に切り上がることを確認する。
			// 整数の割り算で生じる無限小数は安定して丸めが効く。
			name: "正常系: 無限小数は小数点2桁に切り上がる",
			in:   []float64{2, 0, 0},
			want: 0.67,
		},
		{
			name: "境界値: 空スライス",
			in:   []float64{},
			want: 0,
		},
		{
			name: "境界値: 要素1件",
			in:   []float64{3.14159},
			// 3.14159 → 3.14 に丸め
			want: 3.14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Avg(tt.in)
			if !almostEqual(got, tt.want, 1e-9) {
				t.Errorf("Avg(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

// -------------------------------------------------------------------
// Count
// -------------------------------------------------------------------

func TestCount(t *testing.T) {
	t.Run("[]float64: 複数要素", func(t *testing.T) {
		got := Count([]float64{1, 2, 3})
		if got != 3 {
			t.Errorf("Count = %v, want 3", got)
		}
	})

	t.Run("[]float64: 空スライス", func(t *testing.T) {
		got := Count([]float64{})
		if got != 0 {
			t.Errorf("Count = %v, want 0", got)
		}
	})

	t.Run("[]string: 文字列スライス", func(t *testing.T) {
		got := Count([]string{"a", "b", "c", "d"})
		if got != 4 {
			t.Errorf("Count = %v, want 4", got)
		}
	})

	t.Run("[]string: 空スライス", func(t *testing.T) {
		got := Count([]string{})
		if got != 0 {
			t.Errorf("Count = %v, want 0", got)
		}
	})

	// 任意の構造体スライスでも動作することを確認する
	t.Run("任意の構造体スライス", func(t *testing.T) {
		type dummy struct{ v int }
		got := Count([]dummy{{1}, {2}})
		if got != 2 {
			t.Errorf("Count = %v, want 2", got)
		}
	})
}

// -------------------------------------------------------------------
// Rate
// -------------------------------------------------------------------

func TestRate(t *testing.T) {
	tests := []struct {
		name       string
		part, total float64
		want       float64
	}{
		{
			name:  "正常系: 50%",
			part:  1, total: 2,
			want:  50.0,
		},
		{
			name:  "正常系: 100%",
			part:  5, total: 5,
			want:  100.0,
		},
		{
			name:  "正常系: 0%",
			part:  0, total: 10,
			want:  0.0,
		},
		{
			name:  "正常系: 小数点2桁に丸められる",
			// 1/3*100 = 33.333... → 33.33
			part:  1, total: 3,
			want:  33.33,
		},
		{
			name:  "正常系: 小数点3桁以下が切り上がる",
			// 2/3*100 = 66.666... → 66.67
			part:  2, total: 3,
			want:  66.67,
		},
		{
			name:  "境界値: total=0（ゼロ除算ガード）",
			part:  5, total: 0,
			want:  0.0,
		},
		{
			name:  "境界値: part と total が同じ",
			part:  7, total: 7,
			want:  100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Rate(tt.part, tt.total)
			if !almostEqual(got, tt.want, 1e-9) {
				t.Errorf("Rate(%v, %v) = %v, want %v", tt.part, tt.total, got, tt.want)
			}
		})
	}
}

// -------------------------------------------------------------------
// Distribution
// -------------------------------------------------------------------

func TestDistribution(t *testing.T) {
	t.Run("int: 重複あり", func(t *testing.T) {
		got := Distribution([]int{1, 2, 2, 3, 3, 3})
		want := map[int]int{1: 1, 2: 2, 3: 3}
		assertMapEqual(t, got, want)
	})

	t.Run("string: 重複あり", func(t *testing.T) {
		got := Distribution([]string{"a", "b", "a", "c", "b", "b"})
		want := map[string]int{"a": 2, "b": 3, "c": 1}
		assertMapEqual(t, got, want)
	})

	t.Run("全要素が同じ値", func(t *testing.T) {
		got := Distribution([]int{5, 5, 5})
		want := map[int]int{5: 3}
		assertMapEqual(t, got, want)
	})

	t.Run("全要素が異なる値", func(t *testing.T) {
		got := Distribution([]int{1, 2, 3})
		want := map[int]int{1: 1, 2: 1, 3: 1}
		assertMapEqual(t, got, want)
	})

	t.Run("境界値: 空スライス → 空マップ（nil ではない）", func(t *testing.T) {
		got := Distribution([]int{})
		// nil チェック: 空マップが返ることを確認する
		if got == nil {
			t.Error("Distribution(空スライス) = nil, want 空マップ")
		}
		if len(got) != 0 {
			t.Errorf("Distribution(空スライス) の len = %v, want 0", len(got))
		}
	})

	t.Run("境界値: 要素1件", func(t *testing.T) {
		got := Distribution([]string{"only"})
		want := map[string]int{"only": 1}
		assertMapEqual(t, got, want)
	})
}

// -------------------------------------------------------------------
// テストヘルパー
// -------------------------------------------------------------------

// almostEqual は2つの float64 が許容差 epsilon 以内で等しいかを返す。
// 浮動小数点の丸め誤差を吸収するために使用する。
func almostEqual(a, b, epsilon float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff <= epsilon
}

// assertMapEqual は2つのマップが等しいかを検証する。
// comparable な K 型と int 型の値に限定したヘルパー。
func assertMapEqual[K comparable](t *testing.T, got, want map[K]int) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("マップの長さ: got %v, want %v", len(got), len(want))
		return
	}
	for k, wv := range want {
		gv, ok := got[k]
		if !ok {
			t.Errorf("キー %v が存在しない", k)
			continue
		}
		if gv != wv {
			t.Errorf("キー %v: got %v, want %v", k, gv, wv)
		}
	}
}