package main

// red_servidor.go
// Servidor TCP de ShellOS.
// Escucha conexiones entrantes de clientes, recibe comandos en texto plano,
// los ejecuta localmente con ejecutarComando() y devuelve la salida.
// Cada comando y su resultado quedan registrados en logs.txt.
//
// MÉTRICAS: cada 5 segundos el servidor recopila CPU, RAM y Disco propios
// y los empuja a TODOS los clientes conectados con el prefijo <<<METRICS:...>>>
// El cliente detecta esas líneas y actualiza su panel derecho.

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
	puertoServidor = ":9000"    // puerto TCP en el que escucha el servidor
	archivoLog     = "logs.txt" // archivo donde se registran los comandos
)

// ── Registro de clientes conectados ──────────────────────────────────────────

// clientesConectados mantiene un mapa conn→struct{} de todas las conexiones
// activas para que el broadcaster de métricas pueda escribirles.
var (
	clientesMu       sync.RWMutex
	clientesActivos  = make(map[net.Conn]struct{})
)

func registrarCliente(conn net.Conn) {
	clientesMu.Lock()
	clientesActivos[conn] = struct{}{}
	clientesMu.Unlock()
}

func eliminarCliente(conn net.Conn) {
	clientesMu.Lock()
	delete(clientesActivos, conn)
	clientesMu.Unlock()
}

// broadcastMetrics escribe la línea de métricas a todos los clientes.
// Se ignoran los errores de escritura: si un cliente se cayó, el mapa
// se limpiará cuando su goroutine manejarCliente termine.
func broadcastMetrics(linea string) {
	clientesMu.RLock()
	defer clientesMu.RUnlock()
	for conn := range clientesActivos {
		fmt.Fprintf(conn, "%s\n", linea)
	}
}

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

// ── Broadcaster de métricas del servidor ──────────────────────────────────────

// iniciarBroadcasterMetricas lanza una goroutine que cada 5 s lee las métricas
// locales del servidor (CPU, RAM, Disco) y las envía a todos los clientes con
// el formato:   <<<METRICS:CPU=12.3|RAM_USED=1024|RAM_FREE=2048|DISK_USED=40.1|DISK_TOTAL=100.0>>>
// El cliente parsea ese prefijo y lo muestra en su panel derecho.
func iniciarBroadcasterMetricas(stop <-chan struct{}) {
	var prevIdle, prevTotal uint64
	prevIdle, prevTotal = leerStatCPU()

	go func() {
		for {
			select {
			case <-stop:
				return
			case <-time.After(5 * time.Second):
			}

			// CPU
			idle, total := leerStatCPU()
			dIdle := idle - prevIdle
			dTotal := total - prevTotal
			prevIdle = idle
			prevTotal = total

			var cpu float64
			if dTotal > 0 {
				cpu = (1.0 - float64(dIdle)/float64(dTotal)) * 100.0
			}

			// RAM
			ramUsed, ramFree := 0.0, 0.0
			if u, f, ok := leerRAM(); ok {
				ramUsed, ramFree = u, f
			}

			// Disco
			diskUsed, diskTotal := 0.0, 0.0
			if u, t, ok := leerDisco(); ok {
				diskUsed, diskTotal = u, t
			}

			linea := fmt.Sprintf(
				"<<<METRICS:CPU=%.1f|RAM_USED=%.0f|RAM_FREE=%.0f|DISK_USED=%.1f|DISK_TOTAL=%.1f>>>",
				cpu, ramUsed, ramFree, diskUsed, diskTotal,
			)
			broadcastMetrics(linea)
		}
	}()
}

// ── Manejo de cliente ─────────────────────────────────────────────────────────

// manejarCliente atiende una conexión individual en su propio goroutine.
// Lee líneas de texto (un comando por línea), ejecuta cada uno y responde
// con la salida seguida del marcador "<<<END>>>".
func manejarCliente(conn net.Conn, onLog func(string)) {
	defer func() {
		eliminarCliente(conn)
		conn.Close()
	}()

	registrarCliente(conn)
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

		// Obtener directorio actual tras ejecutar el comando (cd lo cambia)
		pwd, _ := os.Getwd()

		// Protocolo: primera línea = pwd, luego salida, luego marcador de fin
		fmt.Fprintf(conn, "<<<PWD:%s>>>\n%s\n<<<END>>>\n", pwd, salida)

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

// obtenerIPs devuelve todas las IPs locales de la máquina (excluye loopback).
func obtenerIPs() []string {
	var ips []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return []string{"(error obteniendo IPs)"}
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			if ip.To4() != nil {
				ips = append(ips, ip.String())
			}
		}
	}
	if len(ips) == 0 {
		return []string{"127.0.0.1 (solo loopback)"}
	}
	return ips
}

// iniciarServidor abre el socket TCP y acepta clientes indefinidamente.
// También arranca el broadcaster de métricas.
func iniciarServidor(onLog func(string), stop <-chan struct{}) error {
	ln, err := net.Listen("tcp", puertoServidor)
	if err != nil {
		return fmt.Errorf("no se pudo iniciar el servidor en %s: %w", puertoServidor, err)
	}

	onLog(fmt.Sprintf("◈  Servidor escuchando en %s", puertoServidor))
	onLog(fmt.Sprintf("◈  Log guardado en: %s", archivoLog))
	onLog("◈  Broadcast de métricas cada 5 s activado")

	// Iniciar broadcaster de métricas del servidor → clientes
	iniciarBroadcasterMetricas(stop)

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