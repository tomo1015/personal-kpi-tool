// Package reader は CSV ファイルの汎用ストリーム読み込みと
// 型変換ユーティリティを提供する。
package reader

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

// Record は CSV の 1 行を表す。
// キーはトリム済みのヘッダー名、値はトリム済みのセル文字列。
type Record map[string]string

// StreamCSV は path のファイルを開き、ヘッダー行を読み取った後、
// データ行を 1 行ずつ fn に渡す。fn が error を返した場合は即座に中断する。
// ファイルのクローズは StreamCSV 内で行うため、呼び出し側は意識しなくてよい。
func StreamCSV(path string, fn func(rec Record) error) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("ファイルを開けません(%s): %w", path, err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	// 先頭行をヘッダーとして読み取る
	headers, err := r.Read()
	if err != nil {
		return fmt.Errorf("ヘッダー読み取りエラー(%s): %w", path, err)
	}
	// ヘッダー名の前後空白をトリム
	for i, h := range headers {
		headers[i] = strings.TrimSpace(h)
	}

	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("CSV 読み取りエラー(%s): %w", path, err)
		}
		rec := make(Record, len(headers))
		for i, h := range headers {
			if i < len(row) {
				rec[h] = strings.TrimSpace(row[i])
			} else {
				rec[h] = ""
			}
		}
		if err := fn(rec); err != nil {
			return err
		}
	}
	return nil
}

// ReadAllCSV は path のファイル全行を []Record として返す。
// ヘッダー行は含まない。
func ReadAllCSV(path string) ([]Record, error) {
	var recs []Record
	err := StreamCSV(path, func(rec Record) error {
		recs = append(recs, rec)
		return nil
	})
	return recs, err
}

// Headers は path の CSV のヘッダー行だけを返す。
// 大量ファイルのスキーマ確認用。
func Headers(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("ファイルを開けません(%s): %w", path, err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	row, err := r.Read()
	if err != nil {
		return nil, err
	}
	out := make([]string, len(row))
	for i, h := range row {
		out[i] = strings.TrimSpace(h)
	}
	return out, nil
}

// -------------------------------------------------------------------
// 型変換ユーティリティ
// -------------------------------------------------------------------

// supportedLayouts は ParseDT が試みる日時フォーマットの一覧。
var supportedLayouts = []string{
	"2006/01/02 15:04:05",
	"2006-01-02 15:04:05",
	"2006/01/02",
	"2006-01-02",
}

// ParseDT は文字列 s を time.Time に変換する。
// 対応フォーマット: "YYYY/MM/DD HH:MM:SS", "YYYY-MM-DD HH:MM:SS",
// "YYYY/MM/DD", "YYYY-MM-DD"。
// いずれにも該当しない場合は error を返す。
func ParseDT(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	for _, layout := range supportedLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("日時フォーマット不明: %q", s)
}

// ToInt は文字列 s を int に変換する。変換失敗時は 0 を返す。
func ToInt(s string) int {
	v, _ := strconv.Atoi(strings.TrimSpace(s))
	return v
}

// ToFloat は文字列 s を float64 に変換する。変換失敗時は 0 を返す。
func ToFloat(s string) float64 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return v
}

// IsTruthy は "1", "true"（大文字小文字不問）を true と判定する。
// その他はすべて false を返す。
func IsTruthy(s string) bool {
	s = strings.TrimSpace(s)
	return s == "1" || strings.EqualFold(s, "true")
}

// Get は Record から col の値を返す。存在しない場合は空文字列を返す。
func Get(rec Record, col string) string {
	return rec[col]
}

// GetInt は Record から col の値を int として返す。
func GetInt(rec Record, col string) int {
	return ToInt(rec[col])
}

// GetBool は Record から col の値を IsTruthy で bool として返す。
func GetBool(rec Record, col string) bool {
	return IsTruthy(rec[col])
}