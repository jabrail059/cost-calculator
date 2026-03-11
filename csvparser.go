package main

import (
	"encoding/csv"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

func ParseBOM(path string) ([]BOMItem, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, &CSVError{
			FileName: filepath.Base(path),
			Row:      0,
			Column:   "",
			Cause:    "Не удалось открыть файл",
		}
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'

	var items []BOMItem

	i := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			return items, nil
		}
		if i == 0 {
			continue
		}
		i++
		if err != nil {
			return nil, &CSVError{
				FileName: filepath.Base(path),
				Row:      0,
				Column:   "",
				Cause:    "Не удалось считать данные из файла",
			}
		}

		if len(record) < 4 {
			return nil, &CSVError{
				FileName: filepath.Base(path),
				Row:      i + 1,
				Column:   "",
				Cause:    "Недостаточно полей!",
			}
		}
		orderID, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, &CSVError{
				FileName: filepath.Base(path),
				Row:      i + 1,
				Column:   "OrderId",
				Cause:    "некорректное OrderID",
			}
		}
		quantity, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			return nil, &CSVError{
				FileName: filepath.Base(path),
				Row:      i + 1,
				Column:   "Quantity",
				Cause:    "Некорректное Quantity",
			}
		}
		unitCost, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			return nil, &CSVError{
				FileName: filepath.Base(path),
				Row:      i + 1,
				Column:   "UnitCost",
				Cause:    "Некорректное UnitCost",
			}
		}
		item := BOMItem{
			OrderID:      orderID,
			Quantity:     quantity,
			UnitCost:     unitCost,
			MaterialCode: record[3],
		}
		items = append(items, item)
	}
}

func ParseOverhead(path string) ([]OverheadItem, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, &CSVError{
			FileName: filepath.Base(path),
			Row:      0,
			Column:   "",
			Cause:    "Не удалось открыть файл",
		}
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'

	var items []OverheadItem
	i := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			return items, nil
		}
		if i == 0 {
			continue
		}
		i++
		if err != nil {
			return nil, &CSVError{
				FileName: filepath.Base(path),
				Row:      0,
				Column:   "",
				Cause:    "Не удалось считать данные из файла",
			}
		}
		if len(record) < 4 {
			return nil, &CSVError{
				FileName: filepath.Base(path),
				Row:      i + 1,
				Column:   "",
				Cause:    "Недостаточно полей!",
			}
		}
		orderID, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, &CSVError{
				FileName: filepath.Base(path),
				Row:      i + 1,
				Column:   "OrderId",
				Cause:    "Некорректное OrderID",
			}
		}
		amount, err := strconv.ParseFloat(record[3], 64)
		if err != nil {
			return nil, &CSVError{
				FileName: filepath.Base(path),
				Row:      i + 1,
				Column:   "Amount",
				Cause:    "Некорректное Amount",
			}
		}
		item := OverheadItem{
			OrderID:  orderID,
			Date:     record[1],
			ProdType: record[2],
			Amount:   amount,
		}
		items = append(items, item)
	}
}

func ParseLabor(path string) ([]LaborItem, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, &CSVError{
			FileName: filepath.Base(path),
			Row:      0,
			Column:   "",
			Cause:    "Не удалось открыть файл",
		}
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'

	var items []LaborItem
	i := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			return items, nil
		}
		if i == 0 {
			continue
		}
		i++
		if err != nil {
			return nil, &CSVError{
				FileName: filepath.Base(path),
				Row:      0,
				Column:   "",
				Cause:    "Не удалось считать данные из файла",
			}
		}
		if len(record) < 4 {
			return nil, &CSVError{
				FileName: filepath.Base(path),
				Row:      i + 1,
				Column:   "",
				Cause:    "Недостаточно полей!",
			}
		}
		orderID, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, &CSVError{
				FileName: filepath.Base(path),
				Row:      i + 1,
				Column:   "Orderid",
				Cause:    "Некорректное OrderId",
			}
		}
		rate, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			return nil, &CSVError{
				FileName: filepath.Base(path),
				Row:      i + 1,
				Column:   "Rate",
				Cause:    "Некорректное Rate",
			}
		}
		hours, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			return nil, &CSVError{
				FileName: filepath.Base(path),
				Row:      i + 1,
				Column:   "Hours",
				Cause:    "Некорректное Hours",
			}
		}
		item := LaborItem{
			OrderID: orderID,
			Rate:    rate,
			Hours:   hours,
		}
		items = append(items, item)
	}
}
