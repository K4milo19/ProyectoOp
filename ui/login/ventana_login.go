package login

// ventana_login.go
// Pantalla de autenticación de ShellOS.
// Valida usuario + contraseña; bloquea tras 3 intentos fallidos.

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"shellos/auth"
	"shellos/tema"
	"shellos/ui/selector"
)

// MostrarLogin muestra la pantalla de autenticación.
// onExito se llama con el modo elegido cuando el login es correcto.
func MostrarLogin(a fyne.App, modo selector.ModoApp, onExito func()) {
	intentos := 0

	w := a.NewWindow("ShellOS — Autenticación")
	w.Resize(fyne.NewSize(460, 500))
	w.CenterOnScreen()

	// ── Cabecera ──────────────────────────────────────────────────────────
	logo := canvas.NewText("S H E L L O S", tema.ColPrimary)
	logo.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	logo.TextSize = 28
	logo.Alignment = fyne.TextAlignCenter

	modotxt := "[ modo: SERVIDOR ]"
	if modo == selector.ModoCliente {
		modotxt = "[ modo: CLIENTE ]"
	}
	subModo := canvas.NewText(modotxt, tema.ColAccent)
	subModo.TextStyle = fyne.TextStyle{Monospace: true}
	subModo.TextSize = 11
	subModo.Alignment = fyne.TextAlignCenter

	sub := canvas.NewText("secure authentication layer", tema.ColMuted)
	sub.TextStyle = fyne.TextStyle{Monospace: true}
	sub.TextSize = 11
	sub.Alignment = fyne.TextAlignCenter

	deco := canvas.NewRectangle(tema.ColAccent)
	deco.SetMinSize(fyne.NewSize(48, 2))

	cabecera := container.NewVBox(
		tema.Spacer(),
		container.NewCenter(logo),
		container.NewCenter(subModo),
		container.NewCenter(sub),
		tema.Spacer(),
		container.NewCenter(container.NewHBox(deco)),
		tema.Spacer(),
	)

	// ── Campos ───────────────────────────────────────────────────────────
	lblUser := canvas.NewText("USUARIO", tema.ColMuted)
	lblUser.TextStyle = fyne.TextStyle{Monospace: true}
	lblUser.TextSize = 10

	inpUser := widget.NewEntry()
	inpUser.SetPlaceHolder("ingresa tu usuario")
	inpUser.TextStyle = fyne.TextStyle{Monospace: true}

	lblPass := canvas.NewText("CONTRASEÑA", tema.ColMuted)
	lblPass.TextStyle = fyne.TextStyle{Monospace: true}
	lblPass.TextSize = 10

	inpPass := widget.NewPasswordEntry()
	inpPass.SetPlaceHolder("••••••••")
	inpPass.TextStyle = fyne.TextStyle{Monospace: true}

	// ── Estado ────────────────────────────────────────────────────────────
	msgStatus := canvas.NewText("", tema.ColMuted)
	msgStatus.TextStyle = fyne.TextStyle{Monospace: true}
	msgStatus.TextSize = 12
	msgStatus.Alignment = fyne.TextAlignCenter

	msgIntentos := canvas.NewText("", tema.ColMuted)
	msgIntentos.TextStyle = fyne.TextStyle{Monospace: true}
	msgIntentos.TextSize = 10
	msgIntentos.Alignment = fyne.TextAlignCenter

	// ── Botón ─────────────────────────────────────────────────────────────
	button := widget.NewButton("[ INICIAR SESIÓN ]", nil)
	button.Importance = widget.HighImportance

	accion := func() {
		valido, err := auth.ValidarCredenciales(inpUser.Text, inpPass.Text)
		if err != nil {
			msgStatus.Text = "✗  error al leer credenciales"
			msgStatus.Color = tema.ColError
			msgStatus.Refresh()
			return
		}
		if valido {
			w.Hide()
			onExito()
			return
		}
		intentos++
		msgIntentos.Text = fmt.Sprintf("intentos: %d de 3", intentos)
		msgIntentos.Refresh()
		msgStatus.Text = "✗  credenciales incorrectas"
		msgStatus.Color = tema.ColError
		msgStatus.Refresh()

		if intentos >= 3 {
			mostrarPantallaBloqueo(w, intentos)
		}
	}

	button.OnTapped = accion
	inpPass.OnSubmitted = func(_ string) { accion() }

	formulario := container.NewVBox(lblUser, inpUser, tema.Spacer(), lblPass, inpPass)
	pie := container.NewVBox(
		container.NewCenter(msgStatus),
		container.NewCenter(msgIntentos),
	)
	cuerpo := container.NewVBox(
		cabecera, tema.HRule(), tema.Spacer(),
		formulario, tema.Spacer(), tema.HRule(), tema.Spacer(),
		container.NewCenter(button), tema.Spacer(), pie,
	)

	w.SetContent(container.NewPadded(container.NewPadded(cuerpo)))
	w.Canvas().Focus(inpUser)
	w.Show()
}

// mostrarPantallaBloqueo reemplaza el contenido de la ventana con un aviso
// de acceso bloqueado tras agotar los intentos permitidos.
func mostrarPantallaBloqueo(w fyne.Window, intentosFallidos int) {
	blk1 := canvas.NewText("⛔  ACCESO BLOQUEADO", tema.ColError)
	blk1.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	blk1.TextSize = 18
	blk1.Alignment = fyne.TextAlignCenter

	blk2 := canvas.NewText("usuario sospechoso — plataforma cerrada", tema.ColMuted)
	blk2.TextStyle = fyne.TextStyle{Monospace: true}
	blk2.TextSize = 12
	blk2.Alignment = fyne.TextAlignCenter

	blk3 := canvas.NewText(fmt.Sprintf("[ %d intentos fallidos ]", intentosFallidos), tema.ColMuted)
	blk3.TextStyle = fyne.TextStyle{Monospace: true}
	blk3.TextSize = 11
	blk3.Alignment = fyne.TextAlignCenter

	w.SetContent(container.NewCenter(container.NewVBox(
		tema.Spacer(), tema.Spacer(),
		container.NewCenter(blk1), tema.Spacer(),
		container.NewCenter(blk2), tema.Spacer(),
		container.NewCenter(blk3),
		tema.Spacer(), tema.Spacer(),
	)))
}
