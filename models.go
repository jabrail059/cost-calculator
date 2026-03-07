package main

type BOMItem struct {
	OrderID      int
	Quantity     float64
	UnitCost     float64
	MaterialCode string
}

type OverheadItem struct {
	OrderID  int
	ProdType string
	Amount   float64
}

type LaborItem struct {
	OrderID int
	Rate    float64
	Hours   float64
}
