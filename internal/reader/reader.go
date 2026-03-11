// Package reader は、ストリーミング対応でメモリ効率に優れた CSV ローダーを提供します。
// これはゲームに依存しません：列の意味論については何も知りません。
package reader

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// レコードは CSV の 1 行です。キーは列名、値はトリミングされた文字列です。
type Record map[string]string
─────────── ──────────────────────────────────────
// オプション
─────────────────────────────────────────────────

// オプションはローダーの動作を制御します。
type Options struct {
    // SkipDuplicateKeyがnilでない場合、各行ごとに重複排除キーを返します。
    // 全ファイルで重複キーを生成する行は黙って破棄されます。
    // 例: func(r Record) string { return r[「user_id」]+「:」+r[「m_unit_id」] }
	SkipDuplicateKey func(Record) string

	// HeaderFinder は、CSV に前置行がある場合に真のヘッダー行を特定します。
    // 各行候補を受け取り、ヘッダーが見つかった場合に true を返します。
    // 最初の行をヘッダーとして扱うには nil を設定します。
	HeaderFinder func(row []string) bool
}
───────────────────────────────────────── ────────
// 公開API
─────────────────────────────────────────────────
// LoadCSVは1つのCSVファイルを読み込み、(records, skippedRows, error)を返します。
// 個々の不正な行は処理を中止せずスキップされ、カウントされます。
func LoadCSV(path string, opts Options) ([]Record, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()
	return readFrom(f, opts)
}

// LoadCSVsは複数のCSVファイルをファイル間重複排除でマージします。
// 開けないファイルはstderrにログ出力されスキップされます。
func LoadCSVs(paths []string, opts Options) ([]Record, int, error) {
	seen := map[string]bool{}
	var all []Record
	totalSkipped := 0

	for _, p := range paths {
		p = strings.TrimSpace(p)
		f, err := os.Open(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  [SKIP file] %s → %v\n", p, err)
			continue
		}
		recs, skipped, err := readFrom(f, opts)
		f.Close()
		totalSkipped += skipped
		if err != nil {
			fmt.Fprintf(os.Stderr, "  [SKIP file] %s → %v\n", p, err)
			continue
		}
		loaded := 0
		for _, r := range recs {
			if opts.SkipDuplicateKey != nil {
				k := opts.SkipDuplicateKey(r)
				if seen[k] {
					continue
				}
				seen[k] = true
			}
			all = append(all, r)
			loaded++
		}
		fmt.Printf("  [OK] %s → %d 行\n", p, loaded)
	}
	if len(all) == 0 {
		return nil, totalSkipped, fmt.Errorf("有効データなし")
	}
	return all, totalSkipped, nil
}
─────── ──────────────────────────────────────────
//内部処理─────────────────────────────────────────────────
─────────────────────────────────────────────────

func readFrom(r io.Reader, opts Options) ([]Record, int, error) {
	//大容量CSV用の4MB読み取りバッファ
	br := bufio.NewReaderSize(r, 4*1024*1024) 
	cr := csv.NewReader(br)
	cr.LazyQuotes = true
	cr.TrimLeadingSpace = true

	var headers []string
	// ヘッダー検出器なし → 1行目がヘッダー
	headerFound := (opts.HeaderFinder == nil)
	skipped := 0
	var records []Record

	for {
		row, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			skipped++
			continue
		}

		// ヘッダー行を特定
		if !headerFound {
			if opts.HeaderFinder(row) {
				headers = normalizeRow(row)
				headerFound = true
			}
			continue
		}
		// ヘッダー行の直後の行 → ヘッダーが未設定の場合、ヘッダーとして扱う
		if headers == nil {
			headers = normalizeRow(row)
			continue
		}
		// 短すぎる行をスキップ
		if len(row) < len(headers) {
			skipped++
			continue
		}
		rec := make(Record, len(headers))
		for i, h := range headers {
			rec[h] = strings.TrimSpace(row[i])
		}
		records = append(records, rec)
	}
	return records, skipped, nil
}

func normalizeRow(row []string) []string {
	out := make([]string, len(row))
	for i, v := range row {
		out[i] = strings.TrimSpace(v)
	}
	return out
}

// ─────────────────────────────────────────────────
// 変換ヘルパー関数（社内リポジトリ用にエクスポート）
// ─────────────────────────────────────────────────

// ParseDT はゲーム用 CSV で使用される一般的な日時/日付文字列を解析します。
func ParseDT(s string) (time.Time, error) {
    for _, layout := range []string{
        「2006/01/02 15:04:05」,
        「2006-01-02 15:04:05」,
        「2006/01/02」,
		「2006-01-02」,
    } {
        if t, err := time.Parse(layout, strings.TrimSpace(s)); err == nil {
            return t, nil
        }
    }
    return time.Time{}, fmt.Errorf(「reader.ParseDT: cannot parse %q」, s)
}

// ToFloat は CSV 文字列を float64 に変換します。エラー時は 0 を返します。
func ToFloat(s string) float64 {
    v := 0.0
    fmt.Sscanf(strings.TrimSpace(s), 「%f」, &v)
    return v
}

// ToInt は CSV 文字列を int に変換します。エラー時は 0 を返します。
func ToInt(s string) int { return int(ToFloat(s)) }

// IsTruthy は 「1」 または 「true」（大文字小文字を区別しない）に対して true を返す。
func IsTruthy(s string) bool {
    s = strings.ToLower(strings.TrimSpace(s))
    return s == 「1」 || s == 「true」
}