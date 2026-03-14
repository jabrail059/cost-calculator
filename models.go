package main

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

type CSVError struct {
	FileName string
	Row      int
	Column   string
	Cause    string
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
