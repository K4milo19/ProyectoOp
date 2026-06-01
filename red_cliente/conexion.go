package red_cliente

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

const PuertoServidor = ":9000"

type ConexionServidor struct {
	conn       net.Conn
	Respuestas chan string
	Metricas   chan string
	OnMetrics  func(cpu, ramUsed, ramFree, diskUsed, diskTotal float64)
	// OnBloqueado se llama si el servidor rechaza la conexión por IP.
	OnBloqueado func(motivo string)
}

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

func (c *ConexionServidor) lectorSocket() {
	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		linea := scanner.Text()
		if strings.HasPrefix(linea, "<<<METRICS:") {
			select {
			case c.Metricas <- linea:
			default:
			}
		} else if strings.HasPrefix(linea, "<<<BLOQUEADO:") && strings.HasSuffix(linea, ">>>") {
			motivo := strings.TrimSuffix(strings.TrimPrefix(linea, "<<<BLOQUEADO:"), ">>>")
			if c.OnBloqueado != nil {
				c.OnBloqueado(motivo)
			}
		} else {
			c.Respuestas <- linea
		}
	}
	close(c.Respuestas)
	close(c.Metricas)
}

func (c *ConexionServidor) Cerrar() {
	if c.conn != nil {
		c.conn.Close()
	}
}