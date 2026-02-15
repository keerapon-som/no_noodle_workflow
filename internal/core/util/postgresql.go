package util

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func NewPostgresql(Host string, Port int, User string, Password string, DBName string, SSLMode string) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", Host, Port, User, Password, DBName, SSLMode)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}
