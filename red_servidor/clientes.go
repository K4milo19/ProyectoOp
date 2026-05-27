package red_servidor

// clientes.go
// Registro thread-safe de las conexiones TCP activas.
// Permite al broadcaster de métricas escribir a todos los clientes.

import (
	"net"
	"sync"
)

var (
	clientesMu      sync.RWMutex
	clientesActivos = make(map[net.Conn]struct{})
)

// registrarCliente añade una conexión al mapa de activos.
func registrarCliente(conn net.Conn) {
	clientesMu.Lock()
	clientesActivos[conn] = struct{}{}
	clientesMu.Unlock()
}

// eliminarCliente retira una conexión del mapa de activos.
func eliminarCliente(conn net.Conn) {
	clientesMu.Lock()
	delete(clientesActivos, conn)
	clientesMu.Unlock()
}

// broadcastMetrics envía una línea de texto a todos los clientes conectados.
// Los errores de escritura se ignoran; las conexiones caídas se limpian
// cuando su goroutine manejarCliente termina.
func broadcastMetrics(linea string) {
	clientesMu.RLock()
	defer clientesMu.RUnlock()
	for conn := range clientesActivos {
		conn.Write([]byte(linea + "\n")) //nolint:errcheck
	}
}
