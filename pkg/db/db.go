// Пакет для работы с БД
package db

import (
	"database/sql"
	"fmt"
	"os"
)

// OpenViaEnvVars - открытие БД через переменные окружения.
// Возвращает БД и ошибку.
func OpenViaEnvVars(driver string) (*sql.DB, error) {
	return OpenViaDsn(getDsnFromEnv(), driver)
}

// OpenViaDsn - открытие БД через строку DSN.
// Принимает строку DSN.
// Возвращает БД и ошибку.
func OpenViaDsn(dsn string, driver string) (*sql.DB, error) {
	if dsn == "" {
		dsn = getDsnFromEnv()
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// getDsnFromEnv - получение строки DSN из переменных окружения.
func getDsnFromEnv() string {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))

	fmt.Println(dsn)

	return dsn
}
