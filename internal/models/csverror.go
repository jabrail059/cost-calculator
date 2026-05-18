package models

import "fmt"

type CSVError struct {
	FileName string
	Row      int
	Column   string
	Cause    string
}

func (err *CSVError) Error() string {
	return fmt.Sprintf("Файл: %s, строка: %d, столбец: %s, ошибка: %s", err.FileName, err.Row, err.Column, err.Cause)
}
