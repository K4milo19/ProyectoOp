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

// ─── Validación ───────────────────────────────────────────────────────────────

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

// ─── Ventana principal ────────────────────────────────────────────────────────

func mostrarVentanaPrincipal(a fyne.App) {
	w := a.NewWindow("Panel principal")
	w.Resize(fyne.NewSize(800, 500))

	// Panel derecho: lista de mensajes recibidos
	mensajes := widget.NewList(
		func() int { return 0 },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {},
	)

	textos := []string{}

	// Panel izquierdo: entrada de texto + botón enviar
	entrada := widget.NewMultiLineEntry()
	entrada.SetPlaceHolder("Escribe tu mensaje aquí...")
	entrada.Wrapping = fyne.TextWrapWord

	botonEnviar := widget.NewButton("Enviar", func() {
		texto := strings.TrimSpace(entrada.Text)
		if texto == "" {
			return
		}
		if texto == "bye"{
			a.Quit()
			return
		}
		textos = append(textos, texto)

		// Actualizar la lista con los nuevos datos
		mensajes.Length = func() int { return len(textos) }
		mensajes.UpdateItem = func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(textos[id])
		}
		mensajes.Refresh()
		mensajes.ScrollToBottom()

		entrada.SetText("")
	})

	panelIzquierdo := container.NewBorder(nil, botonEnviar, nil, nil, entrada)
	panelDerecho   := mensajes

	// HSplit divide la ventana exactamente por la mitad
	split := container.NewVSplit(panelIzquierdo, panelDerecho)
	split.SetOffset(0.3)

	w.SetContent(split)
	w.Show()
}

// ─── Login ────────────────────────────────────────────────────────────────────

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
	message2 := widget.NewLabel("Intentos: 0")
	message3 := widget.NewLabel("❌ Usuario sospechoso - Plataforma cerrada ❌")

	button := widget.NewButton("Iniciar sesión", func() {
		valido, err := validarCredenciales(inpUser.Text, inpPass.Text)
		if err != nil {
			message.SetText("Error al leer credenciales")
			return
		}
		if valido {
			w.Hide()
			mostrarVentanaPrincipal(a)
		} else {
			intentos++
			message2.SetText("Intentos: " + strconv.Itoa(intentos))
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