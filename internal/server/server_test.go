package server

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ─────────────────────────────────────────────────
// テスト用ヘルパー
// ─────────────────────────────────────────────────

// newTestServer はテスト用のサーバーを返す。
// uploadDir / reportPath は t.TempDir() 配下に自動生成する。
func newTestServer(t *testing.T, cfg Config) *Server {
	t.Helper()
	if cfg.UploadDir == "" {
		cfg.UploadDir = t.TempDir()
	}
	if cfg.ReportPath == "" {
		cfg.ReportPath = filepath.Join(t.TempDir(), "report.html")
	}
	if cfg.Username == "" {
		cfg.Username = "admin"
	}
	if cfg.Password == "" {
		cfg.Password = "password"
	}
	return New(cfg)
}

// get は認証付きの GET リクエストを発行してレスポンスを返す。
func get(t *testing.T, srv *Server, path, user, pass string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if user != "" {
		req.SetBasicAuth(user, pass)
	}
	rr := httptest.NewRecorder()
	srv.mux.ServeHTTP(rr, req)
	return rr.Result()
}

// post は認証付きの POST リクエストを発行してレスポンスを返す。
func post(t *testing.T, srv *Server, path, user, pass string, body io.Reader, contentType string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, body)
	req.Header.Set("Content-Type", contentType)
	if user != "" {
		req.SetBasicAuth(user, pass)
	}
	rr := httptest.NewRecorder()
	srv.mux.ServeHTTP(rr, req)
	return rr.Result()
}

// buildCSVUpload は multipart フォームの CSV アップロードリクエストボディを生成する。
func buildCSVUpload(t *testing.T, filename, content string) (io.Reader, string) {
	t.Helper()
	var buf strings.Builder
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("unit_csv", filename)
	if err != nil {
		t.Fatalf("フォームファイル作成失敗: %v", err)
	}
	fmt.Fprint(fw, content)
	w.Close()
	return strings.NewReader(buf.String()), w.FormDataContentType()
}

// bodyString はレスポンスボディを文字列で返す。
func bodyString(t *testing.T, resp *http.Response) string {
	t.Helper()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("レスポンスボディ読み取り失敗: %v", err)
	}
	return string(b)
}

// ─────────────────────────────────────────────────
// /health
// ─────────────────────────────────────────────────

func TestHealth(t *testing.T) {
	t.Run("認証なしで 200 が返る", func(t *testing.T) {
		srv := newTestServer(t, Config{})
		resp := get(t, srv, "/health", "", "")
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
	})
}

// ─────────────────────────────────────────────────
// basicAuth
// ─────────────────────────────────────────────────

func TestBasicAuth_Unauthorized(t *testing.T) {
	t.Run("認証情報なしで 401 が返る", func(t *testing.T) {
		srv := newTestServer(t, Config{})
		resp := get(t, srv, "/upload", "", "")
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401", resp.StatusCode)
		}
	})
}

func TestBasicAuth_WrongPassword(t *testing.T) {
	t.Run("誤パスワードで 401 が返る", func(t *testing.T) {
		srv := newTestServer(t, Config{})
		resp := get(t, srv, "/upload", "admin", "wrong")
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401", resp.StatusCode)
		}
	})
}

func TestBasicAuth_OK(t *testing.T) {
	t.Run("正しい認証情報で 200 が返る", func(t *testing.T) {
		srv := newTestServer(t, Config{})
		resp := get(t, srv, "/upload", "admin", "password")
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
	})
}

// ─────────────────────────────────────────────────
// handleUpload
// ─────────────────────────────────────────────────

func TestHandleUpload_OK(t *testing.T) {
	t.Run("CSV アップロードが成功しファイルが保存される", func(t *testing.T) {
		srv := newTestServer(t, Config{})
		body, ct := buildCSVUpload(t, "unit.csv", "id,user_id\n1,u001\n")
		resp := post(t, srv, "/upload", "admin", "password", body, ct)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
		// ファイルが保存されているか確認
		dest := filepath.Join(srv.cfg.UploadDir, "unit.csv")
		if _, err := os.Stat(dest); os.IsNotExist(err) {
			t.Error("unit.csv が保存されていない")
		}
		// 成功メッセージが含まれているか確認
		b := bodyString(t, resp)
		if !strings.Contains(b, "アップロードしました") {
			t.Errorf("成功メッセージが含まれていない: %s", b)
		}
	})
}

func TestHandleUpload_NotCSV(t *testing.T) {
	t.Run("CSV 以外のファイルはエラーメッセージが返る", func(t *testing.T) {
		srv := newTestServer(t, Config{})
		body, ct := buildCSVUpload(t, "unit.txt", "hello")
		resp := post(t, srv, "/upload", "admin", "password", body, ct)

		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
		b := bodyString(t, resp)
		if !strings.Contains(b, "CSV ファイルのみ") {
			t.Errorf("エラーメッセージが含まれていない: %s", b)
		}
	})
}

func TestHandleUpload_NoField(t *testing.T) {
	t.Run("unit_csv フィールドなしでエラーメッセージが返る", func(t *testing.T) {
		srv := newTestServer(t, Config{})
		// フィールド名を間違えて送信
		var buf strings.Builder
		w := multipart.NewWriter(&buf)
		fw, _ := w.CreateFormFile("wrong_field", "unit.csv")
		fmt.Fprint(fw, "data")
		w.Close()

		resp := post(t, srv, "/upload", "admin", "password",
			strings.NewReader(buf.String()), w.FormDataContentType())

		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
		b := bodyString(t, resp)
		if !strings.Contains(b, "unit_csv フィールドが見つかりません") {
			t.Errorf("エラーメッセージが含まれていない: %s", b)
		}
	})
}

