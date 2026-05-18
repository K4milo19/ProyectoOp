package main

// red_servidor.go
// Servidor TCP de ShellOS.
// Escucha conexiones entrantes de clientes, recibe comandos en texto plano,
// los ejecuta localmente con ejecutarComando() y devuelve la salida.
// Cada comando y su resultado quedan registrados en logs.txt.

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

// ── Constantes ────────────────────────────────────────────────────────────────

const (
	puertoServidor = ":9000"   // puerto TCP en el que escucha el servidor
	archivoLog     = "logs.txt" // archivo donde se registran los comandos
)

// ── Log ───────────────────────────────────────────────────────────────────────

var logMu sync.Mutex // protege las escrituras concurrentes al log

// escribirLog añade una entrada al archivo logs.txt de forma segura entre goroutines.
func escribirLog(cliente, comando, salida string) {
	logMu.Lock()
	defer logMu.Unlock()

	f, err := os.OpenFile(archivoLog, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	marca := time.Now().Format("2006-01-02 15:04:05")
	entrada := fmt.Sprintf(
		"[%s] CLIENTE=%s\n  CMD : %s\n  OUT : %s\n%s\n",
		marca, cliente, comando,
		strings.ReplaceAll(strings.TrimSpace(salida), "\n", "\n       "),
		strings.Repeat("─", 60),
	)
	f.WriteString(entrada)
}

// ── Manejo de cliente ─────────────────────────────────────────────────────────

// manejarCliente atiende una conexión individual en su propio goroutine.
// Lee líneas de texto (un comando por línea), ejecuta cada uno y responde
// con la salida seguida del marcador "<<<END>>>".
func manejarCliente(conn net.Conn, onLog func(string)) {
	defer conn.Close()

	remoto := conn.RemoteAddr().String()
	onLog(fmt.Sprintf("[ CONEXIÓN ]  %s conectado", remoto))

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		linea := strings.TrimSpace(scanner.Text())
		if linea == "" {
			continue
		}

		// El cliente envía "bye" o "exit" para cerrar su sesión
		if linea == "bye" || linea == "exit" {
			fmt.Fprintf(conn, "[ sesión cerrada por el cliente ]\n<<<END>>>\n")
			onLog(fmt.Sprintf("[ DESCONEXIÓN ]  %s cerró la sesión", remoto))
			return
		}

		salida, _ := ejecutarComando(linea)
		if salida == "" {
			salida = "[ sin salida ]"
		}

		// Enviar salida al cliente; el marcador indica fin de respuesta
		fmt.Fprintf(conn, "%s\n<<<END>>>\n", salida)

		escribirLog(remoto, linea, salida)
		onLog(fmt.Sprintf("[ CMD ]  %s  →  %s", remoto, linea))
	}

	if err := scanner.Err(); err != nil {
		onLog(fmt.Sprintf("[ ERROR ]  %s: %v", remoto, err))
	} else {
		onLog(fmt.Sprintf("[ DESCONEXIÓN ]  %s", remoto))
	}
}

// ── Servidor ──────────────────────────────────────────────────────────────────

// iniciarServidor abre el socket TCP y acepta clientes indefinidamente.
// onLog es un callback que se llama con cada evento para mostrarlo en la GUI.
// stop es un canal que, al cerrarse, detiene el listener.
func iniciarServidor(onLog func(string), stop <-chan struct{}) error {
	ln, err := net.Listen("tcp", puertoServidor)
	if err != nil {
		return fmt.Errorf("no se pudo iniciar el servidor en %s: %w", puertoServidor, err)
	}

	onLog(fmt.Sprintf("◈  Servidor escuchando en %s", puertoServidor))
	onLog(fmt.Sprintf("◈  Log guardado en: %s", archivoLog))

	// Goroutine que cierra el listener cuando se señaliza stop
	go func() {
		<-stop
		ln.Close()
	}()

	// Bucle de aceptación de conexiones
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				// Error normal al cerrar el listener con stop
				select {
				case <-stop:
					return
				default:
					onLog(fmt.Sprintf("[ ERROR accept ]  %v", err))
					continue
				}
			}
			go manejarCliente(conn, onLog)
		}
	}()

	return nil
}