package models

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
