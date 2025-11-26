package analytic

type Analytic struct {
	Sum          float64 `json:"Sum"`
	Avg          float64 `json:"Avg"`
	Count        int     `json:"Count"`
	Median       float64 `json:"Median"`
	Percentile90 float64 `json:"Percentile90"`
}

type AnalyticByType struct {
	Income  Analytic            `json:"Income"`
	Expense Analytic            `json:"Expense"`
	All     Analytic            `json:"All"`
	AllMap  map[string]Analytic `json:"-"`
}

type AnalyticGroup struct {
	GroupKey string         `json:"GroupKey"`
	Data     AnalyticByType `json:"Data"`
}

type Analytics struct {
	Summary AnalyticByType  `json:"Summary"`
	Groups  []AnalyticGroup `json:"Groups"`
}

func NewAnalytic(sum float64, avg float64, count int, mediana float64, procentil90 float64) *Analytic {
	return &Analytic{
		Sum:          sum,
		Avg:          avg,
		Count:        count,
		Median:       mediana,
		Percentile90: procentil90,
	}
}
