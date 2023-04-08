package models

import "database/sql"

// Estructura de datos para el modelo de Beat donde tiene "key" del agente y "time" de la hora de la ultima conexion
type Beat struct {
	Key  string
	Time string
}

// Crear tabla de beats en la base de datos
func CreateBeatTable(db *sql.DB) error {

	// Crear la tabla si no existe
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS beats (key TEXT PRIMARY KEY, time TEXT)")
	if err != nil {
		return err
	}

	return nil
}

// Insertar un beat en la base de datos
func InsertBeat(db *sql.DB, beat Beat) error {
	// Insertar beat en la base de datos
	_, err := db.Exec("INSERT INTO beats (key, time) VALUES (?, ?)", beat.Key, beat.Time)
	if err != nil {
		return err
	}

	return nil
}

// GetBeatsByAgentKey obtiene los beats de un agente de la base de datos por su key
func GetBeatsByAgentKey(db *sql.DB, key string) ([]Beat, error) {
	// Obtener los beats de un agente
	rows, err := db.Query(` SELECT key, time FROM beats WHERE key = ?`, key)
	if err != nil {
		return nil, err
	}

	// Crear un slice de beats
	var beats []Beat

	// Iterar sobre los resultados
	for rows.Next() {
		var beat Beat
		err = rows.Scan(&beat.Key, &beat.Time)
		if err != nil {
			return nil, err
		}
		beats = append(beats, beat)
	}

	return beats, nil
}

// GetBeatsLast24Hours obtiene los beats de los ultimos 24 horas de la base de datos

func GetBeatsLast24Hours(db *sql.DB) ([]Beat, error) {
	// Obtener los beats de los ultimos 24 horas
	rows, err := db.Query(` SELECT key, time FROM beats WHERE time > datetime('now', '-1 day')`)
	if err != nil {
		return nil, err
	}

	// Crear un slice de beats
	var beats []Beat

	// Iterar sobre los resultados
	for rows.Next() {
		var beat Beat
		err = rows.Scan(&beat.Key, &beat.Time)
		if err != nil {
			return nil, err
		}
		beats = append(beats, beat)
	}

	return beats, nil
}
