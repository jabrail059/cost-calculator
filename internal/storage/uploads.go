package storage

import (
	"database/sql"
	"fmt"
	"strconv"

	"gitverse.ru/topit/12-40_team20_Zueva/internal/models"
)

func InsertBOMItems(tx *sql.Tx, items []models.BOMItem) error {
	for _, item := range items {
		_, err := tx.Exec("insert into boms(order_id, quantity, unit_cost, material_code) values($1, $2, $3, $4)", item.OrderID, item.Quantity, item.UnitCost, item.MaterialCode)
		if err != nil {
			return err
		}
	}
	return nil
}

func InsertLaborItems(tx *sql.Tx, items []models.LaborItem) error {
	for _, item := range items {
		_, err := tx.Exec("insert into labor(order_id, rate, hours) values($1, $2, $3)", item.OrderID, item.Rate, item.Hours)
		if err != nil {
			return err
		}
	}
	return nil
}

func InsertOverheadItems(tx *sql.Tx, items []models.OverheadItem) error {
	for _, item := range items {
		_, err := tx.Exec("insert into overhead(date, prod_type, amount) values($1, $2, $3)", item.Date, item.ProdType, item.Amount)
		if err != nil {
			return err
		}
	}
	return nil
}

func SaveError(csvErr *models.CSVError) error {
	if csvErr == nil {
		return fmt.Errorf("CSVError is nil")
	}
	_, err := db.Exec("insert into errors(cause, row_and_column) values ($1, $2)",
		fmt.Sprintf("Файл: %s, Ошибка: %s", csvErr.FileName, csvErr.Cause),
		fmt.Sprintf("Строка: %s, Столбец: %s", strconv.Itoa(csvErr.Row), csvErr.Column))
	return err
}
