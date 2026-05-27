package red_cliente

// comandos.go
// Envía un comando al servidor y recoge la respuesta completa.
// Lee del canal Respuestas (ya separado de métricas por lectorSocket).

import (
	"fmt"
	"strings"
)

// EnviarComando manda un comando al servidor y espera hasta recibir "<<<END>>>".
// Devuelve la salida acumulada, el nuevo directorio de trabajo (pwd) y cualquier
// error de red.
func (c *ConexionServidor) EnviarComando(cmd string) (salida, pwd string, err error) {
	_, err = fmt.Fprintf(c.conn, "%s\n", cmd)
	if err != nil {
		return "", "", fmt.Errorf("error al enviar: %w", err)
	}

	var sb strings.Builder
	for linea := range c.Respuestas {
		if linea == "<<<END>>>" {
			break
		}
		if strings.HasPrefix(linea, "<<<PWD:") && strings.HasSuffix(linea, ">>>") {
			pwd = strings.TrimSuffix(strings.TrimPrefix(linea, "<<<PWD:"), ">>>")
			continue
		}
		sb.WriteString(linea)
		sb.WriteByte('\n')
	}
	return strings.TrimRight(sb.String(), "\n"), pwd, nil
}
