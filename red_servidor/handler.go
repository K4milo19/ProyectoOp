package red_servidor

// handler.go
// Atiende una conexión TCP individual.
// Lee comandos línea a línea, los ejecuta y devuelve la salida
// junto con el directorio de trabajo actual (pwd).
//
// Protocolo de respuesta por comando:
//   <<<PWD:/ruta/actual>>>
//   <líneas de salida>
//   <<<END>>>

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"shellos/shell"
)

// ManejarCliente atiende una conexión individual en su propio goroutine.
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
