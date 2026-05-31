package storage

import (
	"encoding/json"

	"gitverse.ru/topit/12-40_team20_Zueva/internal/models"
)

func CreateReportCalculation(createdBy int, data models.CalculationResult) (int, error) {
	data.CalculationID = 0

	rawData, err := json.Marshal(data)
	if err != nil {
		return 0, err
	}

	var id int
	err = db.QueryRow(
		"INSERT INTO report_calculations(created_by, calculation_data) VALUES ($1, $2) RETURNING id",
		createdBy,
		rawData,
	).Scan(&id)
	return id, err
}

func GetReportCalculation(calculationID int, userID int) (*models.CalculationResult, error) {
	var rawData []byte
	err := db.QueryRow(
		"SELECT calculation_data FROM report_calculations WHERE id = $1 AND created_by = $2",
		calculationID,
		userID,
	).Scan(&rawData)
	if err != nil {
		return nil, err
	}

	var data models.CalculationResult
	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, err
	}
	data.CalculationID = calculationID
	return &data, nil
}

func GetReportCalculationByID(calculationID int) (*models.CalculationResult, error) {
	var rawData []byte
	err := db.QueryRow(
		"SELECT calculation_data FROM report_calculations WHERE id = $1",
		calculationID,
	).Scan(&rawData)
	if err != nil {
		return nil, err
	}

	var data models.CalculationResult
	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, err
	}
	data.CalculationID = calculationID
	return &data, nil
}

func CreateReport(calculationID int, createdBy int, data models.ReportRequest) (int, error) {
	rawData, err := json.Marshal(data)
	if err != nil {
		return 0, err
	}

	var id int
	err = db.QueryRow(
		"INSERT INTO reports(calculation_id, created_by, onec_report_data) VALUES ($1, $2, $3) RETURNING id",
		calculationID,
		createdBy,
		rawData,
	).Scan(&id)
	return id, err
}

func GetReportByID(reportID int, userID int) (*models.Report, error) {
	var (
		report  models.Report
		rawData []byte
	)

	err := db.QueryRow(`
		SELECT id, calculation_id, created_by, onec_report_data, created_at
		FROM reports
		WHERE id = $1 AND created_by = $2
	`, reportID, userID).Scan(
		&report.ID,
		&report.CalculationID,
		&report.CreatedBy,
		&rawData,
		&report.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(rawData, &report.OneCReportData); err != nil {
		return nil, err
	}
	return &report, nil
}

func GetReportByIDPublic(reportID int) (*models.Report, error) {
	var (
		report  models.Report
		rawData []byte
	)

	err := db.QueryRow(`
		SELECT id, calculation_id, created_by, onec_report_data, created_at
		FROM reports
		WHERE id = $1
	`, reportID).Scan(
		&report.ID,
		&report.CalculationID,
		&report.CreatedBy,
		&rawData,
		&report.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(rawData, &report.OneCReportData); err != nil {
		return nil, err
	}
	return &report, nil
}
