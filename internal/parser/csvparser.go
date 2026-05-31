package parser

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gitverse.ru/topit/12-40_team20_Zueva/internal/models"
)

func ParseBOM(path string) ([]models.BOMItem, error) {
	fileName := filepath.Base(path)
	reader, file, err := csvReader(path)
	if err != nil {
		return nil, csvError(fileName, 0, "", "failed to open file")
	}
	defer file.Close()

	var items []models.BOMItem
	row := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			return items, nil
		}
		if err != nil {
			return nil, csvError(fileName, row, "", "failed to read file")
		}
		row++
		if row == 1 {
			continue
		}
		trimRecord(record)
		if len(record) != 4 {
			return nil, csvError(fileName, row, "", "invalid number of fields")
		}

		orderID, err := parsePositiveInt(record[0])
		if err != nil {
			return nil, csvError(fileName, row, "OrderId", "invalid OrderId")
		}
		quantity, err := parsePositiveFloat(record[1])
		if err != nil {
			return nil, csvError(fileName, row, "Quantity", "invalid Quantity")
		}
		unitCost, err := parsePositiveFloat(record[2])
		if err != nil {
			return nil, csvError(fileName, row, "UnitCost", "invalid UnitCost")
		}

		items = append(items, models.BOMItem{
			OrderID:      orderID,
			Quantity:     quantity,
			UnitCost:     unitCost,
			MaterialCode: record[3],
		})
	}
}

func ParseOverhead(path string) ([]models.OverheadItem, error) {
	fileName := filepath.Base(path)
	reader, file, err := csvReader(path)
	if err != nil {
		return nil, csvError(fileName, 0, "", "failed to open file")
	}
	defer file.Close()

	var items []models.OverheadItem
	row := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			return items, nil
		}
		if err != nil {
			return nil, csvError(fileName, row, "", "failed to read file")
		}
		row++
		if row == 1 {
			continue
		}
		trimRecord(record)
		if len(record) != 3 && len(record) != 4 {
			return nil, csvError(fileName, row, "", "invalid number of fields")
		}

		dateIndex := 0
		typeIndex := 1
		amountIndex := 2
		var orderID *int
		if len(record) == 4 {
			parsedOrderID, err := parsePositiveInt(record[0])
			if err != nil {
				return nil, csvError(fileName, row, "OrderId", "invalid OrderId")
			}
			orderID = &parsedOrderID
			dateIndex = 1
			typeIndex = 2
			amountIndex = 3
		}

		if err := ValidateDate(record[dateIndex]); err != nil {
			return nil, csvError(fileName, row, "Date", err.Error())
		}
		amount, err := parsePositiveFloat(record[amountIndex])
		if err != nil {
			return nil, csvError(fileName, row, "Amount", "invalid Amount")
		}

		items = append(items, models.OverheadItem{
			OrderID:  orderID,
			Date:     record[dateIndex],
			ProdType: record[typeIndex],
			Amount:   amount,
		})
	}
}

func ParseLabor(path string) ([]models.LaborItem, error) {
	fileName := filepath.Base(path)
	reader, file, err := csvReader(path)
	if err != nil {
		return nil, csvError(fileName, 0, "", "failed to open file")
	}
	defer file.Close()

	var items []models.LaborItem
	row := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			return items, nil
		}
		if err != nil {
			return nil, csvError(fileName, row, "", "failed to read file")
		}
		row++
		if row == 1 {
			continue
		}
		trimRecord(record)
		if len(record) != 3 {
			return nil, csvError(fileName, row, "", "invalid number of fields")
		}

		orderID, err := parsePositiveInt(record[0])
		if err != nil {
			return nil, csvError(fileName, row, "OrderId", "invalid OrderId")
		}
		rate, err := parsePositiveFloat(record[1])
		if err != nil {
			return nil, csvError(fileName, row, "Rate", "invalid Rate")
		}
		hours, err := parsePositiveFloat(record[2])
		if err != nil {
			return nil, csvError(fileName, row, "Hours", "invalid Hours")
		}

		items = append(items, models.LaborItem{
			OrderID: orderID,
			Rate:    rate,
			Hours:   hours,
		})
	}
}

func ValidateDate(dateStr string) error {
	_, err := time.Parse("2006-01-02", strings.TrimSpace(dateStr))
	if err != nil {
		return fmt.Errorf("invalid date: %w", err)
	}
	return nil
}

func csvReader(path string) (*csv.Reader, *os.File, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	reader := csv.NewReader(file)
	reader.Comma = ';'
	reader.FieldsPerRecord = -1
	return reader, file, nil
}

func trimRecord(record []string) {
	for i := range record {
		record[i] = strings.TrimSpace(strings.TrimPrefix(record[i], "\ufeff"))
	}
}

func parsePositiveInt(value string) (int, error) {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("invalid positive integer")
	}
	return parsed, nil
}

func parsePositiveFloat(value string) (float64, error) {
	value = strings.ReplaceAll(strings.TrimSpace(value), ",", ".")
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("invalid positive float")
	}
	return parsed, nil
}

func csvError(fileName string, row int, column string, cause string) *models.CSVError {
	return &models.CSVError{
		FileName: fileName,
		Row:      row,
		Column:   column,
		Cause:    cause,
	}
}
