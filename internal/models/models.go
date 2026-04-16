package models

import "time"

type OrderIDer interface {
	GetOrderId() int
}

func (bomId BOMItem) GetOrderId() int {
	return bomId.OrderID
}

func (overheadId OverheadItem) GetOrderId() int {
	return overheadId.OrderID
}

func (laborId LaborItem) GetOrderId() int {
	return laborId.OrderID
}

type BOMItem struct {
	OrderID      int
	Quantity     float64
	UnitCost     float64
	MaterialCode string
}

type OverheadItem struct {
	OrderID  int
	Date     string
	ProdType string
	Amount   float64
}

type LaborItem struct {
	OrderID int
	Rate    float64
	Hours   float64
}

type OrderResponse struct {
	Id        int        `json:"id"`
	StartDate time.Time  `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
	TotalCost float64    `json:"total_cost"`
	Status    string     `json:"status"`
	ErrorId   *int       `json:"error_id"`
}

type CostResponse struct {
	Materials float64 `json:"materials"`
	Labor     float64 `json:"labor"`
	Overhead  float64 `json:"overhead"`
	Total     float64 `json:"total"`
}

type CalculationResult struct {
	BomCost      float64 `json:"bomCost"`
	LaborCost    float64 `json:"laborCost"`
	OverheadCost float64 `json:"overheadCost"`
	TotalCost    float64 `json:"totalCost"`
}

type BOMItemResponse struct {
	ID           int     `json:"id"`
	OrderID      int     `json:"order_id"`
	Quantity     float64 `json:"quantity"`
	UnitCost     float64 `json:"unit_cost"`
	MaterialCode string  `json:"material_code"`
}

type LaborItemResponse struct {
	ID      int     `json:"id"`
	OrderID int     `json:"order_id"`
	Rate    float64 `json:"rate"`
	Hours   float64 `json:"hours"`
}

type OverheadItemResponse struct {
	ID       int     `json:"id"`
	OrderID  int     `json:"order_id"`
	Date     string  `json:"date"`
	ProdType string  `json:"prod_type"`
	Amount   float64 `json:"amount"`
}

type Log struct {
	Id         int       `json:"id"`
	OrderID    int       `json:"order_id"`
	Filetype   string    `json:"file_type"`
	UploadedAt time.Time `json:"uploaded_at"`
	ChangedBy  string    `json:"changed_by"`
}
