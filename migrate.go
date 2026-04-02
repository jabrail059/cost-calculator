package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP DEFAULT NOW()
		)
	`)
	return err
}

func runMigrations(db *sql.DB) error {
	if err := ensureMigrationsTable(db); err != nil {
		return fmt.Errorf("ошибка создания таблицы migrations: %w", err)
	}

	files, err := os.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("ошибка чтения папки migrations: %w", err)
	}

	var upFiles []string
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := f.Name()
		// ищем файлы вида *.up.sql
		if len(name) > 6 && name[len(name)-6:] == ".up.sql" {
			upFiles = append(upFiles, name)
		}
	}
	sort.Slice(upFiles, func(i, j int) bool { return upFiles[i] < upFiles[j] })

	for _, fileName := range upFiles {
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM migrations WHERE name = $1", fileName).Scan(&count)
		if err != nil {
			return err
		}
		if count > 0 {
			continue // уже применена
		}

		content, err := os.ReadFile(filepath.Join("migrations", fileName))
		if err != nil {
			return fmt.Errorf("ошибка чтения файла %s: %w", fileName, err)
		}

		tx, err := db.Begin()
		if err != nil {
			return err
		}

		_, err = tx.Exec(string(content))
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("ошибка выполнения миграции %s: %w", fileName, err)
		}

		_, err = tx.Exec("INSERT INTO migrations (name) VALUES ($1)", fileName)
		if err != nil {
			tx.Rollback()
			return err
		}

		if err = tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
