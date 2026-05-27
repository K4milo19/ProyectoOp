package red_cliente

// conexion.go
// Abre la conexión TCP con el servidor y lanza el goroutine lector
// que clasifica las líneas entrantes en dos canales: respuestas y métricas.

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

const PuertoServidor = ":9000"

// ConexionServidor representa una sesión TCP activa con el servidor ShellOS.
type ConexionServidor struct {
	conn net.Conn
	// respuestas recibe las líneas del socket que NO son métricas.
	Respuestas chan string
	// metricas recibe las líneas <<<METRICS:...>>> del servidor.
	Metricas   chan string
	// OnMetrics es el callback que la UI asigna para actualizar su panel.
	OnMetrics  func(cpu, ramUsed, ramFree, diskUsed, diskTotal float64)
}

// Conectar abre una conexión TCP al servidor y arranca el lector de socket.
func Conectar(host string) (*ConexionServidor, error) {
	addr := host + PuertoServidor
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("no se pudo conectar a %s: %w", addr, err)
	}
	cs := &ConexionServidor{
		conn:       conn,
		Respuestas: make(chan string, 256),
		Metricas:   make(chan string, 16),
	}
	go cs.lectorSocket()
	return cs, nil
}

// lectorSocket es el ÚNICO goroutine que lee del socket.
// Clasifica cada línea y la envía al canal correspondiente.
// Se detiene cuando la conexión se cierra.
func (c *ConexionServidor) lectorSocket() {
	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		linea := scanner.Text()
		if strings.HasPrefix(linea, "<<<METRICS:") {
			select {
			case c.Metricas <- linea:
			default: // descartar si nadie lee
			}
		} else {
			c.Respuestas <- linea
		}
	}
	close(c.Respuestas)
	close(c.Metricas)
}

// Cerrar cierra la conexión TCP.
func (c *ConexionServidor) Cerrar() {
	if c.conn != nil {
		c.conn.Close()
	}
}
