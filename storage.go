package main

import (
	"database/sql"
)

func InsertBOMItems(tx *sql.Tx, items []BOMItem) error {
	for _, item := range items {
		_, err := tx.Exec("insert into boms(order_id, quantity, unit_cost, material_code) values($1, $2, $3, $4)", item.OrderID, item.Quantity, item.UnitCost, item.MaterialCode)
		if err != nil {
			return err
		}
	}
	return nil
}

func InsertLaborItems(tx *sql.Tx, items []LaborItem) error {
	for _, item := range items {
		_, err := tx.Exec("insert into labor(order_id, rate, hours) values($1, $2, $3)", item.OrderID, item.Rate, item.Hours)
		if err != nil {
			return err
		}
	}
	return nil
}

func InsertOverheadItems(tx *sql.Tx, items []OverheadItem) error {
	for _, item := range items {
		_, err := tx.Exec("insert into overhead(order_id, prod_type, amount) values($1, $2, $3)", item.OrderID, item.ProdType, item.Amount)
		if err != nil {
			return err
		}
	}
	return nil
}
