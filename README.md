# personal-kpi-tool

Go で実装した KPI 分析エンジンの基盤ライブラリです。CSV を読み込み、ゲーム固有の集計ロジック（`KPIDefinition`）で集計し、HTML ダッシュボードとして出力する一連のパイプラインを提供します。

## 必要な環境

- **Go 1.25** 以上

## プロジェクト構成

| ディレクトリ | 説明 |
|-------------|------|
| `cmd/analyzer` | エンジン動作確認用 CLI。サンプル KPI で CSV → 集計 → HTML 出力を実行 |
| `example` | `KPIDefinition` のサンプル実装（`SampleDef`）。自社リポジトリ実装時の参考用 |
| `internal/` | サーバー・集計・レンダラ・CSV 読み込み・Chart.js 読み込みなどの内部実装 |
| `pkg/` | 外部モジュール（company repo 等）から利用する公開 API |

### 公開パッケージ（pkg/）

- **`pkg/kpidef`** — `KPIDefinition` インターフェース、`RenderData`、集計種別（`AggFunc`）の定義
- **`pkg/csvreader`** — CSV の一括読み込み（`ReadAllCSV`）、ストリーム読み込み（`StreamCSV`）、レコード取得（`Get` / `GetInt` / `GetBool`）、日時パース（`ParseDT`）
- **`pkg/aggregator`** — `Compute(def, rows, filterInfo)` で集計実行
- **`pkg/renderer`** — `Render(rd, opts)` で HTML レポート出力（組み込み or カスタムテンプレート、Chart.js 埋め込み）
- **`pkg/server`** — CSV アップロード・分析実行・レポート表示を行う HTTP サーバー（ベーシック認証対応）
- **`pkg/mathutil`** — 集計用ユーティリティ（`Sum` / `Avg` / `Mean` / `Median` / `Max64` / `Count` / `Rate` / `Distribution` / `GroupBy`）
- **`pkg/chartjs`** — Chart.js ソースの読み込み（`Load()`）

## 使い方

### 1. ライブラリとして利用（company repo など）

1. 本モジュールを import する。
2. `kpidef.KPIDefinition` を実装する（`DashboardTitle` / `Compute` / `TemplatePath`）。
3. CSV を `pkg/csvreader` で読み、`pkg/aggregator.Compute(def, rows, filterInfo)` で集計。
4. 得られた `RenderData` を `pkg/renderer.Render(rd, opts)` で HTML に出力。
5. 必要なら `pkg/server` でアップロード画面・分析トリガー・レポート表示用の HTTP サーバーを立て、`Config.AnalyzeFunc` に「CSV パスを受け取り上記 3〜4 を実行する関数」を渡す。

### 2. 動作確認用 CLI（cmd/analyzer）

サンプル `KPIDefinition` でパイプラインを試す場合:

```bash
# ビルド
go build -o analyzer ./cmd/analyzer

# ダミーデータでレポート生成（output: sample_report.html）
./analyzer

# 指定 CSV でレポート生成
./analyzer -input data.csv -output report.html

# カスタムテンプレートを指定
./analyzer -input data.csv -output report.html -template ./templates/dashboard.html
```

## 処理の流れ

1. **CSV 読み込み** — `csvreader.ReadAllCSV(path)` または `csvreader.StreamCSV(path, fn)`
2. **集計** — `aggregator.Compute(kpidef, rows, filterInfo)` → `*kpidef.RenderData`
3. **HTML 出力** — `renderer.Render(rd, renderer.Options{...})`（Chart.js は `chartjs.Load()` で取得して `Options.ChartJS` に渡す）

ゲーム固有ロジックはすべて `KPIDefinition.Compute` に閉じ込めます。

## 開発

```bash
# テスト
go test ./...

# ビルド確認
go build ./...

# Lint（CI でも実行）
golangci-lint run
```

## リリース

`main` ブランチへのプッシュで [semantic-release](https://github.com/semantic-release/semantic-release) により自動リリースされます。

コミットメッセージは [Conventional Commits](https://www.conventionalcommits.org/ja/v1.0.0/) に従ってください。

| プレフィックス | リリース | 例 |
|---------------|---------|-----|
| `feat:` | minor | `feat: ログイン機能を追加` |
| `fix:` | patch | `fix: ボタンが押せないバグを修正` |
| `BREAKING CHANGE:` | major | フッターに `BREAKING CHANGE: APIを変更` |

## ライセンス

MIT License。詳細は [LICENSE](LICENSE) を参照してください。
