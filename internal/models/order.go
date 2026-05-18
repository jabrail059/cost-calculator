package models

import "time"

type OrderResponse struct {
	Id        int        `json:"id"`
	StartDate time.Time  `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
	TotalCost float64    `json:"total_cost"`
	Status    string     `json:"status"`
	ErrorId   *int       `json:"error_id"`
}

type CreateOrderRequest struct {
	ID        int    `json:"id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Status    string `json:"status"`
}
