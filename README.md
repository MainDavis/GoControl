# GoControl - Herramienta de Command & Control (C2)

## Descripción

GoControl es una herramienta de Command & Control (C2) personalizada diseñada para mejorar la efectividad y disminuir la detección de acciones ofensivas llevadas a cabo por equipos de Red Team. Al adaptarse a las necesidades específicas de cada Red Team y aprovechar el protocolo QUIC, GoControl ofrece una solución más eficiente y menos detectable en comparación con las herramientas de C2 convencionales.
Características principales

    Personalización y adaptabilidad a las necesidades específicas de cada Red Team
    Utiliza el protocolo QUIC para mejorar el rendimiento y la seguridad en las comunicaciones
    Diseñado para evadir la detección de sistemas antivirus (AV) y de detección y respuesta de endpoints (EDR)
    Interfaz de usuario intuitiva y fácil de usar
    Facilita la evaluación de la resistencia de un sistema frente a posibles ataques

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
