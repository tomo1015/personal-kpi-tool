package renderer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/tomo1015/personal-kpi-tool/pkg/kpidef"
)

// ─────────────────────────────────────────────────
// テスト用ヘルパー
// ─────────────────────────────────────────────────

// newRenderData は最小限の RenderData を返す。
func newRenderData() *kpidef.RenderData {
	return &kpidef.RenderData{
		Title:      "テストダッシュボード",
		FilterInfo: "テストフィルタ",
		NUsers:     10,
		Payload:    map[string]any{"key": "value"},
	}
}

// outputPath は t.TempDir() 配下に出力ファイルパスを生成する。
func outputPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "out.html")
}

// readOutput は出力された HTML ファイルを文字列で返す。
func readOutput(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("出力ファイルの読み込み失敗: %v", err)
	}
	return string(b)
}

// ─────────────────────────────────────────────────
// 正常系
// ─────────────────────────────────────────────────

func TestRender_Builtin(t *testing.T) {
	// 組み込みテンプレートで HTML ファイルが生成されること
	out := outputPath(t)
	err := Render(newRenderData(), Options{OutputPath: out})
	if err != nil {
		t.Fatalf("Render() エラー: %v", err)
	}
	if _, err := os.Stat(out); os.IsNotExist(err) {
		t.Error("出力ファイルが生成されていない")
	}
}

func TestRender_ContainsTitle(t *testing.T) {
	// RenderData.Title が HTML に含まれること
	out := outputPath(t)
	rd := newRenderData()
	if err := Render(rd, Options{OutputPath: out}); err != nil {
		t.Fatalf("Render() エラー: %v", err)
	}
	html := readOutput(t, out)
	if !strings.Contains(html, rd.Title) {
		t.Errorf("HTML に Title %q が含まれていない", rd.Title)
	}
}

func TestRender_GeneratedAtAutoFill(t *testing.T) {
	// GeneratedAt が空のとき自動補完されること
	out := outputPath(t)
	rd := newRenderData()
	rd.GeneratedAt = ""
	if err := Render(rd, Options{OutputPath: out}); err != nil {
		t.Fatalf("Render() エラー: %v", err)
	}
	if rd.GeneratedAt == "" {
		t.Error("GeneratedAt が補完されていない")
	}
	html := readOutput(t, out)
	if !strings.Contains(html, rd.GeneratedAt) {
		t.Errorf("HTML に GeneratedAt %q が含まれていない", rd.GeneratedAt)
	}
}

func TestRender_GeneratedAtPreserved(t *testing.T) {
	// GeneratedAt が設定済みのとき上書きされないこと
	out := outputPath(t)
	rd := newRenderData()
	rd.GeneratedAt = "2026-01-01 00:00:00"
	if err := Render(rd, Options{OutputPath: out}); err != nil {
		t.Fatalf("Render() エラー: %v", err)
	}
	if rd.GeneratedAt != "2026-01-01 00:00:00" {
		t.Errorf("GeneratedAt が上書きされた: %s", rd.GeneratedAt)
	}
}

func TestRender_ChartJSInjected(t *testing.T) {
	// ChartJS 文字列がプレースホルダーと置換されること
	out := outputPath(t)
	js := "console.log('chartjs');"
	if err := Render(newRenderData(), Options{OutputPath: out, ChartJS: js}); err != nil {
		t.Fatalf("Render() エラー: %v", err)
	}
	html := readOutput(t, out)
	if !strings.Contains(html, js) {
		t.Errorf("HTML に ChartJS %q が含まれていない", js)
	}
	if strings.Contains(html, chartJSPlaceholder) {
		t.Error("プレースホルダーが残っている")
	}
}

func TestRender_ChartJSEscaped(t *testing.T) {
	// ChartJS 内の </script> が <\/script> にエスケープされること
	out := outputPath(t)
	js := "var x = '</script>';"
	if err := Render(newRenderData(), Options{OutputPath: out, ChartJS: js}); err != nil {
		t.Fatalf("Render() エラー: %v", err)
	}
	html := readOutput(t, out)
	if strings.Contains(html, "</script>'") {
		t.Error("</script> がエスケープされていない")
	}
	if !strings.Contains(html, `<\/script>`) {
		t.Error("<\\/script> が HTML に含まれていない")
	}
}

