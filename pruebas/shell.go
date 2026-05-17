package main

// shell.go
// Ejecución de comandos del sistema operativo.
// No depende de ningún paquete de GUI; solo usa os y os/exec.

import (
	"os"
	"os/exec"
	"strings"
)

// ejecutarComando ejecuta un comando de shell y devuelve su salida combinada.
// Si el comando es "bye" o "exit" devuelve quit=true para señalar el cierre.
// El comando "cd" se maneja internamente cambiando el directorio de trabajo.
func ejecutarComando(comando string) (salida string, quit bool) {
	comando = strings.TrimSpace(comando)
	if comando == "" {
		return "", false
	}

	// Comandos especiales de sesión
	if comando == "bye" || comando == "exit" {
		return "[ sesión terminada ]", true
	}

	// Cambio de directorio (cd no funciona con exec.Command por ser un builtin)
	sl := strings.Fields(comando)
	if sl[0] == "cd" {
		if len(sl) > 1 {
			if err := os.Chdir(sl[1]); err != nil {
				return "  ✗ cd: " + err.Error(), false
			}
			pwd, _ := os.Getwd()
			return "  → " + pwd, false
		}
		return "  ✗ cd: especifica un directorio.", false
	}

	// Cualquier otro comando
	shell := exec.Command("bash", "-c", comando)
	out, err := shell.CombinedOutput()
	resultado := strings.TrimRight(string(out), "\n")
	if err != nil && resultado == "" {
		resultado = "  ✗ error: " + err.Error()
	}
	return resultado, false
}