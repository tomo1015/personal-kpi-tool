package reader

import (
	"os"
	"testing"
	"time"
)

// -------------------------------------------------------------------
// ParseDT
// -------------------------------------------------------------------

func TestParseDT(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    time.Time
		wantErr bool
	}{
		{
			name: "YYYY/MM/DD HH:MM:SS",
			in:   "2026/01/15 12:30:00",
			want: time.Date(2026, 1, 15, 12, 30, 0, 0, time.UTC),
		},
		{
			name: "YYYY-MM-DD HH:MM:SS",
			in:   "2026-01-15 12:30:00",
			want: time.Date(2026, 1, 15, 12, 30, 0, 0, time.UTC),
		},
		{
			name: "YYYY/MM/DD",
			in:   "2026/01/15",
			want: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "YYYY-MM-DD",
			in:   "2026-01-15",
			want: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "前後に空白を含む",
			in:   "  2026-01-15  ",
			want: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "不明なフォーマット",
			in:      "01/15/2026",
			wantErr: true,
		},
		{
			name:    "空文字列",
			in:      "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDT(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseDT(%q): エラーを期待したが nil だった", tt.in)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseDT(%q): 予期しないエラー: %v", tt.in, err)
				return
			}
			if !got.Equal(tt.want) {
				t.Errorf("ParseDT(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

// -------------------------------------------------------------------
// ToInt
// -------------------------------------------------------------------

func TestToInt(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{name: "正の整数",     in: "42",    want: 42},
		{name: "負の整数",     in: "-10",   want: -10},
		{name: "ゼロ",         in: "0",     want: 0},
		{name: "前後に空白",   in: "  7  ", want: 7},
		{name: "空文字列",     in: "",      want: 0},
		{name: "数値以外",     in: "abc",   want: 0},
		{name: "小数（切捨）", in: "3.14",  want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToInt(tt.in)
			if got != tt.want {
				t.Errorf("ToInt(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

// -------------------------------------------------------------------
// ToFloat
// -------------------------------------------------------------------

func TestToFloat(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want float64
	}{
		{name: "整数文字列",   in: "42",     want: 42.0},
		{name: "小数文字列",   in: "3.14",   want: 3.14},
		{name: "負の値",       in: "-1.5",   want: -1.5},
		{name: "前後に空白",   in: "  2.5 ", want: 2.5},
		{name: "空文字列",     in: "",       want: 0},
		{name: "数値以外",     in: "abc",    want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToFloat(tt.in)
			if got != tt.want {
				t.Errorf("ToFloat(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

// -------------------------------------------------------------------
// IsTruthy
// -------------------------------------------------------------------

func TestIsTruthy(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: `"1" は true`,          in: "1",     want: true},
		{name: `"true" は true`,       in: "true",  want: true},
		{name: `"TRUE" は true`,       in: "TRUE",  want: true},
		{name: `"True" は true`,       in: "True",  want: true},
		{name: `"0" は false`,         in: "0",     want: false},
		{name: `"false" は false`,     in: "false", want: false},
		{name: `空文字列は false`,     in: "",      want: false},
		{name: `"yes" は false`,       in: "yes",   want: false},
		{name: `前後に空白 + "1"`,     in: " 1 ",   want: true},
		{name: `前後に空白 + "true"`,  in: " true", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsTruthy(tt.in)
			if got != tt.want {
				t.Errorf("IsTruthy(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

// -------------------------------------------------------------------
// Get / GetInt / GetBool
// -------------------------------------------------------------------

func TestGet(t *testing.T) {
	rec := Record{"user_id": "u001", "score": "99", "active": "1", "empty": ""}

	t.Run("存在するキー", func(t *testing.T) {
		got := Get(rec, "user_id")
		if got != "u001" {
			t.Errorf("Get = %q, want %q", got, "u001")
		}
	})

	t.Run("存在しないキー → 空文字列", func(t *testing.T) {
		got := Get(rec, "no_such_key")
		if got != "" {
			t.Errorf("Get = %q, want \"\"", got)
		}
	})

	t.Run("値が空文字列のキー", func(t *testing.T) {
		got := Get(rec, "empty")
		if got != "" {
			t.Errorf("Get = %q, want \"\"", got)
		}
	})
}

func TestGetInt(t *testing.T) {
	rec := Record{"score": "42", "broken": "abc"}

	t.Run("数値文字列 → int", func(t *testing.T) {
		got := GetInt(rec, "score")
		if got != 42 {
			t.Errorf("GetInt = %v, want 42", got)
		}
	})

	t.Run("数値以外 → 0", func(t *testing.T) {
		got := GetInt(rec, "broken")
		if got != 0 {
			t.Errorf("GetInt = %v, want 0", got)
		}
	})

	t.Run("存在しないキー → 0", func(t *testing.T) {
		got := GetInt(rec, "no_such_key")
		if got != 0 {
			t.Errorf("GetInt = %v, want 0", got)
		}
	})
}

func TestGetBool(t *testing.T) {
	rec := Record{"flag_on": "1", "flag_true": "true", "flag_off": "0"}

	t.Run(`"1" → true`, func(t *testing.T) {
		if !GetBool(rec, "flag_on") {
			t.Error("GetBool(flag_on) = false, want true")
		}
	})

	t.Run(`"true" → true`, func(t *testing.T) {
		if !GetBool(rec, "flag_true") {
			t.Error("GetBool(flag_true) = false, want true")
		}
	})

	t.Run(`"0" → false`, func(t *testing.T) {
		if GetBool(rec, "flag_off") {
			t.Error("GetBool(flag_off) = true, want false")
		}
	})

	t.Run("存在しないキー → false", func(t *testing.T) {
		if GetBool(rec, "no_such_key") {
			t.Error("GetBool(no_such_key) = true, want false")
		}
	})
}

// -------------------------------------------------------------------
// StreamCSV / ReadAllCSV
// -------------------------------------------------------------------

// writeTempCSV はテスト用の一時 CSV ファイルを作成して返す。
// テスト終了時に自動削除される。
func writeTempCSV(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "test_*.csv")
	if err != nil {
		t.Fatalf("一時ファイル作成失敗: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("一時ファイル書き込み失敗: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestStreamCSV(t *testing.T) {
	t.Run("正常系: 複数行を読み込める", func(t *testing.T) {
		csv := "user_id,score\nu001,10\nu002,20\n"
		path := writeTempCSV(t, csv)

		var recs []Record
		err := StreamCSV(path, func(rec Record) error {
			recs = append(recs, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if len(recs) != 2 {
			t.Fatalf("レコード数 = %v, want 2", len(recs))
		}
		if recs[0]["user_id"] != "u001" || recs[0]["score"] != "10" {
			t.Errorf("1行目 = %v, want {user_id:u001 score:10}", recs[0])
		}
		if recs[1]["user_id"] != "u002" || recs[1]["score"] != "20" {
			t.Errorf("2行目 = %v, want {user_id:u002 score:20}", recs[1])
		}
	})

	t.Run("正常系: ヘッダーのみ（データ行なし）", func(t *testing.T) {
		path := writeTempCSV(t, "user_id,score\n")

		var recs []Record
		err := StreamCSV(path, func(rec Record) error {
			recs = append(recs, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if len(recs) != 0 {
			t.Errorf("レコード数 = %v, want 0", len(recs))
		}
	})

	t.Run("正常系: ヘッダー名の前後空白がトリムされる", func(t *testing.T) {
		path := writeTempCSV(t, " user_id , score \nu001,10\n")

		var recs []Record
		err := StreamCSV(path, func(rec Record) error {
			recs = append(recs, rec)
			return nil
		})
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		// トリム済みのキーでアクセスできることを確認する
		if recs[0]["user_id"] != "u001" {
			t.Errorf("user_id = %q, want \"u001\"", recs[0]["user_id"])
		}
	})

	t.Run("境界値: 存在しないファイル → error", func(t *testing.T) {
		err := StreamCSV("/no/such/file.csv", func(rec Record) error { return nil })
		if err == nil {
			t.Error("エラーを期待したが nil だった")
		}
	})
}

func TestReadAllCSV(t *testing.T) {
	t.Run("正常系: 全行が返る", func(t *testing.T) {
		csv := "id,name\n1,Alice\n2,Bob\n3,Carol\n"
		path := writeTempCSV(t, csv)

		recs, err := ReadAllCSV(path)
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if len(recs) != 3 {
			t.Fatalf("レコード数 = %v, want 3", len(recs))
		}
		if recs[2]["name"] != "Carol" {
			t.Errorf("recs[2][name] = %q, want \"Carol\"", recs[2]["name"])
		}
	})

	t.Run("境界値: ヘッダーのみ → 空スライス", func(t *testing.T) {
		path := writeTempCSV(t, "id,name\n")

		recs, err := ReadAllCSV(path)
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if len(recs) != 0 {
			t.Errorf("レコード数 = %v, want 0", len(recs))
		}
	})

	t.Run("境界値: 存在しないファイル → error", func(t *testing.T) {
		_, err := ReadAllCSV("/no/such/file.csv")
		if err == nil {
			t.Error("エラーを期待したが nil だった")
		}
	})
}