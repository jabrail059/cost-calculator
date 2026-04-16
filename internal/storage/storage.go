package storage

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/lib/pq"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/models"
)

var db *sql.DB

func SetDB(database *sql.DB) {
	db = database
}

func DB() *sql.DB {
	return db
}

type CSVError struct {
	FileName string
	Row      int
	Column   string
	Cause    string
}

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
		fmt.Sprintf("Файл: %s, Ошибка: %s", csvErr.FileName, csvErr.Cause),
		fmt.Sprintf("Строка: %s, Столбец: %s", strconv.Itoa(csvErr.Row), csvErr.Column))
	return err
}

func (err *CSVError) Error() string {
	return fmt.Sprintf("Файл: %s, строка: %d, столбец: %s, ошибка: %s", err.FileName, err.Row, err.Column, err.Cause)
}

func CalculateCost(OrderID int) (*models.CostResponse, error) {
	start := time.Now()
	defer func() {
		log.Printf("CalculateCost for order %d took %v", OrderID, time.Since(start))
	}()
	bomRows, err := db.Query("select quantity, unit_cost from boms where order_id = $1", OrderID)
	if err != nil {
		return nil, err
	}
	defer bomRows.Close()

	laborRows, err := db.Query("select rate, hours from labor where order_id = $1", OrderID)
	if err != nil {
		return nil, err
	}
	defer laborRows.Close()

	overheadRows, err := db.Query("select amount from overhead where order_id = $1", OrderID)
	if err != nil {
		return nil, err
	}
	defer overheadRows.Close()

	costResp := models.CostResponse{}

	for bomRows.Next() {
		var cost, qnty float64
		err := bomRows.Scan(&qnty, &cost)
		if err != nil {
			return nil, err
		}
		costResp.Materials += cost * qnty
	}

	for laborRows.Next() {
		var rt, hrs float64
		err := laborRows.Scan(&rt, &hrs)
		if err != nil {
			return nil, err
		}
		costResp.Labor += rt * hrs
	}

	for overheadRows.Next() {
		var amt float64
		err := overheadRows.Scan(&amt)
		if err != nil {
			return nil, err
		}
		costResp.Overhead += amt
	}
	costResp.Total = costResp.Labor + costResp.Materials + costResp.Overhead
	return &costResp, nil
}

func ValidateOrders(orderIds []int) error {
	if len(orderIds) == 0 {
		return nil
	}

	problem := make([]int, 0)
	found := make(map[int]string)

	rows, err := db.Query("select id, status from orders where id = ANY($1)", pq.Array(orderIds))
	if err != nil {
		return fmt.Errorf("Ошибка при проверке заказа: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id   int
			stat string
		)
		err = rows.Scan(&id, &stat)
		if err != nil {
			return err
		}
		found[id] = stat
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf("Ошибка при обработке строк: %w", err)
	}

	for _, value := range orderIds {
		if _, ok := found[value]; !ok {
			problem = append(problem, value)
		}
	}

	for value, status := range found {
		if status == "completed" || status == "closed" {
			problem = append(problem, value)
		}
	}

	if len(problem) != 0 {
		return fmt.Errorf("Следующие заказы не были найдены или имеют запрещённый статус: %v", problem)
	}
	return nil
}

func SaveUploadLog(tx *sql.Tx, orderIDs []int, filetype string, changedBy string) error {
	for _, orderId := range orderIDs {
		_, err := tx.Exec("insert into upload_log(order_id, file_type, changed_by) values ($1, $2, $3)", orderId, filetype, changedBy)
		if err != nil {
			return fmt.Errorf("Ошибка логирования изменения данных: %w", err)
		}
	}
	return nil
}
