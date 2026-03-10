type KPIDefinition interface {
    ColumnMap() map[string]string
    Metrics() []Metric
    DashboardTitle() string
}

type Metric struct {
    Name  string
    Label string
    Func  AggFunc  // Sum / Avg / Count / Rate / Distribution
    Field string
}