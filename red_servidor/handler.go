package red_servidor

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"shellos/shell"
)

func ManejarCliente(conn net.Conn, onLog func(string)) {
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

		if linea == "bye" || linea == "exit" {
			fmt.Fprintf(conn, "[ sesión cerrada por el cliente ]\n<<<END>>>\n")
			onLog(fmt.Sprintf("[ DESCONEXIÓN ]  %s cerró la sesión", remoto))
			return
		}

		salida, _ := shell.EjecutarComando(linea)
		if salida == "" {
			salida = "[ sin salida ]"
		}

		pwd, _ := os.Getwd()
		fmt.Fprintf(conn, "<<<PWD:%s>>>\n%s\n<<<END>>>\n", pwd, salida)

		EscribirLog(remoto, linea, salida)
		onLog(fmt.Sprintf("[ CMD ]  %s  →  %s", remoto, linea))
	}

	if err := scanner.Err(); err != nil {
		onLog(fmt.Sprintf("[ ERROR ]  %s: %v", remoto, err))
	} else {
		onLog(fmt.Sprintf("[ DESCONEXIÓN ]  %s", remoto))
	}
}

// RechazarCliente envía un aviso al cliente y cierra la conexión.
func RechazarCliente(conn net.Conn) {
	fmt.Fprintf(conn, "<<<BLOQUEADO:Tu IP no tiene permiso para conectarse a este servidor>>>\n")
	conn.Close()
}