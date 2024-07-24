package database

import (
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func ConnectToPostgres(connectionStr string) (*sqlx.DB, error) {
	// dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", username, password, host, port, dbname)
	// dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s", host, port, dbname, username, password, "disable")
	// postgres://test:test@localhost:57581/auth_db?sslmode=disable
	conn, err := sqlx.Open("pgx", connectionStr)
	if err != nil {
		return nil, err
	}

	err = conn.Ping()
	if err != nil {
		return nil, err
	}

	return conn, nil
}
