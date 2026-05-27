package terminal_cliente

// ventana_conexion.go
// Ventana donde el usuario escribe la IP del servidor antes de conectar.

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"shellos/red_cliente"
	"shellos/tema"
)

// MostrarVentanaConexion muestra el formulario de IP del servidor.
func MostrarVentanaConexion(a fyne.App) {
	w := a.NewWindow("ShellOS — Conectar al servidor")
	w.Resize(fyne.NewSize(420, 320))
	w.CenterOnScreen()

	titulo := canvas.NewText("MODO CLIENTE", tema.ColPrimary)
	titulo.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	titulo.TextSize = 22
	titulo.Alignment = fyne.TextAlignCenter

	sub := canvas.NewText("conexión remota vía TCP", tema.ColMuted)
	sub.TextStyle = fyne.TextStyle{Monospace: true}
	sub.TextSize = 11
	sub.Alignment = fyne.TextAlignCenter

	lblIP := canvas.NewText("IP DEL SERVIDOR", tema.ColMuted)
	lblIP.TextStyle = fyne.TextStyle{Monospace: true}
	lblIP.TextSize = 10

	inpIP := widget.NewEntry()
	inpIP.SetPlaceHolder("192.168.1.x  o  localhost")
	inpIP.TextStyle = fyne.TextStyle{Monospace: true}
	inpIP.Text = "localhost"

	msgErr := canvas.NewText("", tema.ColError)
	msgErr.TextStyle = fyne.TextStyle{Monospace: true}
	msgErr.TextSize = 11
	msgErr.Alignment = fyne.TextAlignCenter

	btnConectar := widget.NewButton("[ CONECTAR ]", nil)
	btnConectar.Importance = widget.HighImportance

	accion := func() {
		host := strings.TrimSpace(inpIP.Text)
		if host == "" {
			msgErr.Text = "✗  ingresa una IP o hostname"
			msgErr.Refresh()
			return
		}
		msgErr.Text = "  conectando..."
		msgErr.Color = tema.ColMuted
		msgErr.Refresh()

		cs, err := red_cliente.Conectar(host)
		if err != nil {
			msgErr.Text = "✗  " + err.Error()
			msgErr.Color = tema.ColError
			msgErr.Refresh()
			return
		}
		w.Hide()
		MostrarTerminalCliente(a, cs, host)
	}

	btnConectar.OnTapped = accion
	inpIP.OnSubmitted = func(_ string) { accion() }

	cuerpo := container.NewVBox(
		tema.Spacer(),
		container.NewCenter(titulo),
		container.NewCenter(sub),
		tema.Spacer(),
		tema.HRule(),
		tema.Spacer(),
		lblIP, inpIP,
		tema.Spacer(),
		container.NewCenter(btnConectar),
		tema.Spacer(),
		container.NewCenter(msgErr),
	)
	w.SetContent(container.NewPadded(container.NewPadded(cuerpo)))
	w.Canvas().Focus(inpIP)
	w.Show()
}
