# GoControl - Herramienta de Command & Control (C2)

## Descripción

GoControl es un proyecto de Trabajo de Fin de Grado (TFG) que ofrece una herramienta de Command & Control (C2) personalizada y menos detectable para equipos de Red Team. Desarrollado como parte de un proyecto académico, GoControl tiene como objetivo proporcionar una solución eficaz y adaptable a las necesidades específicas de cada Red Team, evitando la detección por parte de sistemas de seguridad tradicionales como antivirus (AV) y soluciones de detección y respuesta de endpoints (EDR).

## Características principales

- Personalización y adaptabilidad a las necesidades específicas de cada Red Team
- Utiliza el protocolo QUIC para mejorar el rendimiento y la seguridad en las comunicaciones
- señado para evadir la detección de sistemas antivirus (AV) y de detección y respuesta de endpoints (EDR)
- Iterfaz de usuario intuitiva y fácil de usar
- Facilita la evaluación de la resistencia de un sistema frente a posibles ataques

## Requisitos

Golang 1.17 o superior
Paquetes de Golang para el manejo de QUIC y otros protocolos necesarios
libcurl en C para la integración con bibliotecas de red

## Instalación

Clonar el repositorio de GoControl:

```bash

git clone https://github.com/user/GoControl.git
```
Entrar en el directorio clonado e instalar las dependencias necesarias:

```arduino

cd GoControl
go get
```
Compilar y generar el ejecutable:

```go

go build -o gocontrol main.go
```
Ejecutar GoControl:

```bash

./gocontrol
```

## Uso

Para comenzar a utilizar GoControl, siga los siguientes pasos:

1. Ejecute GoControl utilizando el siguiente comando:

```go

./gocontrol
```
2. Abra un navegador web y navegue hasta localhost:8080 para acceder a la interfaz de usuario web de GoControl.

La interfaz de usuario web consta de tres menús principales: Dashboard, Listeners y Agents.

### Dashboard

En el panel de control, encontrará gráficos e información sobre los agentes y listeners actuales. Esta vista le proporciona una visión general del estado y la actividad de su entorno de C2.

### Listeners

En la sección Listeners, podrá ver todos los listeners existentes y crear nuevos listeners, tanto para HTTP como para QUIC. Para crear un nuevo listener, simplemente haga clic en el botón "Nuevo listener" y complete la información requerida.

### Agents

La sección Agents le permite crear agentes especificando el listener al que deben conectarse. Una vez que una máquina esté infectada con un agente, podrá ver y gestionar sus actividades a través de esta sección.

### Control de agentes y ejecución de comandos

Para controlar un agente, haga clic en el agente o en el listener correspondiente para acceder a su controlador. Una vez dentro, seleccione el agente deseado y utilice la terminal ubicada en la parte inferior de la pantalla para ejecutar comandos.
