package main

// shell.go
// Ejecución de comandos multiplataforma (Windows y Linux/macOS).
// Detecta el sistema operativo en tiempo de ejecución y usa el
// intérprete correcto: cmd.exe en Windows, bash en Linux/macOS.

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// ejecutarComando ejecuta un comando en el shell nativo del sistema operativo.
// Devuelve la salida combinada (stdout + stderr) y quit=true si el usuario
// escribió "bye" o "exit".
func ejecutarComando(comando string) (salida string, quit bool) {
	comando = strings.TrimSpace(comando)
	if comando == "" {
		return "", false
	}

	// Comandos especiales de sesión (iguales en ambos OS)
	if comando == "bye" || comando == "exit" {
		return "[ sesión terminada ]", true
	}

	campos := strings.Fields(comando)

	// ── cd: builtin que os/exec no puede manejar directamente ────────────
	if strings.EqualFold(campos[0], "cd") {
		if len(campos) > 1 {
			destino := strings.Join(campos[1:], " ")
			// Expandir ~ al directorio home
			if destino == "~" {
				if runtime.GOOS == "windows" {
					destino = os.Getenv("USERPROFILE")
				} else {
					destino = os.Getenv("HOME")
				}
			}
			if err := os.Chdir(destino); err != nil {
				return "  ✗ cd: " + err.Error(), false
			}
			pwd, _ := os.Getwd()
			return "  → " + pwd, false
		}
		// cd sin argumento → ir al home
		home := os.Getenv("HOME")
		if runtime.GOOS == "windows" {
			home = os.Getenv("USERPROFILE")
		}
		if home != "" {
			_ = os.Chdir(home)
			pwd, _ := os.Getwd()
			return "  → " + pwd, false
		}
		return "  ✗ cd: no se pudo determinar el directorio home.", false
	}

	// ── Construcción del comando según el OS ──────────────────────────────
	var shell *exec.Cmd
	if runtime.GOOS == "windows" {
		// cmd /C ejecuta el comando y termina (equivale a bash -c en Linux)
		shell = exec.Command("cmd", "/C", comando)
	} else {
		shell = exec.Command("bash", "-c", comando)
	}

	out, err := shell.CombinedOutput()
	resultado := strings.TrimRight(string(out), "\r\n")
	if err != nil && resultado == "" {
		resultado = "  ✗ error: " + err.Error()
	}
	return resultado, false
}

// sistemaOperativo devuelve una cadena legible con el nombre del OS actual.
// Usada en ventana.go para el mensaje de bienvenida.
func sistemaOperativo() string {
	switch runtime.GOOS {
	case "windows":
		return "Windows"
	case "linux":
		return "Linux"
	case "darwin":
		return "macOS"
	default:
		return runtime.GOOS
	}
}