func TestRender_ExternalTemplate(t *testing.T) {
	// 外部テンプレートが組み込みテンプレートより優先されること
	dir := t.TempDir()
	tmplPath := filepath.Join(dir, "custom.html")
	tmplContent := `<html><body>CUSTOM:{{.Title}}</body></html>`
	if err := os.WriteFile(tmplPath, []byte(tmplContent), 0644); err != nil {
		t.Fatalf("テンプレートファイル作成失敗: %v", err)
	}

	out := filepath.Join(dir, "out.html")
	if err := Render(newRenderData(), Options{OutputPath: out, TemplatePath: tmplPath}); err != nil {
		t.Fatalf("Render() エラー: %v", err)
	}
	html := readOutput(t, out)
	if !strings.Contains(html, "CUSTOM:") {
		t.Error("外部テンプレートが使用されていない")
	}
	if !strings.Contains(html, "テストダッシュボード") {
		t.Error("外部テンプレートに Title が展開されていない")
	}
}

func TestRender_ExtraFuncs(t *testing.T) {
	// ExtraFuncs で登録したカスタム関数がテンプレートで使えること
	dir := t.TempDir()
	tmplPath := filepath.Join(dir, "extra.html")
	// {{greet}} を呼び出すテンプレート
	if err := os.WriteFile(tmplPath, []byte(`{{greet}}`), 0644); err != nil {
		t.Fatalf("テンプレートファイル作成失敗: %v", err)
	}

	out := filepath.Join(dir, "out.html")
	err := Render(newRenderData(), Options{
		OutputPath:   out,
		TemplatePath: tmplPath,
		ExtraFuncs: template.FuncMap{
			"greet": func() string { return "HELLO_EXTRA" },
		},
	})
	if err != nil {
		t.Fatalf("Render() エラー: %v", err)
	}
	html := readOutput(t, out)
	if !strings.Contains(html, "HELLO_EXTRA") {
		t.Errorf("ExtraFuncs の関数が展開されていない: %s", html)
	}
}

// ─────────────────────────────────────────────────
// 異常系
// ─────────────────────────────────────────────────

func TestRender_NilRenderData(t *testing.T) {
	// rd が nil のときエラーが返ること
	err := Render(nil, Options{OutputPath: outputPath(t)})
	if err == nil {
		t.Error("nil RenderData でエラーが返らなかった")
	}
}

func TestRender_InvalidOutputPath(t *testing.T) {
	// 存在しないディレクトリへの書き込みはエラーになること
	err := Render(newRenderData(), Options{OutputPath: "/nonexistent/dir/out.html"})
	if err == nil {
		t.Error("不正な OutputPath でエラーが返らなかった")
	}
}

func TestRender_InvalidTemplatePath(t *testing.T) {
	// 存在しないテンプレートパスのときエラーが返ること
	err := Render(newRenderData(), Options{
		OutputPath:   outputPath(t),
		TemplatePath: "/nonexistent/template.html",
	})
	if err == nil {
		t.Error("不正な TemplatePath でエラーが返らなかった")
	}
}

func TestRender_InvalidTemplate(t *testing.T) {
	// 構文エラーのテンプレートでエラーが返ること
	dir := t.TempDir()
	tmplPath := filepath.Join(dir, "broken.html")
	if err := os.WriteFile(tmplPath, []byte(`{{.Unclosed`), 0644); err != nil {
		t.Fatalf("テンプレートファイル作成失敗: %v", err)
	}
	err := Render(newRenderData(), Options{
		OutputPath:   filepath.Join(dir, "out.html"),
		TemplatePath: tmplPath,
	})
	if err == nil {
		t.Error("構文エラーのテンプレートでエラーが返らなかった")
	}
}
