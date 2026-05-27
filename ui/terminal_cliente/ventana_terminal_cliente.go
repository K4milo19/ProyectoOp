package terminal_cliente

// ventana_terminal_cliente.go
// Terminal remota del cliente: historial de comandos a la izquierda,
// panel de métricas del servidor a la derecha.

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"shellos/red_cliente"
	"shellos/tema"
)

// MostrarTerminalCliente abre la ventana de terminal remota.
func MostrarTerminalCliente(a fyne.App, cs *red_cliente.ConexionServidor, host string) {
	w := a.NewWindow(fmt.Sprintf("ShellOS  —  cliente  →  %s", host))
	w.Resize(fyne.NewSize(1280, 700))
	w.CenterOnScreen()

	stopMetrics := make(chan struct{})
	w.SetOnClosed(func() {
		cs.Cerrar()
		close(stopMetrics)
	})

	// ── Barra de título ───────────────────────────────────────────────────
	titBar := canvas.NewText(
		fmt.Sprintf("▸ ShellOS  —  terminal remota  [ servidor: %s%s ]", host, red_cliente.PuertoServidor),
		tema.ColAccent,
	)
	titBar.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	titBar.TextSize = 12

	estadoLbl := canvas.NewText("● CONECTADO", tema.ColGreen)
	estadoLbl.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	estadoLbl.TextSize = 11

	barTitulo := container.NewBorder(nil, nil, nil,
		container.NewPadded(estadoLbl),
		container.NewPadded(titBar),
	)

	// ── Panel de métricas remotas (derecha) ───────────────────────────────
	tituloPanel := canvas.NewText("◈  MÉTRICAS DEL SERVIDOR", tema.ColAccent)
	tituloPanel.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	tituloPanel.TextSize = 11

	lblCPU := widget.NewLabel("CPU\n[ esperando datos... ]")
	lblCPU.TextStyle = fyne.TextStyle{Monospace: true}
	lblCPU.Wrapping = fyne.TextWrapWord

	lblRAM := widget.NewLabel("RAM\nusada: -- MB  |  libre: -- MB")
	lblRAM.TextStyle = fyne.TextStyle{Monospace: true}
	lblRAM.Wrapping = fyne.TextWrapWord

	lblDisco := widget.NewLabel("DISCO\nusado: -- GB  |  total: -- GB  |  libre: -- GB")
	lblDisco.TextStyle = fyne.TextStyle{Monospace: true}
	lblDisco.Wrapping = fyne.TextWrapWord

	lblTick := canvas.NewText("push cada 5 s", tema.ColMuted)
	lblTick.TextStyle = fyne.TextStyle{Monospace: true}
	lblTick.TextSize = 10

	panelDerecho := container.NewVBox(
		container.NewPadded(tituloPanel),
		tema.HRule(),
		container.NewPadded(lblCPU),
		tema.HRule(),
		container.NewPadded(lblRAM),
		tema.HRule(),
		container.NewPadded(lblDisco),
		tema.HRule(),
		container.NewPadded(lblTick),
	)

	// Callback que actualiza el panel derecho desde cualquier goroutine
	cs.OnMetrics = func(cpu, ramUsed, ramFree, diskUsed, diskTotal float64) {
		fyne.Do(func() {
			lblCPU.SetText(fmt.Sprintf("CPU: %.1f%%", cpu))
			lblRAM.SetText(fmt.Sprintf("RAM — usada: %.0f MB  |  libre: %.0f MB", ramUsed, ramFree))
			lblDisco.SetText(fmt.Sprintf(
				"DISCO — usado: %.1f GB  |  total: %.1f GB  |  libre: %.1f GB",
				diskUsed, diskTotal, diskTotal-diskUsed,
			))
		})
	}
	cs.DespachadorMetricas(stopMetrics)

	// ── Historial ─────────────────────────────────────────────────────────
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

	agregar("  ╔═══════════════════════════════════════════════╗")
	agregar("  ║        S H E L L O S  —  C L I E N T E       ║")
	agregar(fmt.Sprintf("  ║  servidor: %-36s║", host+red_cliente.PuertoServidor))
	agregar("  ║  los comandos se ejecutan en el servidor      ║")
	agregar("  ║  métricas del servidor → panel derecho        ║")
	agregar("  ╚═══════════════════════════════════════════════╝")
	agregar("")

	// ── Entrada de comandos ───────────────────────────────────────────────
	promptLbl := tema.LabelMuted("~  »")
	go func() {
		if _, pwd, err := cs.EnviarComando("pwd"); err == nil && pwd != "" {
			fyne.Do(func() { promptLbl.SetText(pwd + "  »") })
		}
	}()

	entrada := widget.NewEntry()
	entrada.SetPlaceHolder("comando remoto…")
	entrada.TextStyle = fyne.TextStyle{Monospace: true}

	var cmdEnCurso bool

	procesar := func() {
		if cmdEnCurso {
			return
		}
		texto := strings.TrimSpace(entrada.Text)
		entrada.SetText("")
		if texto == "" {
			return
		}
		cmdEnCurso = true
		agregar(fmt.Sprintf("  %s  » %s", promptLbl.Text, texto))

		go func() {
			salida, pwd, err := cs.EnviarComando(texto)
			fyne.Do(func() {
				cmdEnCurso = false
				if err != nil {
					agregar("  ✗ error de red: " + err.Error())
					estadoLbl.Text = "● DESCONECTADO"
					estadoLbl.Color = tema.ColError
					estadoLbl.Refresh()
					entrada.Disable()
					agregar("")
					return
				}
				if pwd != "" {
					promptLbl.SetText(pwd + "  »")
				}
				for _, l := range strings.Split(salida, "\n") {
					agregar("  " + l)
				}
				agregar("")
				if texto == "bye" || texto == "exit" {
					estadoLbl.Text = "● DESCONECTADO"
					estadoLbl.Color = tema.ColError
					estadoLbl.Refresh()
					entrada.Disable()
				}
			})
		}()
	}

	entrada.OnSubmitted = func(_ string) { procesar() }
	btnEjec := widget.NewButton("↵ Ejecutar", func() { procesar() })
	btnEjec.Importance = widget.HighImportance

	barCmd := container.NewBorder(nil, nil, promptLbl, btnEjec, entrada)

	terminal := container.NewBorder(
		nil,
		container.NewVBox(tema.HRule(), container.NewPadded(barCmd)),
		nil, nil,
		container.NewPadded(scroll),
	)

	contenido := container.NewBorder(
		container.NewVBox(barTitulo, tema.HRule()),
		nil, nil,
		container.NewPadded(panelDerecho),
		terminal,
	)

	w.SetContent(contenido)
	w.Canvas().Focus(entrada)
	w.Show()
}
