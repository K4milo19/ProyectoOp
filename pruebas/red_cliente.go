package main

// red_cliente.go
// Cliente TCP de ShellOS.
// Ventana dividida en dos zonas:
//   - Izquierda: terminal remota (comandos al servidor)
//   - Derecha:   panel de métricas locales del cliente (CPU, RAM, Disco)

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ── Conexión al servidor ──────────────────────────────────────────────────────

type ConexionServidor struct {
	conn    net.Conn
	scanner *bufio.Scanner
}

func conectarServidor(host string) (*ConexionServidor, error) {
	addr := host + puertoServidor
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("no se pudo conectar a %s: %w", addr, err)
	}
	return &ConexionServidor{
		conn:    conn,
		scanner: bufio.NewScanner(conn),
	}, nil
}

// enviarComando manda un comando al servidor y recoge (salida, pwd, error).
// El servidor responde con <<<PWD:/ruta>>>, luego la salida, luego <<<END>>>.
func (c *ConexionServidor) enviarComando(cmd string) (salida, pwd string, err error) {
	_, err = fmt.Fprintf(c.conn, "%s\n", cmd)
	if err != nil {
		return "", "", fmt.Errorf("error al enviar: %w", err)
	}

	var sb strings.Builder
	for c.scanner.Scan() {
		linea := c.scanner.Text()
		if linea == "<<<END>>>" {
			break
		}
		if strings.HasPrefix(linea, "<<<PWD:") && strings.HasSuffix(linea, ">>>") {
			pwd = strings.TrimSuffix(strings.TrimPrefix(linea, "<<<PWD:"), ">>>")
			continue
		}
		sb.WriteString(linea)
		sb.WriteByte('\n')
	}
	if err = c.scanner.Err(); err != nil {
		return sb.String(), pwd, fmt.Errorf("error al leer respuesta: %w", err)
	}
	return strings.TrimRight(sb.String(), "\n"), pwd, nil
}

func (c *ConexionServidor) cerrar() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// ── Ventana de conexión ───────────────────────────────────────────────────────

func mostrarVentanaConexion(a fyne.App) {
	w := a.NewWindow("ShellOS — Conectar al servidor")
	w.Resize(fyne.NewSize(420, 320))
	w.CenterOnScreen()

	titulo := canvas.NewText("MODO CLIENTE", colPrimary)
	titulo.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	titulo.TextSize = 22
	titulo.Alignment = fyne.TextAlignCenter

	sub := canvas.NewText("conexión remota vía TCP", colMuted)
	sub.TextStyle = fyne.TextStyle{Monospace: true}
	sub.TextSize = 11
	sub.Alignment = fyne.TextAlignCenter

	lblIP := canvas.NewText("IP DEL SERVIDOR", colMuted)
	lblIP.TextStyle = fyne.TextStyle{Monospace: true}
	lblIP.TextSize = 10

	inpIP := widget.NewEntry()
	inpIP.SetPlaceHolder("192.168.1.x  o  localhost")
	inpIP.TextStyle = fyne.TextStyle{Monospace: true}
	inpIP.Text = "localhost"

	msgErr := canvas.NewText("", colError)
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
		msgErr.Color = colMuted
		msgErr.Refresh()

		cs, err := conectarServidor(host)
		if err != nil {
			msgErr.Text = "✗  " + err.Error()
			msgErr.Color = colError
			msgErr.Refresh()
			return
		}

		w.Hide()
		mostrarTerminalCliente(a, cs, host)
	}

	btnConectar.OnTapped = accion
	inpIP.OnSubmitted = func(_ string) { accion() }

	cuerpo := container.NewVBox(
		spacer(),
		container.NewCenter(titulo),
		container.NewCenter(sub),
		spacer(),
		hRule(),
		spacer(),
		lblIP, inpIP,
		spacer(),
		container.NewCenter(btnConectar),
		spacer(),
		container.NewCenter(msgErr),
	)
	w.SetContent(container.NewPadded(container.NewPadded(cuerpo)))
	w.Canvas().Focus(inpIP)
	w.Show()
}

// ── Terminal remota ───────────────────────────────────────────────────────────

