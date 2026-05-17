package main

// auth.go
// Lógica de autenticación: lee contraseña.txt y compara el hash SHA-256.
// No depende de ningún paquete de GUI.

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
)

// validarCredenciales abre "contraseña.txt", busca la línea "usuario:hashSHA256"
// que coincida con las credenciales recibidas y devuelve true si las encuentra.
//
// Formato esperado del archivo (una entrada por línea):
//
//	admin:5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8
func validarCredenciales(usuario, contrasena string) (bool, error) {
	archivo, err := os.Open("contraseña.txt")
	if err != nil {
		return false, fmt.Errorf("no se pudo abrir el archivo: %w", err)
	}
	defer archivo.Close()

	hashIngresado := fmt.Sprintf("%x", sha256.Sum256([]byte(contrasena)))

	scanner := bufio.NewScanner(archivo)
	for scanner.Scan() {
		linea := strings.TrimSpace(scanner.Text())
		if linea == "" {
			continue
		}
		partes := strings.SplitN(linea, ":", 2)
		if len(partes) != 2 {
			continue
		}
		if strings.TrimSpace(partes[0]) == usuario &&
			strings.TrimSpace(partes[1]) == hashIngresado {
			return true, nil
		}
	}
	return false, scanner.Err()
}