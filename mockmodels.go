package main

type MockOrder struct {
	Id        int    `json:"id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	TotalCost int    `json:"total_cost"`
	Status    string `json:"status"`
	ErrorId   *int   `json:"error_id"`
}

type MockOrderCost struct {
	OrderId   int     `json:"order_id"`
	Materials float64 `json:"materials"`
	Labor     float64 `json:"labor"`
	Overhead  float64 `json:"overhead"`
	Total     float64 `json:"total"`
}