// ─────────────────────────────────────────────────
// handleAnalyze
// ─────────────────────────────────────────────────

func TestHandleAnalyze_NoCSV(t *testing.T) {
	t.Run("CSV 未アップロード時に 400 が返る", func(t *testing.T) {
		srv := newTestServer(t, Config{
			AnalyzeFunc: func(csvPath string) error { return nil },
		})
		resp := post(t, srv, "/analyze", "admin", "password", nil, "application/x-www-form-urlencoded")
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", resp.StatusCode)
		}
	})
}

func TestHandleAnalyze_OK(t *testing.T) {
	t.Run("分析成功後に /report へリダイレクトされる", func(t *testing.T) {
		srv := newTestServer(t, Config{
			AnalyzeFunc: func(csvPath string) error { return nil },
		})
		// unit.csv を事前に配置
		dest := filepath.Join(srv.cfg.UploadDir, "unit.csv")
		os.WriteFile(dest, []byte("id,user_id\n1,u001\n"), 0644)

		resp := post(t, srv, "/analyze", "admin", "password", nil, "application/x-www-form-urlencoded")
		if resp.StatusCode != http.StatusSeeOther {
			t.Errorf("status = %d, want 303", resp.StatusCode)
		}
		if loc := resp.Header.Get("Location"); loc != "/report" {
			t.Errorf("Location = %q, want /report", loc)
		}
	})
}

func TestHandleAnalyze_Error(t *testing.T) {
	t.Run("AnalyzeFunc がエラーを返すと 500 が返る", func(t *testing.T) {
		srv := newTestServer(t, Config{
			AnalyzeFunc: func(csvPath string) error {
				return fmt.Errorf("集計エラー")
			},
		})
		dest := filepath.Join(srv.cfg.UploadDir, "unit.csv")
		os.WriteFile(dest, []byte("id,user_id\n1,u001\n"), 0644)

		resp := post(t, srv, "/analyze", "admin", "password", nil, "application/x-www-form-urlencoded")
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("status = %d, want 500", resp.StatusCode)
		}
	})
}

func TestHandleAnalyze_NoFunc(t *testing.T) {
	t.Run("AnalyzeFunc が nil のとき 500 が返る", func(t *testing.T) {
		srv := newTestServer(t, Config{})
		dest := filepath.Join(srv.cfg.UploadDir, "unit.csv")
		os.WriteFile(dest, []byte("id,user_id\n1,u001\n"), 0644)

		resp := post(t, srv, "/analyze", "admin", "password", nil, "application/x-www-form-urlencoded")
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("status = %d, want 500", resp.StatusCode)
		}
	})
}

// ─────────────────────────────────────────────────
// handleReport
// ─────────────────────────────────────────────────

func TestHandleReport_NotFound(t *testing.T) {
	t.Run("レポート未生成時に 404 が返る", func(t *testing.T) {
		srv := newTestServer(t, Config{})
		resp := get(t, srv, "/report", "admin", "password")
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("status = %d, want 404", resp.StatusCode)
		}
	})
}

func TestHandleReport_OK(t *testing.T) {
	t.Run("レポートが存在する場合に 200 が返る", func(t *testing.T) {
		dir := t.TempDir()
		reportPath := filepath.Join(dir, "report.html")
		os.WriteFile(reportPath, []byte("<html>report</html>"), 0644)

		srv := newTestServer(t, Config{ReportPath: reportPath})
		resp := get(t, srv, "/report", "admin", "password")
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
		b := bodyString(t, resp)
		if !strings.Contains(b, "report") {
			t.Errorf("レポート内容が返っていない: %s", b)
		}
	})
}

// ─────────────────────────────────────────────────
// handleInsight
// ─────────────────────────────────────────────────

func TestHandleInsight_Get(t *testing.T) {
	t.Run("GET で考察ページが返る", func(t *testing.T) {
		srv := newTestServer(t, Config{})
		resp := get(t, srv, "/insight", "admin", "password")
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
		b := bodyString(t, resp)
		if !strings.Contains(b, "AI 考察") {
			t.Errorf("考察ページが返っていない: %s", b)
		}
	})
}

func TestHandleInsight_Post_OK(t *testing.T) {
	t.Run("InsightFunc 成功時に考察テキストが返る", func(t *testing.T) {
		srv := newTestServer(t, Config{
			InsightFunc: func() (string, error) {
				return "## 総評\nテスト考察です。", nil
			},
		})
		resp := post(t, srv, "/insight", "admin", "password", nil, "application/x-www-form-urlencoded")
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
		b := bodyString(t, resp)
		if !strings.Contains(b, "テスト考察です") {
			t.Errorf("考察テキストが含まれていない: %s", b)
		}
	})
}

func TestHandleInsight_Post_Error(t *testing.T) {
	t.Run("InsightFunc エラー時にエラーメッセージが返る", func(t *testing.T) {
		srv := newTestServer(t, Config{
			InsightFunc: func() (string, error) {
				return "", fmt.Errorf("API エラー")
			},
		})
		resp := post(t, srv, "/insight", "admin", "password", nil, "application/x-www-form-urlencoded")
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
		b := bodyString(t, resp)
		if !strings.Contains(b, "API エラー") {
			t.Errorf("エラーメッセージが含まれていない: %s", b)
		}
	})
}

func TestHandleInsight_NoFunc(t *testing.T) {
	t.Run("InsightFunc が nil のとき 500 が返る", func(t *testing.T) {
		srv := newTestServer(t, Config{})
		resp := post(t, srv, "/insight", "admin", "password", nil, "application/x-www-form-urlencoded")
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("status = %d, want 500", resp.StatusCode)
		}
	})
}
