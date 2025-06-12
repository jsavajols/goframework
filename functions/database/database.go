package database

import (
	"database/sql"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jsavajols/goframework/functions/logs"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// ConnectDatabase connecte la base de données
func ConnectDatabase(database string, dialect ...string) (*sql.DB, string, error) {
	// Force l'utilisation de la database passée en paramètre si elle est définie
	if database == "" {
		if os.Getenv("DB_NAME") != "" {
			database = os.Getenv("DB_NAME")
		}
	}
	if dialect[0] == "" {
		if os.Getenv("DB_DIALECT") != "" {
			dialect[0] = os.Getenv("DB_DIALECT")
		}
	}
	logs.Logs("Type de base de données : ", dialect[0])
	var db *sql.DB
	var err error
	if dialect[0] == "mysql" || dialect[0] == "" {
		db, err = sql.Open("mysql", os.Getenv("DB_USER")+":"+os.Getenv("DB_PASSWORD")+"@tcp("+os.Getenv("DB_HOST")+":"+os.Getenv("DB_PORT")+")/"+database)
	}
	if dialect[0] == "postgres" {
		db, err = sql.Open("postgres", "host="+os.Getenv("DB_HOST")+" port="+os.Getenv("DB_PORT")+" user="+os.Getenv("DB_USER")+" password="+os.Getenv("DB_PASSWORD")+" dbname="+database+" sslmode=disable")
	}
	if dialect[0] == "sqlite3" {
		db, err = sql.Open("sqlite3", database)
	}
	if err != nil {
		return nil, dialect[0], err
	}
	return db, dialect[0], nil
}
