// Package csvreader は internal/reader の公開ラッパーです。
// company repo など外部モジュールから CSV 読み込み関数を利用するために提供します。
package csvreader

import (
	"time"

	"github.com/tomo1015/personal-kpi-tool/internal/reader"
)

// ReadAllCSV は CSV ファイルを全行読み込み、カラム名をキーとするマップのスライスを返します。
func ReadAllCSV(path string) ([]map[string]string, error) {
	return reader.ReadAllCSV(path)
}

// Get は rec から key に対応する文字列値を返します。
func Get(rec map[string]string, key string) string {
	return reader.Get(rec, key)
}

// GetInt は rec から key に対応する int 値を返します。
func GetInt(rec map[string]string, key string) int {
	return reader.GetInt(rec, key)
}

// GetBool は rec から key に対応する bool 値を返します。
func GetBool(rec map[string]string, key string) bool {
	return reader.GetBool(rec, key)
}

// ParseDT は日時文字列を time.Time にパースします。
func ParseDT(s string) (time.Time, error) {
	return reader.ParseDT(s)
}