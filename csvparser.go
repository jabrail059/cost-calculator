package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

func ParseBOM(path string) ([]BOMItem, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'

	var items []BOMItem

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	for i, record := range records {
		if i == 0 {
			continue
		}
		if len(record) < 4 {
			return nil, fmt.Errorf("Строка %d: недостаточно полей!", i+1)
		}
		orderID, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, fmt.Errorf("Строка %d: некорректное OrderID", i+1)
		}
		quantity, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			return nil, fmt.Errorf("Строка %d: некорректное Quantity", i+1)
		}
		unitCost, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			return nil, fmt.Errorf("Строка %d: некорректное UnitCost", i+1)
		}
		item := BOMItem{
			OrderID:      orderID,
			Quantity:     quantity,
			UnitCost:     unitCost,
			MaterialCode: record[1],
		}
		items = append(items, item)
	}
	return items, nil
}

func ParseOverhead(path string) ([]OverheadItem, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var items []OverheadItem
	for i, record := range records {
		if i == 0 {
			continue
		}
		orderID, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, fmt.Errorf("Строка %d: некорректное OrderID", i+1)
		}
		amount, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			return nil, fmt.Errorf("Строка %d: некорректное Amount", i+1)
		}
		item := OverheadItem{
			OrderID:  orderID,
			ProdType: record[1],
			Amount:   amount,
		}
		items = append(items, item)
	}
	return items, nil
}

func ParseLabor(path string) ([]LaborItem, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var items []LaborItem
	for i, record := range records {
		if i == 0 {
			continue
		}
		orderID, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, fmt.Errorf("Строка %d: некорректное OrderId", i+1)
		}
		rate, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			return nil, fmt.Errorf("Строка %d: некорректное Rate", i+1)
		}
		hours, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			return nil, fmt.Errorf("Строка %d: некорректное Hours", i+1)
		}
		item := LaborItem{
			OrderID: orderID,
			Rate:    rate,
			Hours:   hours,
		}
		items = append(items, item)
	}
	return items, nil
}
