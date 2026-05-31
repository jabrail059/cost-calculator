package models

import "time"

type OrderIDer interface {
	GetOrderId() int
}

func (bomId BOMItem) GetOrderId() int {
	return bomId.OrderID
}

func (overheadId OverheadItem) GetOrderId() int {
	if overheadId.OrderID == nil {
		return 0
	}
	return *overheadId.OrderID
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
	OrderID  *int
	Date     string
	ProdType string
	Amount   float64
}

type LaborItem struct {
	OrderID int
	Rate    float64
	Hours   float64
}

type Log struct {
	Id         int       `json:"id"`
	OrderID    int       `json:"order_id"`
	Filetype   string    `json:"file_type"`
	UploadedAt time.Time `json:"uploaded_at"`
	ChangedBy  string    `json:"changed_by"`
}
