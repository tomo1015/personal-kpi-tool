// パッケージレンダラーはアグリゲーターの出力をHTMLダッシュボードに変換します。
// 継承先のリポジトリはKPIDefinition.TemplatePath()を介してテンプレートを上書きします。
package renderer

import (
	"fmt"
	htmpl "html/template"
	"os"
	"strings"
	ttmpl "text/template"
	"time"

	"github.com/yourname/game-kpi-engine/internal/aggregator"
	"github.com/yourname/game-kpi-engine/pkg/kpidef"
)

// ─────────────────────────────────────────────────
// RenderData
// ────────────────────────────────────── ───────────

// RenderData は HTML テンプレートにそのまま渡されます。
// 継承先テンプレートは .Results 経由でエンジンの結果にアクセスし、
// 継承先固有のレポートは .Extra 経由でアクセスします（テンプレート内でカスタム関数を使用した型アサート、
// または main.go でキャスト後に単純に {{.Extra}} を使用）。
type RenderData struct {
	Title       string
	GeneratedAt string // set automatically by Render()
	FilterInfo  string
	NRecords    int

	// 集計のCompute()からの結果 — エンジンレベルのKPIスカラー/分布
	Results []aggregator.Result

	// Extraには継承先のリポジトリのゲーム固有レポート構造体が保持される。
    // 継承先テンプレートはこれを{{ .Extra }}としてアクセスする。
    // 必要に応じて登録済みテンプレート関数で具体的な型にキャストする。
	Extra any
}

// ─────────────────────────────────────────────────
// Render
// ─────────────────────────────────────────────────

// Render は HTML ダッシュボードを outPath に書き込みます。
//   - chartJS: 埋め込み Chart.js ソース (chartjs.Load で取得)。
//   - def:     テンプレートパスの解決にのみ使用されます。
//   - extraFuncs: 社内リポジトリで登録された追加テンプレート関数
//     (例: 型アサーションヘルパー、ゲーム固有のJSONビルダー)。
//     不要な場合はnilを渡してください。
func Render(
	data RenderData,
	outPath string,
	chartJS string,
	def kpidef.KPIDefinition,
	extraFuncs ttmpl.FuncMap,
) error {
	src, err := templateSource(def)
	if err != nil {
		return fmt.Errorf("renderer: template load: %w", err)
	}

	fm := baseFuncMap()
	for k, v := range extraFuncs {
		fm[k] = v // 継承先の関数は基底関数をオーバーライドできる
	}

	tmpl, err := ttmpl.New("kpi").Funcs(fm).Parse(src)
	if err != nil {
		return fmt.Errorf("renderer: template parse: %w", err)
	}

	data.GeneratedAt = time.Now().Format("2006-01-02 15:04:05")

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("renderer: template execute: %w", err)
	}

	// Chart.js を安全にインラインで記述（</script> タグの早期終了を回避）。
	safeJS := strings.ReplaceAll(chartJS, "</script>", `<\/script>`)
	html := strings.Replace(buf.String(), "/* __CHARTJS__ */", safeJS, 1)

	if err := os.WriteFile(outPath, []byte(html), 0644); err != nil {
		return fmt.Errorf("renderer: write %s: %w", outPath, err)
	}
	return nil
}

// ─────────────────────────────────────────────────
// テンプレートソースの選択
// ─────────────────────────────────────────────────

func templateSource(def kpidef.KPIDefinition) (string, error) {
	if p := def.TemplatePath(); p != "" {
		b, err := os.ReadFile(p)
		if err != nil {
			return "", err
		}
		fmt.Printf("  テンプレート: %s\n", p)
		return string(b), nil
	}
	fmt.Println("  テンプレート: built-in")
	return builtinTemplate, nil
}

// ─────────────────────────────────────────────────
// 基本テンプレート関数（ゲーム非依存）
// ─────────────────────────────────────────────────

func baseFuncMap() ttmpl.FuncMap {
	return ttmpl.FuncMap{
		"printf": fmt.Sprintf,
		"inc":    func(i int) int { return i + 1 },
		"hesc":   func(s string) string { return htmpl.HTMLEscapeString(s) },

		// jsonNC converts []NamedCount to a Chart.js-ready JS object literal.
		// Usage in template: {{ jsonNC .SomeSlice }}
		"jsonNC": func(items []aggregator.NamedCount) string {
			labels, data := []string{}, []string{}
			for _, e := range items {
				labels = append(labels, `"`+e.Name+`"`)
				data = append(data, fmt.Sprintf("%.2f", e.Rate))
			}
			return fmt.Sprintf("{labels:[%s],data:[%s]}",
				strings.Join(labels, ","), strings.Join(data, ","))
		},

		// jsonNCCount uses Count instead of Rate.
		"jsonNCCount": func(items []aggregator.NamedCount) string {
			labels, data := []string{}, []string{}
			for _, e := range items {
				labels = append(labels, `"`+e.Name+`"`)
				data = append(data, fmt.Sprintf("%d", e.Count))
			}
			return fmt.Sprintf("{labels:[%s],data:[%s]}",
				strings.Join(labels, ","), strings.Join(data, ","))
		},

		// jsonNF converts []NamedFloat.
		"jsonNF": func(items []aggregator.NamedFloat) string {
			labels, data := []string{}, []string{}
			for _, e := range items {
				labels = append(labels, `"`+e.Name+`"`)
				data = append(data, fmt.Sprintf("%.1f", e.Value))
			}
			return fmt.Sprintf("{labels:[%s],data:[%s]}",
				strings.Join(labels, ","), strings.Join(data, ","))
		},
	}
}

// ─────────────────────────────────────────────────
// 組み込みの最小限のテンプレート
// ─────────────────────────────────────────────────

const builtinTemplate = `<!DOCTYPE html>
<html lang="ja">
<head>
<meta charset="UTF-8">
<title>{{.Title}}</title>
<script>/* __CHARTJS__ */</script>
<style>
body{font-family:system-ui,sans-serif;background:#0f172a;color:#e2e8f0;padding:2rem}
h1{color:#38bdf8;margin-bottom:.5rem}
.meta{color:#94a3b8;font-size:.85rem;margin-bottom:2rem}
table{border-collapse:collapse;width:100%;margin-bottom:2rem}
th,td{padding:.5rem 1rem;border:1px solid #334155;text-align:left}
th{background:#1e293b;color:#38bdf8}
</style>
</head>
<body>
<h1>{{.Title}}</h1>
<p class="meta">生成日時: {{.GeneratedAt}} | フィルタ: {{.FilterInfo}} | レコード数: {{.NRecords}}</p>
<table>
<tr><th>KPI</th><th>値</th></tr>
{{range .Results}}
<tr>
  <td>{{.Metric.Label}}</td>
  <td>
    {{- if .Distribution -}}
      {{range $k,$v := .Distribution}}{{$k}}: {{$v}}&nbsp;{{end}}
    {{- else -}}
      {{printf "%.2f" .Scalar}}
    {{- end -}}
  </td>
</tr>
{{end}}
</table>
<p style="color:#475569;font-size:.8rem">
  ※ これはエンジン組み込みの最小テンプレートです。<br>
  　 KPIDefinition.TemplatePath() で継承先でテンプレートを指定してください。
</p>
</body>
</html>`