package main

// ventana.go
// Ventana principal del terminal LOCAL (modo shell directa).
// Solo se usa cuando el nodo opera de forma independiente (no servidor/cliente).
// Los comandos se ejecutan localmente con ejecutarComando() de shell.go.

import (
	"fmt"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// mostrarVentanaPrincipal construye y muestra la ventana del terminal local.
// Se llama tras un login exitoso en modo standalone (sin red).
func mostrarVentanaPrincipal(a fyne.App) {
	w := a.NewWindow("ShellOS  —  terminal local")
	w.Resize(fyne.NewSize(980, 680))
	w.CenterOnScreen()

	// ── Canal de parada del monitor de métricas ───────────────────────────
	stopReporte := make(chan struct{})
	w.SetOnClosed(func() { close(stopReporte) })

	// ── Barra de título ───────────────────────────────────────────────────
	titBar := canvas.NewText("▸ ShellOS  —  terminal local", colAccent)
	titBar.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	titBar.TextSize = 13

	// ── Panel de métricas (reporte.go) ────────────────────────────────────
	tituloMetricas := canvas.NewText("◈  MONITOR DEL SISTEMA  —  actualiza cada 5 s", colAccent)
	tituloMetricas.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	tituloMetricas.TextSize = 11

	lblCPU := canvas.NewText("CPU    [ cargando...                    ]  --.-%", colGreen)
	lblCPU.TextStyle = fyne.TextStyle{Monospace: true}
	lblCPU.TextSize = 12

	lblRAM := canvas.NewText("RAM    usada:   ---- MB   disponible:   ---- MB", colGreen)
	lblRAM.TextStyle = fyne.TextStyle{Monospace: true}
	lblRAM.TextSize = 12

	lblDisco := canvas.NewText("DISCO  usada:  --.-- GB   total:       --.-- GB   libre: --.-- GB", colGreen)
	lblDisco.TextStyle = fyne.TextStyle{Monospace: true}
	lblDisco.TextSize = 12

	panelMetricas := container.NewVBox(
		container.NewPadded(tituloMetricas),
		container.NewPadded(lblCPU),
		container.NewPadded(lblRAM),
		container.NewPadded(lblDisco),
	)

	reporte(lblCPU, lblRAM, lblDisco, stopReporte)

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

	soLinea := fmt.Sprintf("  ║      %-38s║", sistemaOperativo()+" — shell local")
	agregar("  ╔═══════════════════════════════════════════════╗")
	agregar("  ║           S H E L L O S   v1.0               ║")
	agregar(soLinea)
	agregar("  ║   'bye' / 'exit'  para cerrar la sesión       ║")
	agregar("  ╚═══════════════════════════════════════════════╝")
	agregar("")

	// ── Entrada de comandos ───────────────────────────────────────────────
	promptLbl := labelMuted("")
	actualizarPrompt := func() {
		pwd, _ := os.Getwd()
		promptLbl.SetText(pwd + "  »")
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

	barCmd := container.NewBorder(nil, nil, promptLbl, btnEjec, entrada)

	// ── Layout ────────────────────────────────────────────────────────────
	top := container.NewVBox(
		container.NewPadded(titBar),
		hRule(),
		panelMetricas,
		hRule(),
	)
	bottom := container.NewVBox(hRule(), container.NewPadded(barCmd))

	w.SetContent(container.NewBorder(
		top, bottom, nil, nil,
		container.NewPadded(scroll),
	))
	w.Canvas().Focus(entrada)
	w.Show()
}