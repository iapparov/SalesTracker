package dto

type AnalyticsReq struct {
	From    string `json:"from"`
	To      string `json:"to"`
	GroupBy string `json:"groupBy"` // day|week|month|category|none
	SplitBy string `json:"splitBy"` // type|category|none
	SortBy  string `json:"sortBy"`  // sum|avg|count|median|percentile90
	SortDir string `json:"sortDir"` // asc|desc
}

type GetTransactionReq struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Type     string `json:"type"` // income|expense|all
	Category string `json:"category"`
	SortBy   string `json:"sortBy"`  // id|type|category|amount|date
	SortDir  string `json:"sortDir"` // asc|desc
}

type SaveTransactionReq struct {
	Type        string  `json:"type"` // income|expense
	Category    string  `json:"category"`
	Amount      float64 `json:"amount"`
	Date        string  `json:"date"`
	Description string  `json:"description"`
}
