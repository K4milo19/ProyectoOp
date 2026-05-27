package selector

// ventana_selector.go
// Primera ventana del flujo: el usuario elige si este nodo operará
// como Servidor o como Cliente.

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"shellos/tema"
)

// ModoApp distingue los dos roles posibles del nodo.
type ModoApp int

const (
	ModoServidor ModoApp = iota
	ModoCliente
)

// MostrarSeleccionModo muestra la ventana de elección de modo.
// Al pulsar un botón llama a onModo con el modo elegido.
func MostrarSeleccionModo(a fyne.App, onModo func(ModoApp)) {
	w := a.NewWindow("ShellOS — Selección de modo")
	w.Resize(fyne.NewSize(480, 420))
	w.CenterOnScreen()

	// ── Cabecera ──────────────────────────────────────────────────────────
	logo := canvas.NewText("S H E L L O S", tema.ColPrimary)
	logo.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	logo.TextSize = 28
	logo.Alignment = fyne.TextAlignCenter

	sub := canvas.NewText("selecciona el modo de operación", tema.ColMuted)
	sub.TextStyle = fyne.TextStyle{Monospace: true}
	sub.TextSize = 11
	sub.Alignment = fyne.TextAlignCenter

	deco := canvas.NewRectangle(tema.ColAccent)
	deco.SetMinSize(fyne.NewSize(48, 2))

	// ── Tarjeta SERVIDOR ──────────────────────────────────────────────────
	icnSrv := canvas.NewText("▣", tema.ColAccent)
	icnSrv.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	icnSrv.TextSize = 28

	tituloSrv := canvas.NewText("SERVIDOR", tema.ColPrimary)
	tituloSrv.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	tituloSrv.TextSize = 15

	descSrv := canvas.NewText("Acepta conexiones remotas.\nEjecuta comandos y guarda logs.txt.", tema.ColMuted)
	descSrv.TextStyle = fyne.TextStyle{Monospace: true}
	descSrv.TextSize = 11

	btnSrv := widget.NewButton("[ INICIAR COMO SERVIDOR ]", func() {
		w.Hide()
		onModo(ModoServidor)
	})
	btnSrv.Importance = widget.HighImportance

	cardSrv := container.NewVBox(
		container.NewCenter(icnSrv),
		container.NewCenter(tituloSrv),
		container.NewCenter(descSrv),
		tema.Spacer(),
		container.NewCenter(btnSrv),
	)

	// ── Tarjeta CLIENTE ───────────────────────────────────────────────────
	icnCli := canvas.NewText("▷", tema.ColAccent)
	icnCli.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	icnCli.TextSize = 28

	tituloCli := canvas.NewText("CLIENTE", tema.ColPrimary)
	tituloCli.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	tituloCli.TextSize = 15

	descCli := canvas.NewText("Conecta a un servidor remoto.\nEnvía comandos y ve la salida.", tema.ColMuted)
	descCli.TextStyle = fyne.TextStyle{Monospace: true}
	descCli.TextSize = 11

	btnCli := widget.NewButton("[ INICIAR COMO CLIENTE ]", func() {
		w.Hide()
		onModo(ModoCliente)
	})
	btnCli.Importance = widget.MediumImportance

	cardCli := container.NewVBox(
		container.NewCenter(icnCli),
		container.NewCenter(tituloCli),
		container.NewCenter(descCli),
		tema.Spacer(),
		container.NewCenter(btnCli),
	)

	// ── Layout ────────────────────────────────────────────────────────────
	cabecera := container.NewVBox(
		tema.Spacer(),
		container.NewCenter(logo),
		container.NewCenter(sub),
		tema.Spacer(),
		container.NewCenter(container.NewHBox(deco)),
		tema.Spacer(),
	)

	cuerpo := container.NewVBox(
		cabecera,
		tema.HRule(),
		tema.Spacer(),
		container.NewGridWithColumns(2, cardSrv, cardCli),
		tema.Spacer(),
	)

	w.SetContent(container.NewPadded(container.NewPadded(cuerpo)))
	w.ShowAndRun()
}
