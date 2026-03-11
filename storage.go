package main

import (
	"database/sql"
	"fmt"
	"strconv"
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
		_, err := tx.Exec("insert into overhead(order_id, date, prod_type, amount) values($1, $2, $3, $4)", item.OrderID, item.Date, item.ProdType, item.Amount)
		if err != nil {
			return err
		}
	}
	return nil
}

func SaveError(csvErr *CSVError) error {
	if csvErr == nil {
		return fmt.Errorf("CSVError is nil")
	}
	_, err := db.Exec("insert into errors(cause, row_and_column) values ($1, $2)",
		"Файл: "+csvErr.FileName+", Ошибка: "+csvErr.Cause,
		"Строка: "+strconv.FormatInt(int64(csvErr.Row), 64)+", Стоблец: "+csvErr.Column)
	return err
}

func (err *CSVError) Error() string {
	return fmt.Sprintf("Файл: %s, строка: %d, столбец: %s, ошибка: %s", err.FileName, err.Row, err.Column, err.Cause)
}
