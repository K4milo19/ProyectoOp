package main

// shell.go
// Ejecución de comandos — solo Linux.
// Usa bash como intérprete y maneja cd como builtin interno.

import (
	"os"
	"os/exec"
	"strings"
)

// ejecutarComando ejecuta un comando bash en el sistema.
// Devuelve la salida combinada (stdout + stderr) y quit=true si el usuario
// escribió "bye" o "exit".
func ejecutarComando(comando string) (salida string, quit bool) {
	comando = strings.TrimSpace(comando)
	if comando == "" {
		return "", false
	}

	if comando == "bye" || comando == "exit" {
		return "[ sesión terminada ]", true
	}

	campos := strings.Fields(comando)

	// cd es un builtin: os/exec no puede cambiar el directorio del proceso padre
	if campos[0] == "cd" {
		if len(campos) > 1 {
			destino := strings.Join(campos[1:], " ")
			if destino == "~" {
				destino = os.Getenv("HOME")
			}
			if err := os.Chdir(destino); err != nil {
				return "  ✗ cd: " + err.Error(), false
			}
			pwd, _ := os.Getwd()
			return "  → " + pwd, false
		}
		// cd sin argumento → home
		home := os.Getenv("HOME")
		if home != "" {
			_ = os.Chdir(home)
			pwd, _ := os.Getwd()
			return "  → " + pwd, false
		}
		return "  ✗ cd: HOME no definido.", false
	}

	cmd := exec.Command("bash", "-c", comando)
	out, err := cmd.CombinedOutput()
	resultado := strings.TrimRight(string(out), "\n")
	if err != nil && resultado == "" {
		resultado = "  ✗ error: " + err.Error()
	}
	return resultado, false
}

// sistemaOperativo devuelve el nombre del OS actual.
func sistemaOperativo() string {
	return "Linux"
}