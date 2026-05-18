package storage

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
	"gitverse.ru/topit/12-40_team20_Zueva/internal/models"
)

func GetOrders() ([]models.OrderResponse, error) {
	rows, err := db.Query("select id, start_date, end_date, total_cost, status, error_id from orders")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := []models.OrderResponse{}
	for rows.Next() {
		order := models.OrderResponse{}
		err = rows.Scan(&order.Id, &order.StartDate, &order.EndDate, &order.TotalCost, &order.Status, &order.ErrorId)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

func GetOrderByID(id int) (*models.OrderResponse, error) {
	order := models.OrderResponse{}
	err := db.QueryRow("select id, start_date, end_date, total_cost, status, error_id from orders where id = $1", id).
		Scan(&order.Id, &order.StartDate, &order.EndDate, &order.TotalCost, &order.Status, &order.ErrorId)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func GetOrderBOMs(orderID int) ([]models.BOMItemResponse, error) {
	rows, err := db.Query("SELECT id, order_id, quantity, unit_cost, material_code FROM boms WHERE order_id = $1", orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []models.BOMItemResponse{}
	for rows.Next() {
		item := models.BOMItemResponse{}
		err := rows.Scan(&item.ID, &item.OrderID, &item.Quantity, &item.UnitCost, &item.MaterialCode)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func GetOrderLabor(orderID int) ([]models.LaborItemResponse, error) {
	rows, err := db.Query("SELECT id, order_id, rate, hours FROM labor WHERE order_id = $1", orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []models.LaborItemResponse{}
	for rows.Next() {
		item := models.LaborItemResponse{}
		err := rows.Scan(&item.ID, &item.OrderID, &item.Rate, &item.Hours)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func GetOrderOverhead(orderID int) ([]models.OverheadItemResponse, error) {
	rows, err := db.Query("SELECT id, order_id, date, prod_type, amount FROM overhead WHERE order_id = $1", orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []models.OverheadItemResponse{}
	for rows.Next() {
		item := models.OverheadItemResponse{}
		err := rows.Scan(&item.ID, &item.OrderID, &item.Date, &item.ProdType, &item.Amount)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func OrderExists(id int) (bool, error) {
	var exists int
	err := db.QueryRow("SELECT 1 FROM orders WHERE id=$1", id).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func CreateOrder(id int, startDate time.Time, endDate any, status string) error {
	_, err := db.Exec("INSERT INTO orders(id, start_date, end_date, status) VALUES ($1, $2, $3, $4)",
		id,
		startDate,
		endDate,
		status)
	return err
}

func CalculateCost(OrderID int, Method string) (*models.CostResponse, error) {
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

	overheadCost, err := DistributeOverhead(OrderID, Method)
	if err != nil {
		return nil, err
	}

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
	costResp.Overhead = overheadCost
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

func DistributeOverhead(orderID int, method string) (float64, error) {
	var endDate time.Time
	err := db.QueryRow("SELECT COALESCE(end_date, start_date) FROM orders WHERE id = $1", orderID).Scan(&endDate)
	if err != nil {
		return 0, fmt.Errorf("не удалось определить дату заказа: %w", err)
	}
	startOfMonth := time.Date(endDate.Year(), endDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	rowsOverhead, err := db.Query(`
        SELECT amount FROM overhead 
        WHERE date >= $1 AND date < $2
    `, startOfMonth, endOfMonth)
	if err != nil {
		return 0, fmt.Errorf("ошибка получения накладных: %w", err)
	}
	defer rowsOverhead.Close()

	var totalOverhead float64
	for rowsOverhead.Next() {
		var amount float64
		if err := rowsOverhead.Scan(&amount); err != nil {
			return 0, err
		}
		totalOverhead += amount
	}

	if totalOverhead == 0 {
		return 0, nil
	}

	rowsOrders, err := db.Query(`
        SELECT id FROM orders 
        WHERE end_date >= $1 AND end_date < $2
    `, startOfMonth, endOfMonth)
	if err != nil {
		return 0, fmt.Errorf("ошибка получения заказов периода: %w", err)
	}
	defer rowsOrders.Close()

	var orderIDs []int
	for rowsOrders.Next() {
		var oid int
		if err := rowsOrders.Scan(&oid); err != nil {
			return 0, err
		}
		orderIDs = append(orderIDs, oid)
	}

	if len(orderIDs) == 0 {
		return 0, nil
	}

	var totalBase float64
	var currentBase float64

	switch method {
	case "labor":
		for _, oid := range orderIDs {
			var hours float64
			err := db.QueryRow("SELECT COALESCE(SUM(hours), 0) FROM labor WHERE order_id = $1", oid).Scan(&hours)
			if err != nil {
				return 0, err
			}
			totalBase += hours
			if oid == orderID {
				currentBase = hours
			}
		}
	case "bom":
		for _, oid := range orderIDs {
			var matCost float64
			err := db.QueryRow("SELECT COALESCE(SUM(quantity * unit_cost), 0) FROM boms WHERE order_id = $1", oid).Scan(&matCost)
			if err != nil {
				return 0, err
			}
			totalBase += matCost
			if oid == orderID {
				currentBase = matCost
			}
		}
	default:
		return 0, fmt.Errorf("Неизвестный метод")
	}

	if totalBase == 0 {
		return 0, nil
	}

	share := currentBase / totalBase
	return totalOverhead * share, nil
}
