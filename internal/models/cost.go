package models

type CostResponse struct {
	Materials float64 `json:"materials"`
	Labor     float64 `json:"labor"`
	Overhead  float64 `json:"overhead"`
	Total     float64 `json:"total"`
}

type CalculationResult struct {
	CalculationID int     `json:"calculationId,omitempty"`
	BomCost       float64 `json:"bomCost"`
	LaborCost     float64 `json:"laborCost"`
	OverheadCost  float64 `json:"overheadCost"`
	TotalCost     float64 `json:"totalCost"`
}
