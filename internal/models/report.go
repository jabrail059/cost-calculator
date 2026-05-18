package models

import "time"

type ReportItem struct {
	Order string  `json:"order"`
	Type  string  `json:"type"`
	Sum   float64 `json:"sum"`
}

type ReportRequest struct {
	Status string       `json:"status"`
	Date   []ReportItem `json:"date"`
}

type Report struct {
	ID             int           `json:"id"`
	CalculationID  int           `json:"calculation_id"`
	CreatedBy      int           `json:"created_by"`
	OneCReportData ReportRequest `json:"onec_report_data"`
	CreatedAt      time.Time     `json:"created_at"`
}

type CreateReportRequest struct {
	ReportData CalculationResult `json:"report_data"`
}

type GenerateReportRequest struct {
	CalculationID      int `json:"calculation_id"`
	CalculationIDCamel int `json:"calculationId,omitempty"`
}
