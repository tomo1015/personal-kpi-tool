// analyzer は game-kpi-engine の動作確認用 CLI。
// サンプルの KPIDefinition 実装（sampleDef）を使って集計→HTML 出力の
// 一連のパイプラインが正常に動くことを確認するためのエントリーポイント。
// 本番利用には company repo の cmd/run を使用すること。
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/tomo1015/personal-kpi-tool/example"
	"github.com/tomo1015/personal-kpi-tool/internal/aggregator"
	"github.com/tomo1015/personal-kpi-tool/internal/chartjs"
	"github.com/tomo1015/personal-kpi-tool/internal/renderer"
	"github.com/tomo1015/personal-kpi-tool/pkg/csvreader"
	"github.com/tomo1015/personal-kpi-tool/pkg/kpidef"
)

func main() {
	input := flag.String("input", "", "入力CSVファイルパス（省略時はダミーデータを使用)")
	output := flag.String("output", "sample_report.html", "出力HTMLファイルパス")
	tmplPath := flag.String("template", "", "カスタムテンプレートパス")
	flag.Parse()

	fmt.Println("=== game-kpi-engine サンプル動作確認 ===")

	// --inputが指定された場合CSVを読み込む。省略時はダミーデータを使用する。
	var rows []map[string]string
	if *input != "" {
		var err error
		rows, err = csvreader.ReadAllCSV(*input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "CSV 読み込みエラー: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("CSV 読み込み完了: %s (%d 行)\n", *input, len(rows))
	} else {
		fmt.Println("--input 未指定のためダミーデータを使用します")
		rows = []map[string]string{
			{"user_id": "u001", "value": "10"},
			{"user_id": "u002", "value": "20"},
			{"user_id": "u003", "value": "30"},
		}
	}
	// サンプル KPIDefinition を使って集計
	def := &example.SampleDef{}
	rd, err := aggregator.Compute(def, rows, "なし")
	if err != nil {
		fmt.Fprintf(os.Stderr, "集計エラー: %v\n", err)
		os.Exit(1)
	}

	// Chart.js を取得
	chartJS, err := chartjs.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Chart.js 読み込みエラー: %v\n", err)
		os.Exit(1)
	}

	// HTML 出力
	if err := renderer.Render(rd, renderer.Options{
		OutputPath:   *output,
		TemplatePath: *tmplPath,
		ChartJS:      chartJS,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "レンダリングエラー: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("レポートを出力しました: %s\n", *output)
}

// -------------------------------------------------------------------
// サンプル KPIDefinition 実装
// -------------------------------------------------------------------

// sampleDef は engine の動作確認用の最小 KPIDefinition 実装。
// 渡された []map[string]string を集計し、件数と平均値を返す。
type sampleDef struct{}

func (s *sampleDef) DashboardTitle() string { return "サンプル KPI レポート" }
func (s *sampleDef) TemplatePath() string   { return "" }

func (s *sampleDef) Compute(rows any, filterInfo string) (*kpidef.RenderData, error) {
	// any を []map[string]string に型アサーション
	recs, ok := rows.([]map[string]string)
	if !ok {
		return nil, fmt.Errorf("rows の型が想定外です")
	}

	// 簡易集計：value カラムの合計と平均
	sum := 0.0
	for _, rec := range recs {
		v := 0.0
		if _, err := fmt.Sscanf(rec["value"], "%f", &v); err != nil {
			continue //パース失敗行はスキップ
		}
		sum += v
	}
	avg := 0.0
	if len(recs) > 0 {
		avg = sum / float64(len(recs))
	}

	// ユーザー数（重複なし）
	users := map[string]bool{}
	for _, rec := range recs {
		if uid := rec["user_id"]; uid != "" {
			users[uid] = true
		}
	}

	payload := map[string]any{
		"count": len(recs),
		"sum":   sum,
		"avg":   avg,
	}

	return &kpidef.RenderData{
		Title:       s.DashboardTitle(),
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
		FilterInfo:  filterInfo,
		NUsers:      len(users),
		Payload:     payload,
	}, nil
}
