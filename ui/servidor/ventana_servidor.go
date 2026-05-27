package servidor

// ventana_servidor.go
// Panel de control del nodo servidor: muestra IPs locales,
// información del sistema y un log en vivo de conexiones/comandos.

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"shellos/red_servidor"
	"shellos/tema"
)

// MostrarVentanaServidor abre el panel del servidor.
func MostrarVentanaServidor(a fyne.App) {
	w := a.NewWindow("ShellOS — Servidor")
	w.Resize(fyne.NewSize(860, 620))
	w.CenterOnScreen()

	stop := make(chan struct{})
	w.SetOnClosed(func() { close(stop) })

	// ── Título ────────────────────────────────────────────────────────────
	titBar := canvas.NewText("▣  ShellOS  —  modo SERVIDOR", tema.ColAccent)
	titBar.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	titBar.TextSize = 13

	estadoLbl := canvas.NewText("● iniciando...", tema.ColMuted)
	estadoLbl.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	estadoLbl.TextSize = 11

	barTitulo := container.NewBorder(nil, nil, nil,
		container.NewPadded(estadoLbl),
		container.NewPadded(titBar),
	)

	// ── IPs locales ───────────────────────────────────────────────────────
	ips := red_servidor.ObtenerIPs()
	ipResumen := strings.Join(ips, "  |  ")

	infoLbl := canvas.NewText(
		fmt.Sprintf("  OS: %s    Puerto: %s    Log: %s",
			"Linux", red_servidor.PuertoServidor, red_servidor.ArchivoLog),
		tema.ColMuted,
	)
	infoLbl.TextStyle = fyne.TextStyle{Monospace: true}
	infoLbl.TextSize = 11

	ipTitulo := canvas.NewText("  IP(s) para conectar clientes:", tema.ColAccent)
	ipTitulo.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	ipTitulo.TextSize = 12

	ipValor := canvas.NewText("  "+ipResumen, tema.ColPrimary)
	ipValor.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	ipValor.TextSize = 14

	panelIP := container.NewVBox(
		container.NewPadded(infoLbl),
		container.NewPadded(ipTitulo),
		container.NewPadded(ipValor),
	)

	// ── Log en vivo ───────────────────────────────────────────────────────
	logLabel := widget.NewLabel("")
	logLabel.TextStyle = fyne.TextStyle{Monospace: true}
	logLabel.Wrapping = fyne.TextWrapWord
	scroll := container.NewVScroll(logLabel)

	var lineas []string
	logCh := make(chan string, 64)
	go func() {
		for msg := range logCh {
			lineas = append(lineas, "  "+msg)
			logLabel.SetText(strings.Join(lineas, "\n"))
			scroll.ScrollToBottom()
		}
	}()
	onLog := func(msg string) { logCh <- msg }

	onLog("═══════════════════════════════════════════════════════════")
	onLog("  ShellOS  —  Servidor TCP iniciando")
	onLog("═══════════════════════════════════════════════════════════")
	for _, ip := range ips {
		onLog(fmt.Sprintf("  ◈  IP disponible:  %s%s", ip, red_servidor.PuertoServidor))
	}
	onLog("═══════════════════════════════════════════════════════════")

	// ── Iniciar servidor ──────────────────────────────────────────────────
	err := red_servidor.IniciarServidor(onLog, stop)
	if err != nil {
		estadoLbl.Text = "● ERROR"
		estadoLbl.Color = tema.ColError
		estadoLbl.Refresh()
		onLog("✗  " + err.Error())
	} else {
		estadoLbl.Text = fmt.Sprintf("● ESCUCHANDO  %s", red_servidor.PuertoServidor)
		estadoLbl.Color = tema.ColGreen
		estadoLbl.Refresh()
	}

	// ── Layout ────────────────────────────────────────────────────────────
	top := container.NewVBox(barTitulo, tema.HRule(), panelIP, tema.HRule())
	w.SetContent(container.NewBorder(top, nil, nil, nil, container.NewPadded(scroll)))
	w.Show()
}
