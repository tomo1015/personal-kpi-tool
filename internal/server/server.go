// Package server は CSV アップロード・分析実行・レポート表示を提供する HTTP サーバー。
// ベーシック認証・ファイルアップロード・分析トリガーの汎用実装を含む。
// ゲーム固有の分析ロジックは Config.AnalyzeFunc に注入する。
package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// Config はサーバーの動作設定。
type Config struct {
	// Addr は listen するアドレス（例: ":8080"）。
	Addr string
	// Username / Password はベーシック認証の認証情報。
	Username string
	Password string
	// UploadDir はアップロードされた CSV の保存先ディレクトリ。
	UploadDir string
	// ReportPath は生成済みダッシュボード HTML のファイルパス。
	ReportPath string
	// AnalyzeFunc はアップロード後に呼び出す分析関数。
	// 引数は保存された unit CSV のパス。
	AnalyzeFunc func(csvPath string) error
	// InsightFunc はAI考察生成ボタン押下時に呼び出す関数。
	// 生成された Markdownテキストを返す
	InsightFunc func() (string, error)
}

// Server は HTTP サーバーの状態を保持する。
type Server struct {
	cfg Config
	mux *http.ServeMux
}

// New は Config を受け取り Server を初期化して返す。
func New(cfg Config) *Server {
	s := &Server{cfg: cfg, mux: http.NewServeMux()}
	s.routes()
	return s
}

// Run は HTTP サーバーを起動する。
func (s *Server) Run() error {
	fmt.Printf("サーバー起動: http://localhost%s\n", s.cfg.Addr)
	return http.ListenAndServe(s.cfg.Addr, s.mux)
}

// routes はルーティングを設定する。
func (s *Server) routes() {
	s.mux.HandleFunc("/upload", s.basicAuth(s.handleUploadPage))
	s.mux.HandleFunc("/analyze", s.basicAuth(s.handleAnalyze))
	s.mux.HandleFunc("/report", s.basicAuth(s.handleReport))
	s.mux.HandleFunc("/insight", s.basicAuth(s.handleInsight))
	// ルートは /upload にリダイレクト
	s.mux.HandleFunc("/", s.basicAuth(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/upload", http.StatusFound)
	}))
	// /health はベーシック認証なし（ALB ヘルスチェック用）
	s.mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

