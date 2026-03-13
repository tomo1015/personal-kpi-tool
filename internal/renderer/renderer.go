// Package renderer は RenderData を受け取り HTML ファイルへ出力する。
// テンプレートは外部ファイル（TemplatePath）または組み込み最小テンプレートを使用する。
// Chart.js の埋め込みは chartJS 文字列のプレースホルダー置換で行う。
package renderer

import (
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/tomo1015/personal-kpi-tool/pkg/kpidef"
)

// chartJSPlaceholder は HTML テンプレート内の Chart.js 埋め込み位置を示す文字列。
const chartJSPlaceholder = "/* __CHARTJS_PLACEHOLDER__ */"

// Options はレンダラーの動作設定。
type Options struct {
	// OutputPath は出力先 HTML ファイルのパス。
	OutputPath string
	// TemplatePath はカスタムテンプレートのファイルパス。
	// 空文字列の場合は組み込みテンプレート（builtin.html）を使用する。
	TemplatePath string
	// ChartJS は埋め込む Chart.js のソースコード。
	// 空文字列の場合はプレースホルダーをそのまま残す。
	ChartJS string
}

// Render は rd の内容を HTML ファイルに書き出す。
// テンプレートには rd.Payload がそのまま渡されるため、
// テンプレート側で型アサーションまたはカスタム FuncMap を用いること。
func Render(rd *kpidef.RenderData, opts Options) error {
	if rd == nil {
		return fmt.Errorf("RenderData が nil です")
	}

	// テンプレート文字列の取得（外部ファイル優先）
	tmplSrc, err := loadTemplate(opts.TemplatePath)
	if err != nil {
		return fmt.Errorf("テンプレート読み込みエラー: %w", err)
	}

	// 生成日時の補完
	if rd.GeneratedAt == "" {
		rd.GeneratedAt = time.Now().Format("2006-01-02 15:04:05")
	}

	// テンプレートのパースと実行
	tmpl, err := template.New("kpi").Funcs(defaultFuncMap()).Parse(tmplSrc)
	if err != nil {
		return fmt.Errorf("テンプレートパースエラー: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, rd); err != nil {
		return fmt.Errorf("テンプレート実行エラー: %w", err)
	}

	// Chart.js をプレースホルダーと置換
	// </script> タグが含まれているとブラウザが HTML を誤解釈するためエスケープする
	safeJS := strings.ReplaceAll(opts.ChartJS, "</script>", `<\/script>`)
	html := strings.Replace(buf.String(), chartJSPlaceholder, safeJS, 1)

	if err := os.WriteFile(opts.OutputPath, []byte(html), 0644); err != nil {
		return fmt.Errorf("HTML 書き込みエラー(%s): %w", opts.OutputPath, err)
	}
	return nil
}

// loadTemplate はテンプレート文字列を返す。
// path が空文字列の場合は組み込みテンプレートを使用する。
func loadTemplate(path string) (string, error) {
	if path == "" {
		return builtinTemplate, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("テンプレートファイル読み取り失敗(%s): %w", path, err)
	}
	return string(b), nil
}

// defaultFuncMap は全テンプレートで利用可能な共通ヘルパー関数を返す。
// ゲーム固有の関数はカスタムテンプレート側で template.FuncMap を拡張すること。
func defaultFuncMap() template.FuncMap {
	return template.FuncMap{
		// inc は整数を 1 増やす（テンプレート内の連番表示用）
		"inc": func(i int) int { return i + 1 },
		// printf は fmt.Sprintf の別名
		"printf": fmt.Sprintf,
	}
}