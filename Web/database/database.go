package database

import (
	"database/sql"

	"gocontrol/Web/models"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func Init() {
	var err error
	db, err = sql.Open("sqlite3", "./gocontrol.db")
	if err != nil {
		panic(err)
	}

	// Crear las tablas si no existe
	models.CreateListenerTable(db)
	models.CreateAgentTable(db)
	models.CreateBeatTable(db)

}

func CloseDatabase() error {
	return db.Close()
}

func GetDatabase() *sql.DB {
	return db
}
