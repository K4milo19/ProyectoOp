package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {

	// Variables
	intentos := 0

	// Ventana
	a := app.New()
	w := a.NewWindow("Login")
	w.Resize(fyne.NewSize(800, 600))

	// Mensajes en pantalla
	inpPass := widget.NewPasswordEntry()
	inpPass.SetPlaceHolder("Ingrese su contraseña")

	message := widget.NewLabel("Digite su contraseña")
	message2 := widget.NewLabel("0")
	message3 := widget.NewLabel("❌ Usuario sospechoso - Plataforma cerrada ❌")

	// Leer contraseña guardada
	hdbPasswdBytes, _ := os.ReadFile("contraseña.txt")
	hdbPasswd := strings.TrimSpace(string(hdbPasswdBytes))

	// Boton
	button := widget.NewButton("Iniciar sesión", func() {
		uPasswd := inpPass.Text
		vUserHash := sha256.Sum256([]byte(uPasswd))
		huPasswd := fmt.Sprintf("%x", vUserHash)

		if huPasswd == hdbPasswd {
			message.SetText("LOGIN CORRECTO")
			intentos = 0
			message2.SetText("0")

		} else {
			intentos++
			message2.SetText(strconv.Itoa(intentos))
			message.SetText("Contraseña incorrecta")
			if intentos >= 3 {
				// Cambiar cont
				w.SetContent(container.NewCenter(container.NewVBox(message3)))
			}
		}
	})

	// Pantalla inicial
	w.SetContent(container.NewCenter(container.NewVBox(message,inpPass,button,message2,)))
	w.ShowAndRun()
}