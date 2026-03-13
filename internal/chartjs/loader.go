// Package chartjs は Chart.js のソースコードを CDN からダウンロードし、
// 実行ファイルと同じディレクトリにキャッシュする機能を提供する。
package chartjs

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const (
	// cdnURL は Chart.js の CDN URL。
	cdnURL = "https://cdnjs.cloudflare.com/ajax/libs/Chart.js/4.4.1/chart.umd.min.js"
	// cacheFileName はキャッシュファイル名。
	cacheFileName = "chart.umd.min.js"
)

// Load は Chart.js のソースコードを文字列で返す。
// 実行ファイルと同じディレクトリにキャッシュが存在する場合はそれを使用し、
// 存在しない場合は CDN からダウンロードしてキャッシュに保存する。
func Load() (string, error) {
	cachePath, err := resolveCachePath()
	if err != nil {
		// キャッシュパスの解決に失敗しても CDN から直接取得を試みる
		fmt.Fprintf(os.Stderr, "  Chart.js: キャッシュパス解決失敗（%v）、直接ダウンロードします\n", err)
		return download("")
	}

	// キャッシュが有効であれば読み取って返す（1KB 以上であることを確認）
	if data, err := os.ReadFile(cachePath); err == nil && len(data) > 1024 {
		fmt.Printf("  Chart.js: キャッシュ使用 (%s)\n", cachePath)
		return string(data), nil
	}

	fmt.Println("  Chart.js: ダウンロード中...")
	return download(cachePath)
}

// resolveCachePath は実行ファイルと同じディレクトリのキャッシュパスを返す。
func resolveCachePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("実行ファイルパス取得失敗: %w", err)
	}
	return filepath.Join(filepath.Dir(exe), cacheFileName), nil
}

// download は CDN から Chart.js をダウンロードする。
// cachePath が空文字列でなければダウンロード結果をキャッシュに保存する。
func download(cachePath string) (string, error) {
	resp, err := http.Get(cdnURL)
	if err != nil {
		return "", fmt.Errorf("Chart.js ダウンロード失敗: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Chart.js ダウンロード HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Chart.js 読み込み失敗: %w", err)
	}

	// キャッシュ保存（失敗してもエラーにしない）
	if cachePath != "" {
		if err := os.WriteFile(cachePath, body, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "  Chart.js: キャッシュ保存失敗（%v）\n", err)
		} else {
			fmt.Printf("  Chart.js: ダウンロード完了 (%d bytes)、キャッシュ保存 (%s)\n", len(body), cachePath)
		}
	}

	return string(body), nil
}