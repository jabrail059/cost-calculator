package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gitverse.ru/topit/12-40_team20_Zueva/internal/models"
)

func ParseBOM(path string) ([]models.BOMItem, error) {
	fileName := filepath.Base(path)
	file, err := os.Open(path)
	if err != nil {
		return nil, &models.CSVError{
			FileName: fileName,
			Row:      0,
			Column:   "",
			Cause:    "Не удалось открыть файл",
		}
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'

	var items []models.BOMItem

	i := -1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			return items, nil
		}
		if err != nil {
			return nil, &models.CSVError{
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
			return nil, &models.CSVError{
				FileName: fileName,
				Row:      i + 1,
				Column:   "",
				Cause:    "Неккоректное число полей!",
			}
		}
		order_id, err := strconv.Atoi(record[0])
		if err != nil || order_id <= 0 {
			return nil, &models.CSVError{
				FileName: fileName,
				Row:      i + 1,
				Column:   "OrderId",
				Cause:    "Некорректное OrderId",
			}
		}
		quantity, err := strconv.ParseFloat(record[1], 64)
		if err != nil || quantity <= 0 {
			return nil, &models.CSVError{
				FileName: fileName,
				Row:      i + 1,
				Column:   "Quantity",
				Cause:    "Некорректное Quantity",
			}
		}
		unitCost, err := strconv.ParseFloat(record[2], 64)
		if err != nil || unitCost <= 0 {
			return nil, &models.CSVError{
				FileName: fileName,
				Row:      i + 1,
				Column:   "UnitCost",
				Cause:    "Некорректное UnitCost",
			}
		}
		item := models.BOMItem{
			OrderID:      order_id,
			Quantity:     quantity,
			UnitCost:     unitCost,
			MaterialCode: record[3],
		}
		items = append(items, item)
	}
}

func ParseOverhead(path string) ([]models.OverheadItem, error) {
	fileName := filepath.Base(path)
	file, err := os.Open(path)
	if err != nil {
		return nil, &models.CSVError{
			FileName: fileName,
			Row:      0,
			Column:   "",
			Cause:    "Не удалось открыть файл",
		}
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'

	var items []models.OverheadItem
	i := -1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			return items, nil
		}
		if err != nil {
			return nil, &models.CSVError{
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
		if len(record) != 3 {
			return nil, &models.CSVError{
				FileName: fileName,
				Row:      i + 1,
				Column:   "",
				Cause:    "Неккоректное число полей!",
			}
		}

		err = ValidateDate(record[0])
		if err != nil {
			return nil, &models.CSVError{
				FileName: fileName,
				Row:      i + 1,
				Column:   "Date",
				Cause:    err.Error(),
			}
		}

		amount, err := strconv.ParseFloat(record[2], 64)
		if err != nil || amount <= 0 {
			return nil, &models.CSVError{
				FileName: fileName,
				Row:      i + 1,
				Column:   "Amount",
				Cause:    "Некорректное Amount",
			}
		}
		item := models.OverheadItem{
			Date:     record[0],
			ProdType: record[1],
			Amount:   amount,
		}
		items = append(items, item)
	}
}

func ParseLabor(path string) ([]models.LaborItem, error) {
	fileName := filepath.Base(path)
	file, err := os.Open(path)
	if err != nil {
		return nil, &models.CSVError{
			FileName: fileName,
			Row:      0,
			Column:   "",
			Cause:    "Не удалось открыть файл",
		}
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'

	var items []models.LaborItem
	i := -1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			return items, nil
		}
		if err != nil {
			return nil, &models.CSVError{
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
		if len(record) != 3 {
			return nil, &models.CSVError{
				FileName: fileName,
				Row:      i + 1,
				Column:   "",
				Cause:    "Неккоректное число полей!",
			}
		}
		order_id, err := strconv.Atoi(record[0])
		if err != nil || order_id <= 0 {
			return nil, &models.CSVError{
				FileName: fileName,
				Row:      i + 1,
				Column:   "OrderId",
				Cause:    "Некорректное OrderId",
			}
		}
		rate, err := strconv.ParseFloat(record[1], 64)
		if err != nil || rate <= 0 {
			return nil, &models.CSVError{
				FileName: fileName,
				Row:      i + 1,
				Column:   "Rate",
				Cause:    "Некорректное Rate",
			}
		}
		hours, err := strconv.ParseFloat(record[2], 64)
		if err != nil || hours <= 0 {
			return nil, &models.CSVError{
				FileName: fileName,
				Row:      i + 1,
				Column:   "Hours",
				Cause:    "Некорректное Hours",
			}
		}
		item := models.LaborItem{
			OrderID: order_id,
			Rate:    rate,
			Hours:   hours,
		}
		items = append(items, item)
	}
}

func ValidateDate(DateStr string) error {
	_, err := time.Parse("2006-01-02", DateStr)
	if err != nil {
		return fmt.Errorf("Некорректная дата: %w", err)
	}
	return nil
}
