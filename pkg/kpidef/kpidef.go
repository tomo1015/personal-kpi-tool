// Package kpidef は KPI 集計エンジンが依存する共通インターフェースと定数を定義する。
// ゲーム固有の実装はこのパッケージに依存し、独自の KPIDefinition を実装すること。
package kpidef

// AggFunc は集計関数の種別を表す定数型。
type AggFunc int

const (
	// AggSum は合計集計を示す。
	AggSum AggFunc = iota
	// AggAvg は平均集計を示す。
	AggAvg
	// AggCount は件数集計を示す。
	AggCount
	// AggRate は割合集計（%）を示す。
	AggRate
	// AggDistribution は分布集計を示す。
	AggDistribution
)

// RenderData は集計結果をレンダラーに渡すための汎用コンテナ。
// 実装側は任意の構造体をこの型にラップして返す。
type RenderData struct {
	// Title はダッシュボードのタイトル文字列。
	Title string
	// GeneratedAt は生成日時（フォーマット済み文字列）。
	GeneratedAt string
	// FilterInfo はフィルタ条件の説明文字列。
	FilterInfo string
	// NUsers は集計対象のユーザー数。
	NUsers int
	// Payload はゲーム固有の集計結果を格納する任意の値。
	// renderer はこの値をテンプレートに渡す。
	Payload any
}

// KPIDefinition はゲーム固有の KPI 集計ロジックを抽象化するインターフェース。
//
// engine 側は KPIDefinition を受け取り、以下の順序で処理を実行する:
//  1. aggregator.Compute(def, rows) を呼び出す
//  2. 返却された RenderData を renderer.Render に渡す
//
// ゲーム固有の実装はこのインターフェースを実装し、
// 集計ロジックをすべて Compute メソッドに閉じ込める。
type KPIDefinition interface {
	// DashboardTitle はレポートの表示タイトルを返す。
	DashboardTitle() string

	// Compute は rows（任意のレコードスライス）を受け取り、
	// 集計結果を RenderData として返す。
	// rows は []T 型を any にラップして渡す想定。
	// エラーが発生した場合は nil, error を返す。
	Compute(rows any, filterInfo string) (*RenderData, error)

	// TemplatePath はレンダリングに使用する HTML テンプレートのファイルパスを返す。
	// 空文字列を返した場合は renderer の組み込みテンプレートを使用する。
	TemplatePath() string
}