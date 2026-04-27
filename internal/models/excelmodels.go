package models

type ReportItem struct {
	Order string  `json:"order"`
	Type  string  `json:"type"`
	Sum   float64 `json:"sum"`
}

type ReportRequest struct {
	Status string       `json:"status"`
	Date   []ReportItem `json:"date"`
}