func mostrarTerminalCliente(a fyne.App, cs *ConexionServidor, host string) {
	w := a.NewWindow(fmt.Sprintf("ShellOS  —  cliente  →  %s", host))
	w.Resize(fyne.NewSize(1280, 700))
	w.CenterOnScreen()

	// Canal para detener el reporte al cerrar
	stopReporte := make(chan struct{})
	w.SetOnClosed(func() {
		cs.cerrar()
		close(stopReporte)
	})

	// ── Barra de título ───────────────────────────────────────────────────
	titBar := canvas.NewText(
		fmt.Sprintf("▸ ShellOS  —  terminal remota  [ servidor: %s%s ]", host, puertoServidor),
		colAccent,
	)
	titBar.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	titBar.TextSize = 12

	estadoLbl := canvas.NewText("● CONECTADO", colGreen)
	estadoLbl.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	estadoLbl.TextSize = 11

	barTitulo := container.NewBorder(nil, nil, nil,
		container.NewPadded(estadoLbl),
		container.NewPadded(titBar),
	)

	// ── Panel de métricas locales (derecha) ───────────────────────────────
	tituloPanel := canvas.NewText("◈  ESTE EQUIPO", colAccent)
	tituloPanel.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	tituloPanel.TextSize = 11

	lblCPU := widget.NewLabel("CPU\n[ cargando... ]")
	lblCPU.TextStyle = fyne.TextStyle{Monospace: true}
	lblCPU.Wrapping = fyne.TextWrapWord

	lblRAM := widget.NewLabel("RAM\nusada: -- MB\ndisponible: -- MB")
	lblRAM.TextStyle = fyne.TextStyle{Monospace: true}
	lblRAM.Wrapping = fyne.TextWrapWord

	lblDisco := widget.NewLabel("DISCO\nusada: -- GB\ntotal: -- GB\nlibre: -- GB")
	lblDisco.TextStyle = fyne.TextStyle{Monospace: true}
	lblDisco.Wrapping = fyne.TextWrapWord

	lblTick := canvas.NewText("cada 5 s", colMuted)
	lblTick.TextStyle = fyne.TextStyle{Monospace: true}
	lblTick.TextSize = 10

	panelDerecho := container.NewVBox(
		container.NewPadded(tituloPanel),
		hRule(),
		container.NewPadded(lblCPU),
		hRule(),
		container.NewPadded(lblRAM),
		hRule(),
		container.NewPadded(lblDisco),
		hRule(),
		container.NewPadded(lblTick),
	)

	// Iniciar goroutine de métricas locales
	go reporte(lblCPU, lblRAM, lblDisco, stopReporte)

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
	agregar(fmt.Sprintf("  ║  servidor: %-36s║", host+puertoServidor))
	agregar("  ║  los comandos se ejecutan en el servidor      ║")
	agregar("  ╚═══════════════════════════════════════════════╝")
	agregar("")

	// ── Entrada de comandos ───────────────────────────────────────────────
	promptLbl := labelMuted("~  »")

	if _, pwd, err := cs.enviarComando("pwd"); err == nil && pwd != "" {
		promptLbl.SetText(pwd + "  »")
	}

	entrada := widget.NewEntry()
	entrada.SetPlaceHolder("comando remoto…")
	entrada.TextStyle = fyne.TextStyle{Monospace: true}

	procesar := func() {
		texto := strings.TrimSpace(entrada.Text)
		entrada.SetText("")
		if texto == "" {
			return
		}

		agregar(fmt.Sprintf("  %s  » %s", promptLbl.Text, texto))

		salida, pwd, err := cs.enviarComando(texto)
		if err != nil {
			agregar("  ✗ error de red: " + err.Error())
			estadoLbl.Text = "● DESCONECTADO"
			estadoLbl.Color = colError
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
			estadoLbl.Color = colError
			estadoLbl.Refresh()
			entrada.Disable()
		}
	}

	entrada.OnSubmitted = func(_ string) { procesar() }

	btnEjec := widget.NewButton("↵ Ejecutar", func() { procesar() })
	btnEjec.Importance = widget.HighImportance

	barCmd := container.NewBorder(nil, nil, promptLbl, btnEjec, entrada)

	// ── Layout: terminal izquierda | panel métricas derecha ───────────────
	terminal := container.NewBorder(
		nil,
		container.NewVBox(hRule(), container.NewPadded(barCmd)),
		nil, nil,
		container.NewPadded(scroll),
	)

	contenido := container.NewBorder(
		container.NewVBox(barTitulo, hRule()),
		nil, nil,
		container.NewPadded(panelDerecho),  // panel fijo a la derecha
		terminal,
	)

	w.SetContent(contenido)
	w.Canvas().Focus(entrada)
	w.Show()
}