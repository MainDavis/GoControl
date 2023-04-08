package models

import (
	"database/sql"
	"fmt"
)

// Estructura de datos para el modelo de Agent
type Agent struct {
	UUID         string
	Hostname     string
	Username     string
	OS           string
	Architecture string
	LocalIP      string
	Type         string
	IsRoot       bool
	ListenerUUID string
}

// Crear tabla de agentes en la base de datos
func CreateAgentTable(db *sql.DB) error {

	// Crear la tabla si no existe
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS agents (uuid TEXT PRIMARY KEY, hostname TEXT, username TEXT, os TEXT, architecture TEXT, LocalIP TEXT, type TEXT, IsRoot INTEGER, ListenerUUID TEXT)")
	if err != nil {
		return err
	}

	return nil
}

// Insertar un agente en la base de datos
func InsertAgent(db *sql.DB, agent Agent) error {

	// Insertar agente en la base de datos
	_, err := db.Exec("INSERT INTO agents (uuid, hostname, username, os, architecture, LocalIP, type, IsRoot, ListenerUUID) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", agent.UUID, agent.Hostname, agent.Username, agent.OS, agent.Architecture, agent.LocalIP, agent.Type, agent.IsRoot, agent.ListenerUUID)
	if err != nil {
		return err
	}

	// Notifico que se ha insertado un agente
	fmt.Printf("Se ha insertado el agente %s en la base de datos", agent.UUID)

	return nil

}

// GetAgent obtiene un agente de la base de datos por su key
func GetAgentByKey(db *sql.DB, key string) (*Agent, error) {

	row := db.QueryRow(`
		SELECT UUID, Hostname, Username, OS, Architecture, LocalIP, Type, IsRoot, ListenerUUID
		FROM agents
		WHERE Key = ?
	`, key)

	var agent Agent

	if err := row.Scan(&agent.UUID, &agent.Hostname, &agent.Username, &agent.OS, &agent.Architecture, &agent.LocalIP, &agent.Type, &agent.IsRoot, &agent.ListenerUUID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &agent, nil
}

// GetAgentsByListener obtiene una lista de agentes de la base de datos por su listener
func GetAgentsByListener(db *sql.DB, listenerUUID string) ([]Agent, error) {

	rows, err := db.Query(`
		SELECT UUID, Hostname, Username, OS, Architecture, LocalIP, Type, IsRoot, ListenerUUID
		FROM agents
		WHERE ListenerUUID = ?
	`, listenerUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []Agent
	for rows.Next() {
		var agent Agent

		if err := rows.Scan(&agent.UUID, &agent.Hostname, &agent.Username, &agent.OS, &agent.Architecture, &agent.LocalIP, &agent.Type, &agent.IsRoot, &agent.ListenerUUID); err != nil {
			return nil, err
		}

		agents = append(agents, agent)
	}

	return agents, nil
}

// GetAgents obtiene una lista de agentes de la base de datos
func GetAgents(db *sql.DB) ([]Agent, error) {

	rows, err := db.Query(`SELECT UUID, Hostname, Username, OS, Architecture, LocalIP, Type, IsRoot, ListenerUUID FROM agents`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []Agent
	for rows.Next() {
		var agent Agent

		if err := rows.Scan(&agent.UUID, &agent.Hostname, &agent.Username, &agent.OS, &agent.Architecture, &agent.LocalIP, &agent.Type, &agent.IsRoot, &agent.ListenerUUID); err != nil {
			return nil, err
		}

		agents = append(agents, agent)
	}

	return agents, nil
}

// UpdateAgent actualiza un agente en la base de datos
func UpdateAgent(db *sql.DB, agent Agent) error {

	// Actualizar agente en la base de datos
	_, err := db.Exec("UPDATE agents SET UUID = ?, hostname = ?, username = ?, os = ?, architecture = ?, LocalIP = ?, type = ?, IsRoot = ?, ListenerUUID = ? WHERE key = ?", agent.UUID, agent.Hostname, agent.Username, agent.OS, agent.Architecture, agent.LocalIP, agent.Type, agent.IsRoot, agent.ListenerUUID)
	if err != nil {
		return err
	}

	// Notifico que se ha actualizado un agente
	fmt.Printf("Se ha actualizado el agente %s en la base de datos", agent.UUID)

	return nil
}

// DeleteAgent elimina un agente de la base de datos
func DeleteAgent(db *sql.DB, UUID string) error {

	// Eliminar agee de la base de datos
	_, err := db.Exec("DELETE FROM agents WHERE UUID = ?", UUID)
	if err != nil {
		return err
	}

	// Notifico que se ha eliminado un agente
	fmt.Printf("Se ha eliminado el agente %s de la base de datos", UUID)

	return nil
}

func AgentExists(db *sql.DB, UUID string) (bool, error) {

	// Comprobar si el agente existe
	row := db.QueryRow("SELECT EXISTS(SELECT 1 FROM agents WHERE UUID = ?)", UUID)

	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}
