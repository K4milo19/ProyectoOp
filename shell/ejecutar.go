package shell

// ejecutar.go
// Ejecuta comandos del sistema usando bash como intérprete.
// Maneja "cd" como builtin interno porque os/exec no puede cambiar
// el directorio de trabajo del proceso padre.

import (
	"os"
	"os/exec"
	"strings"
)

// EjecutarComando ejecuta un comando en bash.
// Devuelve la salida combinada (stdout + stderr) y quit=true si el
// usuario escribió "bye" o "exit".
func EjecutarComando(comando string) (salida string, quit bool) {
	comando = strings.TrimSpace(comando)
	if comando == "" {
		return "", false
	}

	if comando == "bye" || comando == "exit" {
		return "[ sesión terminada ]", true
	}

	campos := strings.Fields(comando)

	// "cd" es un builtin: debe cambiar el directorio del proceso actual
	if campos[0] == "cd" {
		return ejecutarCd(campos[1:])
	}

	cmd := exec.Command("bash", "-c", comando)
	out, err := cmd.CombinedOutput()
	resultado := strings.TrimRight(string(out), "\n")
	if err != nil && resultado == "" {
		resultado = "  ✗ error: " + err.Error()
	}
	return resultado, false
}

// ejecutarCd cambia el directorio de trabajo del proceso.
func ejecutarCd(args []string) (salida string, quit bool) {
	destino := ""
	if len(args) > 0 {
		destino = strings.Join(args, " ")
	}
	if destino == "" || destino == "~" {
		destino = os.Getenv("HOME")
		if destino == "" {
			return "  ✗ cd: HOME no definido.", false
		}
	}
	if err := os.Chdir(destino); err != nil {
		return "  ✗ cd: " + err.Error(), false
	}
	pwd, _ := os.Getwd()
	return "  → " + pwd, false
}
