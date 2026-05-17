package main

// main.go
// Punto de entrada de ShellOS.
// Responsabilidades: crear la app, aplicar el tema, construir la UI de login
// y delegar en los otros archivos según el flujo de autenticación.

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	intentos := 0

	a := app.New()
	a.Settings().SetTheme(&shellTheme{}) // definido en theme.go

	w := a.NewWindow("ShellOS — Autenticación")
	w.Resize(fyne.NewSize(460, 500))
	w.CenterOnScreen()

	// ── Cabecera ──────────────────────────────────────────────────────────
	logo := canvas.NewText("S H E L L O S", colPrimary)
	logo.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	logo.TextSize = 28
	logo.Alignment = fyne.TextAlignCenter

	sub := canvas.NewText("secure authentication layer", colMuted)
	sub.TextStyle = fyne.TextStyle{Monospace: true}
	sub.TextSize = 11
	sub.Alignment = fyne.TextAlignCenter

	deco := canvas.NewRectangle(colAccent)
	deco.SetMinSize(fyne.NewSize(48, 2))

	cabecera := container.NewVBox(
		spacer(),
		container.NewCenter(logo),
		container.NewCenter(sub),
		spacer(),
		container.NewCenter(container.NewHBox(deco)),
		spacer(),
	)

	// ── Campos de entrada ─────────────────────────────────────────────────
	lblUser := canvas.NewText("USUARIO", colMuted)
	lblUser.TextStyle = fyne.TextStyle{Monospace: true}
	lblUser.TextSize = 10

	inpUser := widget.NewEntry()
	inpUser.SetPlaceHolder("ingresa tu usuario")
	inpUser.TextStyle = fyne.TextStyle{Monospace: true}

	lblPass := canvas.NewText("CONTRASEÑA", colMuted)
	lblPass.TextStyle = fyne.TextStyle{Monospace: true}
	lblPass.TextSize = 10

	inpPass := widget.NewPasswordEntry()
	inpPass.SetPlaceHolder("••••••••")
	inpPass.TextStyle = fyne.TextStyle{Monospace: true}

	// ── Mensajes de estado ────────────────────────────────────────────────
	msgStatus := canvas.NewText("", colMuted)
	msgStatus.TextStyle = fyne.TextStyle{Monospace: true}
	msgStatus.TextSize = 12
	msgStatus.Alignment = fyne.TextAlignCenter

	msgIntentos := canvas.NewText("", colMuted)
	msgIntentos.TextStyle = fyne.TextStyle{Monospace: true}
	msgIntentos.TextSize = 10
	msgIntentos.Alignment = fyne.TextAlignCenter

	// ── Botón de login ────────────────────────────────────────────────────
	button := widget.NewButton("[ INICIAR SESIÓN ]", nil)
	button.Importance = widget.HighImportance

	accion := func() {
		// validarCredenciales está definido en auth.go
		valido, err := validarCredenciales(inpUser.Text, inpPass.Text)
		if err != nil {
			msgStatus.Text = "✗  error al leer credenciales"
			msgStatus.Color = colError
			msgStatus.Refresh()
			return
		}

		if valido {
			w.Hide()
			// mostrarVentanaPrincipal está definido en ventana.go
			mostrarVentanaPrincipal(a)
			return
		}

		// Credenciales incorrectas
		intentos++
		msgIntentos.Text = fmt.Sprintf("intentos: %d de 3", intentos)
		msgIntentos.Refresh()
		msgStatus.Text = "✗  credenciales incorrectas"
		msgStatus.Color = colError
		msgStatus.Refresh()

		if intentos >= 3 {
			mostrarPantallaBloqueo(w, intentos)
		}
	}

	button.OnTapped = accion
	inpPass.OnSubmitted = func(_ string) { accion() }

	// ── Layout ────────────────────────────────────────────────────────────
	formulario := container.NewVBox(
		lblUser, inpUser,
		spacer(),
		lblPass, inpPass,
	)
	pie := container.NewVBox(
		container.NewCenter(msgStatus),
		container.NewCenter(msgIntentos),
	)
	cuerpo := container.NewVBox(
		cabecera,
		hRule(),
		spacer(),
		formulario,
		spacer(),
		hRule(),
		spacer(),
		container.NewCenter(button),
		spacer(),
		pie,
	)

	w.SetContent(container.NewPadded(container.NewPadded(cuerpo)))
	w.Canvas().Focus(inpUser)
	w.ShowAndRun()
}

// mostrarPantallaBloqueo reemplaza el contenido de la ventana con un mensaje
// de acceso bloqueado tras agotar los intentos permitidos.
func mostrarPantallaBloqueo(w fyne.Window, intentosFallidos int) {
	blk1 := canvas.NewText("⛔  ACCESO BLOQUEADO", colError)
	blk1.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	blk1.TextSize = 18
	blk1.Alignment = fyne.TextAlignCenter

	blk2 := canvas.NewText("usuario sospechoso — plataforma cerrada", colMuted)
	blk2.TextStyle = fyne.TextStyle{Monospace: true}
	blk2.TextSize = 12
	blk2.Alignment = fyne.TextAlignCenter

	blk3 := canvas.NewText(fmt.Sprintf("[ %d intentos fallidos ]", intentosFallidos), colMuted)
	blk3.TextStyle = fyne.TextStyle{Monospace: true}
	blk3.TextSize = 11
	blk3.Alignment = fyne.TextAlignCenter

	w.SetContent(container.NewCenter(container.NewVBox(
		spacer(), spacer(),
		container.NewCenter(blk1),
		spacer(),
		container.NewCenter(blk2),
		spacer(),
		container.NewCenter(blk3),
		spacer(), spacer(),
	)))
}