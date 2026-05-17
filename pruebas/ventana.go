package main

// ventana.go
// Ventana principal del terminal: historial scrollable, barra de comandos
// y panel de métricas en tiempo real. Se muestra tras un login exitoso.

import (
	"fmt"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// mostrarVentanaPrincipal construye y muestra la ventana del terminal.
// Recibe la instancia de la app para poder llamar a a.Quit() si el usuario
// escribe "bye" o "exit".
func mostrarVentanaPrincipal(a fyne.App) {
	w := a.NewWindow("ShellOS")
	w.Resize(fyne.NewSize(980, 680))
	w.CenterOnScreen()

	// ── Canal de parada: se cierra cuando el usuario cierra la ventana ────
	stopReporte := make(chan struct{})
	w.SetOnClosed(func() { close(stopReporte) })

	// ── Barra de título ───────────────────────────────────────────────────
	titBar := canvas.NewText("▸ ShellOS  —  terminal interactiva", colAccent)
	titBar.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	titBar.TextSize = 13

	// ── Panel de métricas ─────────────────────────────────────────────────
	tituloMetricas := canvas.NewText("◈  MONITOR DEL SISTEMA", colAccent)
	tituloMetricas.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	tituloMetricas.TextSize = 11

	lblCPU := canvas.NewText("CPU  [ cargando...               ]   --.--%", colGreen)
	lblCPU.TextStyle = fyne.TextStyle{Monospace: true}
	lblCPU.TextSize = 12

	lblRAM := canvas.NewText("RAM  [ cargando...               ]   --.--%   (-- MB / -- MB)", colGreen)
	lblRAM.TextStyle = fyne.TextStyle{Monospace: true}
	lblRAM.TextSize = 12

	lblRed := canvas.NewText("NET  ↓ --           ↑ --", colGreen)
	lblRed.TextStyle = fyne.TextStyle{Monospace: true}
	lblRed.TextSize = 12

	lblActualizado := canvas.NewText("actualización cada 5 s", colMuted)
	lblActualizado.TextStyle = fyne.TextStyle{Monospace: true}
	lblActualizado.TextSize = 10

	panelMetricas := container.NewVBox(
		container.NewPadded(tituloMetricas),
		container.NewPadded(lblCPU),
		container.NewPadded(lblRAM),
		container.NewPadded(lblRed),
		container.NewPadded(lblActualizado),
	)

	// Iniciar goroutine de métricas (definido en reporte.go)
	reporte(lblCPU, lblRAM, lblRed, stopReporte)

	// ── Historial del terminal ────────────────────────────────────────────
	historial := widget.NewLabel("")
	historial.TextStyle = fyne.TextStyle{Monospace: true}
	historial.Wrapping = fyne.TextWrapWord

	scroll := container.NewVScroll(historial)

	var lineas []string
	agregar := func(linea string) {
		lineas = append(lineas, linea)
		historial.SetText(strings.Join(lineas, "\n"))
		scroll.ScrollToBottom()
	}

	// sistemaOperativo() está definido en shell.go y detecta el OS actual
	soLinea := fmt.Sprintf("  ║      %-38s║", sistemaOperativo()+" — shell interactiva")
	agregar("  ╔═══════════════════════════════════════════════╗")
	agregar("  ║           S H E L L O S   v1.0               ║")
	agregar(soLinea)
	agregar("  ║   'bye' / 'exit'  para cerrar la sesión       ║")
	agregar("  ╚═══════════════════════════════════════════════╝")
	agregar("")

	// ── Barra de entrada de comandos ──────────────────────────────────────
	promptLabel := labelMuted("")
	actualizarPrompt := func() {
		pwd, _ := os.Getwd()
		promptLabel.SetText(pwd + "  »")
	}
	actualizarPrompt()

	entrada := widget.NewEntry()
	entrada.SetPlaceHolder("comando…")
	entrada.TextStyle = fyne.TextStyle{Monospace: true}

	procesar := func() {
		texto := strings.TrimSpace(entrada.Text)
		entrada.SetText("")
		if texto == "" {
			return
		}
		pwd, _ := os.Getwd()
		agregar(fmt.Sprintf("  %s  » %s", pwd, texto))

		// ejecutarComando está definido en shell.go
		salida, quit := ejecutarComando(texto)
		if salida != "" {
			for _, l := range strings.Split(salida, "\n") {
				agregar("  " + l)
			}
		}
		agregar("")
		actualizarPrompt()
		if quit {
			a.Quit()
		}
	}

	entrada.OnSubmitted = func(_ string) { procesar() }

	btnEjec := widget.NewButton("↵ Ejecutar", func() { procesar() })
	btnEjec.Importance = widget.HighImportance

	barCmd := container.NewBorder(nil, nil, promptLabel, btnEjec, entrada)

	// ── Layout final ──────────────────────────────────────────────────────
	top := container.NewVBox(
		container.NewPadded(titBar),
		hRule(),
		panelMetricas,
		hRule(),
	)
	bottom := container.NewVBox(
		hRule(),
		container.NewPadded(barCmd),
	)

	w.SetContent(container.NewBorder(
		top,
		bottom,
		nil, nil,
		container.NewPadded(scroll),
	))
	w.Canvas().Focus(entrada)
	w.Show()
}