package database

import (
	"database/sql"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

// ConnectDatabase connecte la base de données
func ConnectDatabase(database string) (*sql.DB, error) {
	// Force l'utilisation de la database passée en paramètre si elle est définie
	if database == "" {
		if os.Getenv("DB_NAME") != "" {
			database = os.Getenv("DB_NAME")
		}
	}
	db, err := sql.Open("mysql", os.Getenv("DB_USER")+":"+os.Getenv("DB_PASSWORD")+"@tcp("+os.Getenv("DB_HOST")+":"+os.Getenv("DB_PORT")+")/"+database)
	if err != nil {
		return nil, err
	}
	return db, nil
}
