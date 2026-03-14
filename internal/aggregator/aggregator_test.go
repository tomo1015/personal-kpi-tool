package aggregator

import (
	"errors"
	"testing"

	"github.com/tomo1015/personal-kpi-tool/pkg/kpidef"
)

// -------------------------------------------------------------------
// モック KPIDefinition
// -------------------------------------------------------------------

// mockDef は Compute() のテスト用モック実装。
// computeFn にテストごとの挙動を差し込む。
type mockDef struct {
	title      string
	computeFn  func(rows any, filterInfo string) (*kpidef.RenderData, error)
	tmplPath   string
}

func (m *mockDef) DashboardTitle() string { return m.title }
func (m *mockDef) TemplatePath() string   { return m.tmplPath }
func (m *mockDef) Compute(rows any, filterInfo string) (*kpidef.RenderData, error) {
	return m.computeFn(rows, filterInfo)
}

// -------------------------------------------------------------------
// Compute
// -------------------------------------------------------------------

func TestCompute(t *testing.T) {
	t.Run("正常系: RenderData が返る", func(t *testing.T) {
		def := &mockDef{
			title: "テストレポート",
			computeFn: func(rows any, filterInfo string) (*kpidef.RenderData, error) {
				return &kpidef.RenderData{
					Title:      "テストレポート",
					FilterInfo: filterInfo,
					NUsers:     3,
					Payload:    "ok",
				}, nil
			},
		}

		rd, err := Compute(def, []string{"row1", "row2", "row3"}, "なし")
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if rd.NUsers != 3 {
			t.Errorf("NUsers = %v, want 3", rd.NUsers)
		}
		if rd.FilterInfo != "なし" {
			t.Errorf("FilterInfo = %q, want \"なし\"", rd.FilterInfo)
		}
	})

	t.Run("正常系: Title が空のとき DashboardTitle で補完される", func(t *testing.T) {
		def := &mockDef{
			title: "補完タイトル",
			// Title を空で返す
			computeFn: func(rows any, filterInfo string) (*kpidef.RenderData, error) {
				return &kpidef.RenderData{Title: "", NUsers: 1, Payload: nil}, nil
			},
		}

		rd, err := Compute(def, nil, "なし")
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if rd.Title != "補完タイトル" {
			t.Errorf("Title = %q, want \"補完タイトル\"", rd.Title)
		}
	})

	t.Run("正常系: Title が設定済みのとき DashboardTitle で上書きされない", func(t *testing.T) {
		def := &mockDef{
			title: "上書きされるはずのタイトル",
			computeFn: func(rows any, filterInfo string) (*kpidef.RenderData, error) {
				return &kpidef.RenderData{Title: "設定済みタイトル", NUsers: 1, Payload: nil}, nil
			},
		}

		rd, err := Compute(def, nil, "なし")
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if rd.Title != "設定済みタイトル" {
			t.Errorf("Title = %q, want \"設定済みタイトル\"", rd.Title)
		}
	})

	t.Run("異常系: def が nil → error", func(t *testing.T) {
		_, err := Compute(nil, nil, "なし")
		if err == nil {
			t.Error("エラーを期待したが nil だった")
		}
	})

	t.Run("異常系: Compute がエラーを返す → error", func(t *testing.T) {
		def := &mockDef{
			title: "エラーケース",
			computeFn: func(rows any, filterInfo string) (*kpidef.RenderData, error) {
				return nil, errors.New("集計失敗")
			},
		}

		_, err := Compute(def, nil, "なし")
		if err == nil {
			t.Error("エラーを期待したが nil だった")
		}
	})

	t.Run("異常系: Compute が nil の RenderData を返す → error", func(t *testing.T) {
		def := &mockDef{
			title: "nilケース",
			computeFn: func(rows any, filterInfo string) (*kpidef.RenderData, error) {
				return nil, nil
			},
		}

		_, err := Compute(def, nil, "なし")
		if err == nil {
			t.Error("エラーを期待したが nil だった")
		}
	})
}