// basicAuth はベーシック認証ミドルウェア。
func (s *Server) basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != s.cfg.Username || pass != s.cfg.Password {
			w.Header().Set("WWW-Authenticate", `Basic realm="KPI Dashboard"`)
			http.Error(w, "認証が必要です", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

// handleUploadPage は GET でアップロード画面を、POST で CSV 保存を行う。
func (s *Server) handleUploadPage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.renderUploadPage(w, "", "")
	case http.MethodPost:
		s.handleUpload(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// handleUpload は multipart/form-data で送られた unit CSV を UploadDir に保存する。
func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	// 最大 4GB まで受け付ける
	if err := r.ParseMultipartForm(4 << 30); err != nil {
		s.renderUploadPage(w, "error", "ファイルの解析に失敗しました: "+err.Error())
		return
	}

	file, header, err := r.FormFile("unit_csv")
	if err != nil {
		s.renderUploadPage(w, "error", "unit_csv フィールドが見つかりません: "+err.Error())
		return
	}
	defer file.Close()

	// 拡張子チェック
	if filepath.Ext(header.Filename) != ".csv" {
		s.renderUploadPage(w, "error", "CSV ファイルのみアップロード可能です")
		return
	}

	// 保存先ディレクトリ作成
	if err := os.MkdirAll(s.cfg.UploadDir, 0755); err != nil {
		s.renderUploadPage(w, "error", "保存先ディレクトリの作成に失敗しました: "+err.Error())
		return
	}

	// unit.csv として固定ファイル名で保存
	dest := filepath.Join(s.cfg.UploadDir, "unit.csv")
	out, err := os.Create(dest)
	if err != nil {
		s.renderUploadPage(w, "error", "ファイルの保存に失敗しました: "+err.Error())
		return
	}
	defer out.Close()

	written, err := io.Copy(out, file)
	if err != nil {
		s.renderUploadPage(w, "error", "ファイルの書き込みに失敗しました: "+err.Error())
		return
	}

	s.renderUploadPage(w, "success",
		fmt.Sprintf("%s をアップロードしました（%.1f MB）", header.Filename, float64(written)/1024/1024))
}

// handleAnalyze は分析を実行し、完了後に /report にリダイレクトする。
func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	csvPath := filepath.Join(s.cfg.UploadDir, "unit.csv")
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		http.Error(w, "unit.csv がアップロードされていません", http.StatusBadRequest)
		return
	}

	if s.cfg.AnalyzeFunc == nil {
		http.Error(w, "AnalyzeFunc が設定されていません", http.StatusInternalServerError)
		return
	}

	if err := s.cfg.AnalyzeFunc(csvPath); err != nil {
		http.Error(w, "分析エラー: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/report", http.StatusSeeOther)
}

// handleReport は生成済みの dashboard.html を返す。
func (s *Server) handleReport(w http.ResponseWriter, r *http.Request) {
	if _, err := os.Stat(s.cfg.ReportPath); os.IsNotExist(err) {
		http.Error(w, "レポートがまだ生成されていません。先に分析を実行してください。", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeFile(w, r, s.cfg.ReportPath)
}

// renderUploadPage はアップロード画面の HTML を返す。
// status: "" | "success" | "error"
func (s *Server) renderUploadPage(w http.ResponseWriter, status, message string) {
	var statusHTML string
	switch status {
	case "success":
		statusHTML = fmt.Sprintf(
			`<p class="msg success">✅ %s</p>`, message)
	case "error":
		statusHTML = fmt.Sprintf(
			`<p class="msg error">❌ %s</p>`, message)
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="ja">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>KPI ダッシュボード — CSV アップロード</title>
<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body { background: #0f172a; color: #e2e8f0; font-family: system-ui, sans-serif;
       display: flex; align-items: center; justify-content: center; min-height: 100vh; }
.card { background: #1e293b; border: 1px solid #334155; border-radius: 1rem;
        padding: 2rem; width: 100%%; max-width: 480px; }
h1 { color: #38bdf8; font-size: 1.3rem; margin-bottom: 1.5rem; }
label { display: block; color: #94a3b8; font-size: .85rem; margin-bottom: .4rem; }
input[type=file] { width: 100%%; background: #0f172a; border: 1px solid #334155;
                   border-radius: .5rem; padding: .6rem; color: #e2e8f0;
                   font-size: .85rem; margin-bottom: 1rem; cursor: pointer; }
.btn { display: block; width: 100%%; padding: .75rem;
       background: #38bdf8; color: #0f172a; font-weight: 700;
       border: none; border-radius: .5rem; cursor: pointer; font-size: 1rem;
       margin-bottom: .75rem; }
.btn:hover { background: #7dd3fc; }
.btn.analyze { background: #818cf8; color: #fff; }
.btn.analyze:hover { background: #a5b4fc; }
.msg { padding: .75rem 1rem; border-radius: .5rem; font-size: .85rem; margin-bottom: 1rem; }
.success { background: #34d39922; color: #34d399; border: 1px solid #34d399; }
.error   { background: #f8717122; color: #f87171; border: 1px solid #f87171; }
.divider { border: none; border-top: 1px solid #334155; margin: 1rem 0; }
</style>
</head>
<body>
<div class="card">
  <h1>📊 KPI ダッシュボード</h1>
  %s
  <form method="POST" action="/upload" enctype="multipart/form-data">
    <label>unit CSV ファイルを選択</label>
    <input type="file" name="unit_csv" accept=".csv" required>
    <button type="submit" class="btn">アップロード</button>
  </form>
  <hr class="divider">
  <form method="POST" action="/analyze">
    <button type="submit" class="btn analyze">▶ 分析を実行してレポートを表示</button>
  </form>
   <hr class="divider">
   <a href="/insight" style="color:#818cf8;font-size:.85rem;">🤖 AI 考察を見る</a>
</div>
</body>
</html>`, statusHTML)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

func (s *Server) handleInsight(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.renderInsightPage(w, "", "")
	case http.MethodPost:
		if s.cfg.InsightFunc == nil {
			http.Error(w, "InsightFunc が設定されていません", http.StatusInternalServerError)
			return
		}
		insight, err := s.cfg.InsightFunc()
		if err != nil {
			s.renderInsightPage(w, "error", err.Error())
			return
		}
		s.renderInsightPage(w, "success", insight)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) renderInsightPage(w http.ResponseWriter, status, content string) {
	var bodyHTML string
	switch status {
	case "success":
		// Markdown を <pre> で表示（シンプル実装）
		bodyHTML = fmt.Sprintf(`<div class="result"><pre>%s</pre></div>`, content)
	case "error":
		bodyHTML = fmt.Sprintf(`<p class="msg error">❌ %s</p>`, content)
	}
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="ja">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>KPI AI 考察</title>
<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body { background: #0f172a; color: #e2e8f0; font-family: system-ui, sans-serif; padding: 2rem; }
h1 { color: #38bdf8; font-size: 1.3rem; margin-bottom: 1.5rem; }
.card { background: #1e293b; border: 1px solid #334155; border-radius: 1rem; padding: 1.5rem; max-width: 800px; margin-bottom: 1.5rem; }
.btn { display: inline-block; padding: .75rem 2rem; background: #818cf8; color: #fff;
       font-weight: 700; border: none; border-radius: .5rem; cursor: pointer; font-size: 1rem; }
.btn:hover { background: #a5b4fc; }
.result { background: #0f172a; border: 1px solid #334155; border-radius: .5rem; padding: 1.5rem; }
pre { white-space: pre-wrap; word-break: break-word; font-family: system-ui, sans-serif; font-size: .9rem; line-height: 1.7; }
.msg { padding: .75rem 1rem; border-radius: .5rem; font-size: .85rem; margin-bottom: 1rem; }
.error { background: #f8717122; color: #f87171; border: 1px solid #f87171; }
.nav { margin-bottom: 1.5rem; }
.nav a { color: #38bdf8; text-decoration: none; font-size: .85rem; margin-right: 1rem; }
</style>
</head>
<body>
<h1>🤖 AI 考察レポート</h1>
<div class="nav">
  <a href="/upload">← アップロード画面</a>
  <a href="/report">📊 ダッシュボード</a>
</div>
<div class="card">
  <form method="POST" action="/insight">
    <button type="submit" class="btn">✨ AI 考察を生成する</button>
  </form>
</div>
%s
</body>
</html>`, bodyHTML)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}
