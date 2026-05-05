package main

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Abre el archivo, lee línea por línea y valida sin almacenar nada
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
			continue // línea mal formada, la salta
		}

		usuarioArchivo := strings.TrimSpace(partes[0])
		hashArchivo    := strings.TrimSpace(partes[1])

		if usuarioArchivo == usuario && hashArchivo == hashIngresado {
			return true, nil // encontró coincidencia, sale inmediatamente
		}
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("error leyendo el archivo: %w", err)
	}

	return false, nil // no encontró coincidencia
}

func main() {

	intentos := 0

	a := app.New()
	w := a.NewWindow("Login")
	w.Resize(fyne.NewSize(800, 600))

	inpUser := widget.NewEntry()
	inpUser.SetPlaceHolder("Ingrese su usuario")

	inpPass := widget.NewPasswordEntry()
	inpPass.SetPlaceHolder("Ingrese su contraseña")

	message  := widget.NewLabel("Digite su usuario y contraseña")
	message2 := widget.NewLabel("0")
	message3 := widget.NewLabel("❌ Usuario sospechoso - Plataforma cerrada ❌")

	button := widget.NewButton("Iniciar sesión", func() {
		uUser   := inpUser.Text
		uPasswd := inpPass.Text

		valido, err := validarCredenciales(uUser, uPasswd)

		if err != nil {
			message.SetText("Error al leer credenciales")
			return
		}

		if valido {
			message.SetText("LOGIN CORRECTO")
			intentos = 0
			message2.SetText("0")
		} else {
			intentos++
			message2.SetText(strconv.Itoa(intentos))
			message.SetText("Usuario o contraseña incorrectos")
			if intentos >= 3 {
				w.SetContent(container.NewCenter(container.NewVBox(message3)))
			}
		}
	})

	w.SetContent(container.NewCenter(container.NewVBox(
		message, inpUser, inpPass, button, message2,
	)))
	w.ShowAndRun()
}