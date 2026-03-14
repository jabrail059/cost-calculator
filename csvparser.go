package main

import (
	"encoding/csv"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

func ParseBOM(path string) ([]BOMItem, error) {
	fileName := filepath.Base(path)
	file, err := os.Open(path)
	if err != nil {
		return nil, &CSVError{
			FileName: fileName,
			Row:      0,
			Column:   "",
			Cause:    "Не удалось открыть файл",
		}
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'

	var items []BOMItem

	i := -1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			return items, nil
		}
		if err != nil {
			return nil, &CSVError{
				FileName: fileName,
				Row:      0,
				Column:   "",
				Cause:    "Не удалось считать данные из файла",
			}
		}
		i++
		if i == 0 {
			continue
		}

		if len(record) != 4 {
			return nil, &CSVError{
				FileName: fileName,
				Row:      i + 1,
				Column:   "",
				Cause:    "Неккоректное число полей!",
			}
		}
		order_id, err := strconv.Atoi(record[0])
		if err != nil || order_id <= 0 {
			return nil, &CSVError{
				FileName: fileName,
				Row:      i + 1,
				Column:   "OrderId",
				Cause:    "Некорректное OrderId",
			}
		}
		quantity, err := strconv.ParseFloat(record[1], 64)
		if err != nil || quantity <= 0 {
			return nil, &CSVError{
				FileName: fileName,
				Row:      i + 1,
				Column:   "Quantity",
				Cause:    "Некорректное Quantity",
			}
		}
		unitCost, err := strconv.ParseFloat(record[2], 64)
		if err != nil || unitCost <= 0 {
			return nil, &CSVError{
				FileName: fileName,
				Row:      i + 1,
				Column:   "UnitCost",
				Cause:    "Некорректное UnitCost",
			}
		}
		item := BOMItem{
			OrderID:      order_id,
			Quantity:     quantity,
			UnitCost:     unitCost,
			MaterialCode: record[3],
		}
		items = append(items, item)
	}
}

func ParseOverhead(path string) ([]OverheadItem, error) {
	fileName := filepath.Base(path)
	file, err := os.Open(path)
	if err != nil {
		return nil, &CSVError{
			FileName: fileName,
			Row:      0,
			Column:   "",
			Cause:    "Не удалось открыть файл",
		}
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'

	var items []OverheadItem
	i := -1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			return items, nil
		}
		if err != nil {
			return nil, &CSVError{
				FileName: fileName,
				Row:      0,
				Column:   "",
				Cause:    "Не удалось считать данные из файла",
			}
		}
		i++
		if i == 0 {
			continue
		}
		if len(record) != 4 {
			return nil, &CSVError{
				FileName: fileName,
				Row:      i + 1,
				Column:   "",
				Cause:    "Неккоректное число полей!",
			}
		}
		order_id, err := strconv.Atoi(record[0])
		if err != nil || order_id <= 0 {
			return nil, &CSVError{
				FileName: fileName,
				Row:      i + 1,
				Column:   "OrderId",
				Cause:    "Некорректное OrderId",
			}
		}
		amount, err := strconv.ParseFloat(record[3], 64)
		if err != nil || amount <= 0 {
			return nil, &CSVError{
				FileName: fileName,
				Row:      i + 1,
				Column:   "Amount",
				Cause:    "Некорректное Amount",
			}
		}
		item := OverheadItem{
			OrderID:  order_id,
			Date:     record[1],
			ProdType: record[2],
			Amount:   amount,
		}
		items = append(items, item)
	}
}

func ParseLabor(path string) ([]LaborItem, error) {
	fileName := filepath.Base(path)
	file, err := os.Open(path)
	if err != nil {
		return nil, &CSVError{
			FileName: fileName,
			Row:      0,
			Column:   "",
			Cause:    "Не удалось открыть файл",
		}
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'

	var items []LaborItem
	i := -1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			return items, nil
		}
		if err != nil {
			return nil, &CSVError{
				FileName: fileName,
				Row:      0,
				Column:   "",
				Cause:    "Не удалось считать данные из файла",
			}
		}
		i++
		if i == 0 {
			continue
		}
		if len(record) != 4 {
			return nil, &CSVError{
				FileName: fileName,
				Row:      i + 1,
				Column:   "",
				Cause:    "Неккоректное число полей!",
			}
		}
		order_id, err := strconv.Atoi(record[0])
		if err != nil || order_id <= 0 {
			return nil, &CSVError{
				FileName: fileName,
				Row:      i + 1,
				Column:   "OrderId",
				Cause:    "Некорректное OrderId",
			}
		}
		rate, err := strconv.ParseFloat(record[1], 64)
		if err != nil || rate <= 0 {
			return nil, &CSVError{
				FileName: fileName,
				Row:      i + 1,
				Column:   "Rate",
				Cause:    "Некорректное Rate",
			}
		}
		hours, err := strconv.ParseFloat(record[2], 64)
		if err != nil || hours <= 0 {
			return nil, &CSVError{
				FileName: fileName,
				Row:      i + 1,
				Column:   "Hours",
				Cause:    "Некорректное Hours",
			}
		}
		item := LaborItem{
			OrderID: order_id,
			Rate:    rate,
			Hours:   hours,
		}
		items = append(items, item)
	}
}
