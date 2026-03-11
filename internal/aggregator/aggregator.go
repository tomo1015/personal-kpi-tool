// Package aggregatorはゲーム非依存のKPI計算を提供する。
// レコード（map[string]string）とAggFunc列挙型のみを認識する。
// ゲーム固有の意味付けは全て継承先に存在する。
package aggregator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/yourname/game-kpi-engine/internal/reader"
	"github.com/yourname/game-kpi-engine/pkg/kpidef"
)

// ─────────────────────────────────────────────────
// 結果型
// ─────────────────────────────────────────────────

// Result は 1 つの Metric に対する計算結果値を保持します。
type Result struct {
	Metric kpidef.Metric
	// Scalar は Sum / Avg / Count / Rate に設定されます。
	Scalar float64
    // Distribution は AggDistribution: map[fieldValue]count に設定されます。
	Distribution map[string]int
}

// レポートは Compute() の出力です。
type Report struct {
	Title    string
	NRecords int
	Results  []Result
}

─────────────────────────────────────────────────
Compute
─ ────────────────────────────────────────────────

// Compute は、定義された各メトリックに対してレコードに対して実行されます。
// レコードは、継承先によって事前に処理済みである必要があります
// （つまり、派生フィールドが各レコードに書き戻されている状態）。
func Compute(def kpidef.KPIDefinition, records []reader.Record) (*Report, error) {
	rpt := &Report{
		Title:    def.DashboardTitle(),
		NRecords: len(records),
	}
	for _, m := range def.Metrics() {
		res, err := computeOne(m, records)
		if err != nil {
			return nil, fmt.Errorf("metric %q: %w", m.Name, err)
		}
		rpt.Results = append(rpt.Results, res)
	}
	return rpt, nil
}

func computeOne(m kpidef.Metric, records []reader.Record) (Result, error) {
	res := Result{Metric: m}
	switch m.Func {
	case kpidef.AggSum:
		res.Scalar = aggSum(records, m.Field)
	case kpidef.AggAvg:
		res.Scalar = aggAvg(records, m.Field)
	case kpidef.AggCount:
		res.Scalar = float64(aggCount(records, m.Field))
	case kpidef.AggRate:
		res.Scalar = aggRate(records, m.Field)
	case kpidef.AggDistribution:
		res.Distribution = aggDistribution(records, m.Field)
	default:
		return res, fmt.Errorf("unknown AggFunc %q", m.Func)
	}
	return res, nil
}

// ─────────────────────────────────────────────────
// 集計関数
// ─────────────────────────────────────────────────

func aggSum(records []reader.Record, field string) float64 {
	s := 0.0
	for _, r := range records {
		s += reader.ToFloat(r[field])
	}
	return s
}

func aggAvg(records []reader.Record, field string) float64 {
	if len(records) == 0 {
		return 0
	}
	return aggSum(records, field) / float64(len(records))
}

func aggCount(records []reader.Record, field string) int {
	n := 0
	for _, r := range records {
		if r[field] != "" {
			n++
		}
	}
	return n
}

// aggRate は、フィールドが真値である行数を、総行数で割った値に 100 を乗じた値です。
func aggRate(records []reader.Record, field string) float64 {
	if len(records) == 0 {
		return 0
	}
	n := 0
	for _, r := range records {
		if reader.IsTruthy(r[field]) {
			n++
		}
	}
	return float64(n) / float64(len(records)) * 100
}

func aggDistribution(records []reader.Record, field string) map[string]int {
	m := map[string]int{}
	for _, r := range records {
		m[r[field]]++
	}
	return m
}

// ─────────────────────────────────────────────────
// 汎用統計ヘルパー (継承先リポジトリ向けにエクスポート)
// ─────────────────────────────────────────────────

// Mean は v の算術平均を返します。空のスライスの場合は 0 を返します。
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

// Median は v のメディアンを返します。空のスライスの場合は 0 を返します。
func Median(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	s := make([]float64, len(v))
	copy(s, v)
	sort.Float64s(s)
	n := len(s)
	if n%2 == 0 {
		return (s[n/2-1] + s[n/2]) / 2
	}
	return s[n/2]
}

// Max64 は v の最大値を返します。空のスライスには 0 を返します。
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

// GroupBy は、keyFn が返す文字列によってレコードをグループ化します。
func GroupBy(records []reader.Record, keyFn func(reader.Record) string) map[string][]reader.Record {
	m := map[string][]reader.Record{}
	for _, r := range records {
		k := keyFn(r)
		m[k] = append(m[k], r)
	}
	return m
}

// BuildGroupTop5 はグループキー（例：要素、軍種）ごとに上位5ユニットのランク付けリストを生成します。
// excluded はスキップする m_unit_id 値の集合です（例：95%以上の所有権ユニット）。
// keyFn はレコードからグループ名を抽出します。
// このヘルパーは、継承先リポジトリから渡されるレコードフィールドのみを操作するため、ゲームに依存しません。
func BuildGroupTop5(
	records []reader.Record,
	nUsers int,
	excluded map[string]bool,
	groupKeyFn func(reader.Record) string,
	unitIDField string,   // e.g. "m_unit_id"
	unitNameField string, // e.g. "unit_name"
	userIDField string,   // e.g. "user_id"
) []GroupTop5 {
	type unitAcc struct {
		name   string
		owners map[string]bool
	}
	// groups[groupName][unitID] → acc
	groups := map[string]map[string]*unitAcc{}
	for _, r := range records {
		if excluded[r[unitIDField]] {
			continue
		}
		gn := groupKeyFn(r)
		if gn == "" {
			continue
		}
		if groups[gn] == nil {
			groups[gn] = map[string]*unitAcc{}
		}
		a, ok := groups[gn][r[unitIDField]]
		if !ok {
			a = &unitAcc{name: r[unitNameField], owners: map[string]bool{}}
			groups[gn][r[unitIDField]] = a
		}
		a.owners[r[userIDField]] = true
	}

	var result []GroupTop5
	for gn, units := range groups {
		var entries []NamedCount
		for _, a := range units {
			entries = append(entries, NamedCount{
				Name:  a.name,
				Count: len(a.owners),
				Rate:  float64(len(a.owners)) / float64(nUsers) * 100,
			})
		}
		sort.Slice(entries, func(i, j int) bool { return entries[i].Count > entries[j].Count })
		if len(entries) > 5 {
			entries = entries[:5]
		}
		result = append(result, GroupTop5{GroupName: gn, Units: entries})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].GroupName < result[j].GroupName
	})
	return result
}

// ─────────────────────────────────────────────────
// Shared display types (game-agnostic, used by renderer)
// ─────────────────────────────────────────────────

type NamedCount struct {
	Name  string
	Count int
	Rate  float64
}

type NamedFloat struct {
	Name  string
	Value float64
}

type GroupTop5 struct {
	GroupName string
	Units     []NamedCount
}

// suppress unused import
var _ = strings.TrimSpace