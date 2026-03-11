package kpidef
//===============================================
// 集計関数
//===============================================
type AggFunc string

const (
    AggSum AggFunc = "sum"
    AggAvg AggFunc = "avg"
    AggCount AggFunc = "count"
    AggRate AggFunc = "rate"
    AddDistribution AggFunc = "distribution"
)

//===============================================
// 指標：1つのKPI項目
//===============================================
type Metric struct {
    Name string // internal key（内部キー）e.g. "avg_unit_lv"
    Label string // display label(表示ラベル) e.g. "平均ユニットレベル"
    Func AggFunc // aggregation function(集計関数) e.g. "avg"
    Field string // cav column name to aggregate(集計するカラム名) e.g. "unit_lv"
}

//===============================================
// KPIDefinition：継承先で実装するKPI指標
//===============================================
// ここは唯一の拡張ポイントです
// 継承先のリポジトリはこのインターフェースを満たす構造体を作成し
// tool.run() に渡します
type KPIDefinition interface {
    // DashboardTitleはHTMLレポートに表示されるタイトルを返します
    DashboardTitle() string
    //Metricsは計算対象のKPI項目のリストを返す
    Metrics() []Metric
    // TemplatePathはカスタムHTMLテンプレートの絶対パスを返す
    // ” ”を返すとエンジンの組み込みテンプレートを使用する
    TemplatePath() string
}